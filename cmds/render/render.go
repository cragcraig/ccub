package render

import (
	"errors"
	"flag"
	"github.com/cragcraig/ccub/buildlog"
	"github.com/cragcraig/ccub/cli"
	"io"
	"os"
	"time"
)

var CmdFactory = cli.NewStaticCommandFactory(
	cli.CommandMetadata{
		Description: "Render build logs using a user-specified template",
	},
	func(name string) cli.Command { return &renderCmd{name: name} })

type renderCmd struct {
	name     string
	tmplFile string
}

func (c *renderCmd) Parse(args []string) error {
	flags := flag.NewFlagSet(c.name, flag.ContinueOnError)
	// Raw flags
	tmplFile := flags.String("tmpl", "", "Template text file. Required.")
	// Parse
	if err := flags.Parse(args); err != nil {
		return err
	}
	// Template
	if len(*tmplFile) == 0 {
		return errors.New("'template' is required")
	}
	c.tmplFile = *tmplFile
	return nil
}

func (c *renderCmd) Execute() error {
	tmpl, err := buildlog.LoadTemplateFromFile(c.tmplFile)
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
