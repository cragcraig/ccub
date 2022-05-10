package render

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/cragcraig/ccub/buildlog"
	"github.com/cragcraig/ccub/cli"
	"github.com/cragcraig/ccub/protos"
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

type renderData struct {
	*protos.BuildLogEntry
	Details string
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
		return errors.New("'tmpl' is required")
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

	// Render each log entry
	for _, log := range logs.LogEntry {
		// Read in details file
		details, err := buildlog.ReadFile(strings.Join([]string{buildlog.LogsDir, log.DetailsFile}, "/"))
		if err != nil {
			if os.IsNotExist(err) {
				details = "No details"
			} else {
				return err
			}
		}
		// Render log entry using template
		if err := tmpl.Execute(os.Stdout, renderData{BuildLogEntry: log, Details: strings.TrimSpace(details)}); err != nil {
			return err
		}
	}

	return nil
}
