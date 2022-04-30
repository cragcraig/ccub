package cli

import (
	"fmt"
    "flag"
	"github.com/cragcraig/ccub/protos"
	"github.com/golang/protobuf/proto"
	"io"
	"os"
	"sort"
	"strings"
	"time"
    "errors"
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

func parseAssemblyArg(arg string) (string, error) {
	if !containsString(validAssemblies, arg) {
		return "", fmt.Errorf("Assembly must be one of:\n  %s", strings.Join(validAssemblies, "\n  "))
	}
    return arg, nil
}

type logCmdFactory struct {
}

func (f *logCmdFactory) Metadata() CommandMetadata {
    return CommandMetadata{
        Name: "log",
        Description: "Log a build entry",
    }
}

func (f *logCmdFactory) Create(name string) Command {
    flags := flag.NewFlagSet(name, flag.ContinueOnError)
    rawFlags := logCmdRawFlags{
        flags: flags,
        assembly: flags.String("assembly", "", "Top-level assembly. Required."),
        date: flags.String("date", "", "Date of work. Required."),
        workPeriods: flags.String("time", "", "Time period(s) of work. Required."),
        tags: flags.String("tags", "", "Comma-separated list of arbitrary tags"),
        overwrite: flags.Bool("overwrite", false, "Replace existing log entry on specified date"),
    }
    return &logCmd{
        rawFlags: &rawFlags,
    }
}

type logCmdRawFlags struct {
    flags *flag.FlagSet
    assembly *string
    date *string
    workPeriods *string
    tags *string
    overwrite *bool
}

type logCmd struct {
    rawFlags *logCmdRawFlags
    assembly string
    date time.Time
    workPeriods []*protos.TimePeriod
    tags []string
}

func (c *logCmd) Parse(args []string) error {
    if err := c.rawFlags.flags.Parse(args); err != nil {
        return err
    }
    raw := c.rawFlags
    // Assembly
    if !containsString(validAssemblies, *raw.assembly) {
        return fmt.Errorf(
            "assembly must be one of:\n  %s\n", strings.Join(validAssemblies, "\n  "))
    }
    c.assembly = *raw.assembly
    // Date
    // TODO: Required, print usage if missing
    if date, err := parseDateArg(*raw.date); err != nil {
        return err
    } else {
        c.date = date
    }
    // Work periods
    // TODO: Required, print usage if missing
    if w, err := parseWorkPeriodsArg(c.date.Year(), c.date.Month(), c.date.Day(), *raw.workPeriods); err != nil {
        return err
    } else {
        c.workPeriods = w
    }
    // Tags
    if len(*raw.tags) > 0 {
        c.tags = strings.Split(*raw.tags, ",")
    }
    for _, t := range c.tags {
        if len(t) == 0 {
            return errors.New("Tags must not be an empty string")
        }
    }
    return nil
}

func (c *logCmd) Execute() error {
	entry := protos.BuildLogEntry{
		Assembly:    c.assembly,
		Date:        formatDateForLog(c.date),
		WorkPeriod:  c.workPeriods,
		DetailsFile: logDetailsFile([]string{}, c.date),
		Tags:        c.tags,
	}

	if err := updateLogMetadataFile(logsPath, &entry, *c.rawFlags.overwrite); err != nil {
		return err
	}
	fmt.Printf("Logged:   %s\n", logsPath)
	if f, err := createLogDetailsFile(c.assembly, c.date, false); err != nil {
        return err
	} else {
        fmt.Printf("Details:  %s\n", f)
        return launchEditor(f)
    }
}

func RenderCmd(argv []string) error {
	if len(argv) != 1 {
		return errors.New("invalid args")
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
			if os.IsNotExist(err) {
				details = "No details"
			} else {
				return err
			}
		}
		if _, err := io.WriteString(os.Stdout, details+"\n"); err != nil {
			return err
		}
	}

	return nil
}
