package log

import (
	"fmt"
    "flag"
	"github.com/cragcraig/ccub/buildlog"
	"github.com/cragcraig/ccub/cli"
	"github.com/cragcraig/ccub/protos"
	"io"
	"os"
	"strings"
	"time"
    "errors"
)

const (
	logDetailsTemplate = ""
)

func containsString(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func NewFactory() cli.CommandFactory {
    return &logCmdFactory{}
}

type logCmdFactory struct {
}

func (f *logCmdFactory) Metadata() cli.CommandMetadata {
    return cli.CommandMetadata{
        Name: "log",
        Description: "Log a build entry",
    }
}

func (f *logCmdFactory) Create(name string) cli.Command {
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
    if assembly, err := buildlog.ParseAssemblyArg(*raw.assembly); err != nil {
        return err
    } else {
        c.assembly = assembly
    }
    // Date
    // TODO: Required, print usage if missing
    if date, err := buildlog.ParseDateArg(*raw.date); err != nil {
        return err
    } else {
        c.date = date
    }
    // Work periods
    // TODO: Required, print usage if missing
    if w, err := buildlog.ParseWorkPeriodsArg(c.date.Year(), c.date.Month(), c.date.Day(), *raw.workPeriods); err != nil {
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
            return errors.New("Tags must not be empty strings")
        }
    }
    return nil
}

func (c *logCmd) Execute() error {
	entry := protos.BuildLogEntry{
		Assembly:    c.assembly,
		Date:        buildlog.FormatDateForLog(c.date),
		WorkPeriod:  c.workPeriods,
		DetailsFile: buildlog.LogDetailsFile([]string{}, c.date),
		Tags:        c.tags,
	}

	if err := buildlog.UpdateLogMetadataFile(buildlog.LogsPath, &entry, *c.rawFlags.overwrite); err != nil {
		return err
	}
	fmt.Printf("Logged:   %s\n", buildlog.LogsPath)
	if f, err := buildlog.CreateLogDetailsFile(c.assembly, c.date, false); err != nil {
        return err
	} else {
        fmt.Printf("Details:  %s\n", f)
        return buildlog.LaunchEditor(f)
    }
}

func RenderCmd(argv []string) error {
	if len(argv) != 1 {
		return errors.New("invalid args")
	}
	tmpl, err := buildlog.LoadTemplateFromFile(argv[0])
	if err != nil {
		return err
	}
	logs, err := buildlog.ReadLogs(buildlog.LogsPath)

	// Render each log
	for _, log := range logs.LogEntry {
		// Render log entry using template
		err := tmpl.Execute(os.Stdout, log)
		if err != nil {
			return err
		}
		date, err := time.Parse(buildlog.DateLayout, log.Date)
		if err != nil {
			return err
		}
		// Append associated details Markdown file
		details, err := buildlog.ReadFile(buildlog.LogDetailsFile([]string{buildlog.LogsDir}, date))
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
