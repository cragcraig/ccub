package cli

import (
	"fmt"
	"github.com/cragcraig/ccub/protos"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const MonthLayout = "2006-Jan"
const DateLayout = "2006-Jan-02"
const LogsDir = "log"
const LogsFile = "buildlog.textproto"

var kitchenTimePattern = regexp.MustCompile(`^(\d+)(:\d\d)?(AM|PM|am|pm)$`)
var validAssemblies = []string{
	"left wing",
	"right wing",
	"fuselage",
	"skin",
	"avionics",
	"powerplant",
	"gear",
}

const logDetailsTemplate = "## title - %s \n\ntext"

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func logExists(date string, logs *protos.BuildLogs) bool {
	if logs.LogEntry == nil {
		return false
	}
	for _, v := range logs.LogEntry {
		if v.Date == date {
			return true
		}
	}
	return false
}

func readLogs(filename string) (*protos.BuildLogs, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return &protos.BuildLogs{}, err
	}
	defer fp.Close()
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		return &protos.BuildLogs{}, err
	}
	text := string(data)
	entries := protos.BuildLogs{}
	err = proto.UnmarshalText(text, &entries)
	return &entries, err
}

func ensureDirExists(dir string) error {
	fp, err := os.Open(dir)
	if os.IsNotExist(err) {
		return os.Mkdir(dir, 0755)
	} else if err != nil {
		return err
	}
	defer fp.Close()
	fi, err := fp.Stat()
	if err != nil {
		return err
	}
	if !fi.Mode().IsDir() {
		return fmt.Errorf("Logs directory %s already exists but is not a directory", dir)
	}
	return nil
}

func writeLogs(filename string, logs *protos.BuildLogs) error {
	err := ensureDirExists(LogsDir)
	if err != nil {
		return err
	}
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	return proto.MarshalText(fp, logs)
}

func parseDateArg(arg string) (time.Time, error) {
	if arg == "today" {
		return time.Now(), nil
	}
	return time.Parse(DateLayout, arg)
}

func parseKitchenTime(year int, month time.Month, day int, kitchen string) (time.Time, error) {
	submatches := kitchenTimePattern.FindStringSubmatch(kitchen)
	if submatches == nil {
		return time.Time{}, fmt.Errorf("Invalid time: %s", kitchen)
	}
	hours, _ := strconv.Atoi(submatches[1])
	minutes := 0
	if len(submatches[2]) > 0 {
		minutes, _ = strconv.Atoi(submatches[2][1:])
	}
	ampm := strings.ToUpper(submatches[3])
	if ampm == "PM" {
		hours += 12
	}
	return time.Date(year, month, day, hours, minutes, 0, 0, time.UTC), nil
}

func parseDurationsArg(year int, month time.Month, day int, arg string) ([]*protos.TimePeriod, error) {
	var periods []*protos.TimePeriod
	durations := strings.Split(arg, ",")
	for _, v := range durations {
		s := strings.Split(v, "-")
		if len(s) != 2 {
			return nil, fmt.Errorf("Time period must consist of both a start time and an end time")
		}
		start, err := parseKitchenTime(year, month, day, s[0])
		if err != nil {
			return nil, fmt.Errorf("Bad start time: %s", s[0])
		}
		end, err := parseKitchenTime(year, month, day, s[1])
		if err != nil {
			return nil, fmt.Errorf("Bad end time: %s", s[1])
		}
		if start.After(end) {
			return nil, fmt.Errorf("Start time %s is after end time %s", s[0], s[1])
		}
		periods = append(periods, &protos.TimePeriod{
			StartTime:   start.Format(time.Kitchen),
			DurationMin: uint32(math.Ceil(end.Sub(start).Minutes())),
		})
	}
	return periods, nil
}

func logDetailsDir(basedir []string, date time.Time) string {
	return strings.Join(append(basedir, date.Format(MonthLayout)), "/")
}

func logDetailsFile(basedir []string, date time.Time) string {
	return strings.Join([]string{logDetailsDir(basedir, date), date.Format(DateLayout) + ".md"}, "/")
}

func createLogDetailsFile(assembly string, date time.Time) (string, error) {
	if err := ensureDirExists(logDetailsDir([]string{LogsDir}, date)); err != nil {
		return "", err
	}
	file := logDetailsFile([]string{LogsDir}, date)
	if err := os.WriteFile(file, []byte(fmt.Sprintf(logDetailsTemplate, strings.Title(assembly))), 0666); err != nil {
		return "", err
	}
	return file, nil
}

func updateLogMetadataFile(file string, entry *protos.BuildLogEntryMetadata) error {
	logs, err := readLogs(file)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Could not open logs metadata from %s\n", file, err.Error())
	}

	if logExists(entry.Date, logs) {
		return fmt.Errorf("Log entry already exists for %s", entry.Date)
	}

	logs.LogEntry = append(logs.LogEntry, entry)

	sort.Slice(logs.LogEntry, func(i, j int) bool {
		it, _ := time.Parse(DateLayout, logs.LogEntry[i].Date)
		jt, _ := time.Parse(DateLayout, logs.LogEntry[j].Date)
		return jt.After(it)
	})

	return writeLogs(file, logs)
}

func NewLogCmd(cmd CommandEntry, argv []string) error {
	if len(argv) < 3 {
		return cmd.getUsageError()
	}

	// Assembly
	assembly := argv[0]
	if !contains(validAssemblies, assembly) {
		return fmt.Errorf("Assembly must be one of:\n  %s", strings.Join(validAssemblies, "\n  "))
	}

	// Date
	date, err := parseDateArg(argv[1])
	if err != nil {
		return err
	}

	// Time periods
	times, err := parseDurationsArg(date.Year(), date.Month(), date.Day(), argv[2])
	if err != nil {
		return err
	}

	// Tags
	var tags []string
	if len(argv) > 3 {
		tags = argv[3:]
	}

	entry := protos.BuildLogEntryMetadata{
		Assembly:    assembly,
		Date:        date.Format(DateLayout),
		WorkPeriod:  times,
		DetailsFile: logDetailsFile([]string{}, date),
		Tags:        tags,
	}

	logsfile := LogsDir + "/" + LogsFile
	if err := updateLogMetadataFile(logsfile, &entry); err != nil {
		return err
	}
	fmt.Printf("Logged:   %s\n", logsfile)
	//fmt.Printf("%s\n", entry.String())
	if file, err := createLogDetailsFile(assembly, date); err != nil {
		return err
	} else {
		fmt.Printf("Details:  %s\n", file)
	}
	return nil
}

func LogCmd(cmd CommandEntry, argv []string) error {
	if len(argv) != 1 {
		return cmd.getUsageError()
	}
	fmt.Println("TODO: Log")
	return nil
}

func EndCmd(cmd CommandEntry, argv []string) error {
	if len(argv) != 1 {
		return cmd.getUsageError()
	}
	fmt.Println("TODO: End")
	return nil
}
