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

var CmdFactory = cli.ConstructCommand(
	cli.CommandMetadata{
		Description: "Render build logs using a user-specified template",
	},
	parse,
	execute)

type renderArgs struct {
	tmplFile string
}

func parse(name string, argv []string) (*renderArgs, error) {
	args := &renderArgs{}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	// Raw flags
	tmplFile := flags.String("tmpl", "", "Template text file. Required.")
	// Parse
	if err := flags.Parse(argv); err != nil {
		return nil, err
	}
	// Template
	if len(*tmplFile) == 0 {
		return nil, errors.New("'tmpl' is required")
	}
	args.tmplFile = *tmplFile
	return args, nil
}

func execute(args *renderArgs) error {
	tmpl, err := buildlog.LoadTemplateFromFile(args.tmplFile)
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
		data := struct {
			*protos.BuildLogEntry
			Details string
		}{
			BuildLogEntry: log,
			Details:       strings.TrimSpace(details),
		}
		if err := tmpl.Execute(os.Stdout, data); err != nil {
			return err
		}
	}

	return nil
}
