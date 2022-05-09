package log

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/cragcraig/ccub/buildlog"
	"github.com/cragcraig/ccub/cli"
	"github.com/cragcraig/ccub/protos"
)

const (
	logDetailsTemplate = ""
)

var CmdFactory = cli.NewStaticCommandFactory(
	cli.CommandMetadata{
		Description: "Log a build entry",
	},
	func(name string) cli.Command { return &logCmd{name: name} })

type logCmd struct {
	name        string
	assembly    string
	date        time.Time
	workPeriods []*protos.TimePeriod
	title       string
	tags        []string
	overwrite   bool
}

func (c *logCmd) Parse(args []string) error {
	flags := flag.NewFlagSet(c.name, flag.ContinueOnError)
	// Raw flags
	assembly := flags.String("assembly", "", "Top-level assembly. Required.")
	date := flags.String("date", "", "Date of work. Required.")
	workPeriods := flags.String("time", "", "Time period(s) of work. Required.")
	title := flags.String("title", "", "Title for the log entry")
	tags := flags.String("tags", "", "Comma-separated list of arbitrary tags")
	overwrite := flags.Bool("overwrite", false, "Replace existing log entry on specified date")
	// Parse
	if err := flags.Parse(args); err != nil {
		return err
	}
	// Assembly
	if len(*assembly) == 0 {
		return errors.New("'assembly' is required")
	}
	if a, err := buildlog.ParseAssemblyArg(*assembly); err != nil {
		return err
	} else {
		c.assembly = a
	}
	// Date
	// TODO: Required, print usage if missing
	if len(*date) == 0 {
		return errors.New("'date' is required")
	}
	if d, err := buildlog.ParseDateArg(*date); err != nil {
		return err
	} else {
		c.date = d
	}
	// Work periods
	if len(*workPeriods) == 0 {
		return errors.New("'time' is required")
	}
	if w, err := buildlog.ParseWorkPeriodsArg(c.date.Year(), c.date.Month(), c.date.Day(), *workPeriods); err != nil {
		return err
	} else {
		c.workPeriods = w
	}
	if len(*title) == 0 {
		return errors.New("'title' is required")
	}
	for _, r := range *title {
		if !unicode.IsPrint(r) {
			return errors.New("Title must be a single line of text (no newlines)")
		}
	}
	c.title = *title
	// Tags
	if len(*tags) > 0 {
		c.tags = strings.Split(*tags, ",")
	}
	for _, t := range c.tags {
		if len(t) == 0 {
			return errors.New("Tags must not be empty strings")
		}
	}
	// Overwrite
	c.overwrite = *overwrite
	return nil
}

func (c *logCmd) Execute() error {
	entry := protos.BuildLogEntry{
		Assembly:    c.assembly,
		Date:        buildlog.FormatDateForLog(c.date),
		WorkPeriod:  c.workPeriods,
		Title:       c.title,
		DetailsFile: buildlog.LogDetailsFile([]string{}, c.date),
		Tags:        c.tags,
	}

	if err := buildlog.UpdateLogMetadataFile(buildlog.LogsPath, &entry, c.overwrite); err != nil {
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
