package cli

import (
    "math"
    "strconv"
	"fmt"
    "io/ioutil"
    "os"
    "github.com/golang/protobuf/proto"
    "time"
    "strings"
    "regexp"
    "github.com/cragcraig/ccub/logprotos"
)

const MonthLayout = "2006-Jan"
const DateLayout = "2006-Jan-02"
var kitchenTimePattern = regexp.MustCompile(`^(\d+)(:\d\d)?(AM|PM|am|pm)$`)
const LogsDir = "logs"
const LogsFile = "metadata.textproto"

func readLogs(filename string) (*logprotos.BuildLogs, error) {
    fp, err := os.Open(filename)
    if err != nil {
        return &logprotos.BuildLogs{}, err
    }
    defer fp.Close()
    data, err := ioutil.ReadAll(fp)
    if err != nil {
        return &logprotos.BuildLogs{}, err
    }
    text := string(data)
    entries := logprotos.BuildLogs{}
    err = proto.UnmarshalText(text, &entries)
    return &entries, err
}

func ensureLogsDirExists() error {
    fp, err := os.Open(LogsDir)
    if os.IsNotExist(err) {
        return os.Mkdir(LogsDir, 0755)
    } else if err != nil {
        return err
    }
    defer fp.Close()
    fi, err := fp.Stat()
    if err != nil {
        return err
    }
    if !fi.Mode().IsDir() {
        return fmt.Errorf("Logs directory %s already exists but is not a directory", LogsDir)
    }
    return nil
}

func writeLogs(filename string, logs *logprotos.BuildLogs) error {
    err := ensureLogsDirExists()
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

func parseDurationsArg(year int, month time.Month, day int, arg string) ([]*logprotos.TimePeriod, error) {
    var periods []*logprotos.TimePeriod
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
        periods = append(periods, &logprotos.TimePeriod{
            StartTime: start.Format(time.Kitchen),
            DurationMin: uint32(math.Ceil(end.Sub(start).Minutes())),
        })
    }
    return periods, nil
}

func NewLogCmd(cmd CommandEntry, argv []string) error {
	if len(argv) < 3 {
		return cmd.getUsageError()
	}
    
    // Assembly
    assembly := argv[0]

    // Date
    date, err := parseDateArg(argv[1])
    if err != nil {
        return err
    }

    // Times
    times, err := parseDurationsArg(date.Year(), date.Month(), date.Day(), argv[2]);
    if err != nil {
        return err
    }

    // Tags
    var tags []string
    if len(argv) > 3 {
        tags = argv[3:]
    }

    logfile := date.Format(DateLayout) + ".md"
    entry := logprotos.BuildLogEntryMetadata{
        Assembly: assembly,
        Date: date.Format(DateLayout),
        WorkPeriod: times,
        LogFile: logfile,
        Tags: tags,
    }

    //fmt.Printf("%v\n", entry)

    fname := LogsDir + "/" + LogsFile
    logs := &logprotos.BuildLogs{}
    logs, err = readLogs(fname)
    if err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("Could not open logs metadata from %s\n", fname, err.Error())
    }
    logs.LogEntry = append(logs.LogEntry, &entry)
    return writeLogs(fname, logs)
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
