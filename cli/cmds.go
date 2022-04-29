package cli

import (
	"fmt"
	"github.com/cragcraig/ccub/protos"
	"github.com/golang/protobuf/proto"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	LogsDir            = "log"
	LogsFile           = "buildlog.textproto"
	logsPath           = LogsDir + "/" + LogsFile
	logDetailsTemplate = ""
)

var validAssemblies = []string{
	"left wing",
	"right wing",
	"fuselage",
	"skin",
	"avionics",
	"powerplant",
	"gear",
}

func containsString(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func logExists(date time.Time, logs *protos.BuildLogs) (exists bool, index int) {
	if logs.LogEntry == nil {
		return false, -1
	}
	d := formatDateForLog(date)
	for i, v := range logs.LogEntry {
		if v.Date == d {
			return true, i
		}
	}
	return false, -1
}

func readLogs(f string) (*protos.BuildLogs, error) {
	text, err := readFile(f)
	if err != nil {
		return &protos.BuildLogs{}, err
	}
	entries := protos.BuildLogs{}
	err = proto.UnmarshalText(text, &entries)
	return &entries, err
}

func writeLogs(f string, logs *protos.BuildLogs) error {
	err := ensureDirExists(LogsDir)
	if err != nil {
		return err
	}
	fp, err := os.Create(f)
	if err != nil {
		return err
	}
	defer fp.Close()
	return proto.MarshalText(fp, logs)
}

func logDetailsDir(basedir []string, date time.Time) string {
	return strings.Join(append(basedir, date.Format(MonthLayout)), "/")
}

func logDetailsFile(basedir []string, date time.Time) string {
	return strings.Join([]string{logDetailsDir(basedir, date), date.Format(DateLayout) + ".md"}, "/")
}

func createLogDetailsFile(assembly string, date time.Time, overwrite bool) (string, error) {
	if err := ensureDirExists(logDetailsDir([]string{LogsDir}, date)); err != nil {
		return "", err
	}
	f := logDetailsFile([]string{LogsDir}, date)
	if !overwrite {
		if exists, err := fileExists(f); err != nil {
			return "", err
		} else if exists {
			return f, fmt.Errorf("Log details file %s already exists", f)
		}
	}
	if err := os.WriteFile(f, []byte(logDetailsTemplate), 0666); err != nil {
		return "", err
	}
	return f, nil
}

func updateLogMetadataFile(f string, entry *protos.BuildLogEntry, overwrite bool) error {
	logs, err := readLogs(f)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Could not open logs metadata from %s\n", f, err.Error())
	}

	date, err := parseDateOfLog(entry)
	if err != nil {
		return err
	}
	exists, index := logExists(date, logs)
	// Don't overwrite
	if !overwrite && exists {
		return fmt.Errorf("Log entry already exists for %s", entry.Date)
	}

	// Overwrite if exists, otherwise append
	if exists {
		logs.LogEntry[index] = entry
	} else {
		logs.LogEntry = append(logs.LogEntry, entry)
	}

	// Ensure logs are ordered by date
	sort.Slice(logs.LogEntry, func(i, j int) bool {
		it, _ := parseDateOfLog(logs.LogEntry[i])
		jt, _ := parseDateOfLog(logs.LogEntry[j])
		return jt.After(it)
	})

	return writeLogs(f, logs)
}

func LogCmd(cmd CommandEntry, argv []string) error {
	if len(argv) < 3 {
		return cmd.getUsageError()
	}

	// Assembly
	assembly := argv[0]
	if !containsString(validAssemblies, assembly) {
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

	entry := protos.BuildLogEntry{
		Assembly:    assembly,
		Date:        formatDateForLog(date),
		WorkPeriod:  times,
		DetailsFile: logDetailsFile([]string{}, date),
		Tags:        tags,
	}

	if err := updateLogMetadataFile(logsPath, &entry, false); err != nil {
		return err
	}
	fmt.Printf("Logged:   %s\n", logsPath)
	if f, err := createLogDetailsFile(assembly, date, false); err != nil {
		return err
	} else {
		fmt.Printf("Details:  %s\n", f)
		return launchEditor(f)
	}
}

func RenderCmd(cmd CommandEntry, argv []string) error {
	if len(argv) != 1 {
		return cmd.getUsageError()
	}
	tmpl, err := loadTemplateFromFile(argv[0])
	if err != nil {
		return err
	}
	logs, err := readLogs(logsPath)

	// Render each log
	for _, log := range logs.LogEntry {
		// Render log entry using template
		err := tmpl.Execute(os.Stdout, log)
		if err != nil {
			return err
		}
		date, err := time.Parse(DateLayout, log.Date)
		if err != nil {
			return err
		}
		// Append associated details Markdown file
		details, err := readFile(logDetailsFile([]string{LogsDir}, date))
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			details = "No details"
		}
		if _, err := io.WriteString(os.Stdout, details+"\n"); err != nil {
			return err
		}
	}

	return nil
}

/*
func ExampleCmd(cmd CommandEntry, argv []string) error {
	if len(argv) != 1 {
		return cmd.getUsageError()
	}
	fmt.Println("TODO: Implement")
	return nil
}
*/
