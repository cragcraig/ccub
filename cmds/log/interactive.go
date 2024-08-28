package log

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"time"

	"github.com/cragcraig/ccub/buildlog"
	"github.com/cragcraig/ccub/cli"
	"github.com/cragcraig/ccub/protos"
)

type startArgs struct {
	assembly string
}

var StartCmd = cli.ConstructCommand(
	cli.CommandMetadata{
		Description: "Start working in an interactive build log session",
	},
	parseStart,
	executeStart)

var EditCmd = cli.ConstructCommand(
	cli.CommandMetadata{
		Description: "Edit details for an interactive build log session",
	},
	parseEdit,
	executeEdit)

var StopCmd = cli.ConstructCommand(
	cli.CommandMetadata{
		Description: "Stop working in an interactive build log session",
	},
	parseStop,
	executeStop)

func parseStart(name string, argv []string) (*startArgs, error) {
	args := &startArgs{}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	// Raw flags
	assembly := flags.String("assembly", "", "Top-level assembly; if not set assumes unchanged from the prior log entry.")
	// Parse
	if err := flags.Parse(argv); err != nil {
		return nil, err
	}
	// Assembly
	if len(*assembly) > 0 {
		if a, err := buildlog.ParseAssemblyArg(*assembly); err != nil {
			return nil, err
		} else {
			args.assembly = a
		}
	}
	return args, nil
}

func StartLogUpdater(entry *protos.BuildLogEntry) buildlog.LogUpdater {
	return func(logs []*protos.BuildLogEntry) ([]*protos.BuildLogEntry, error) {
		date, err := buildlog.ParseDateOfLog(entry)
		if err != nil {
			return logs, err
		}
		exists, index := buildlog.LogExists(date, logs)
		if exists {
			merged := logs[index]
			pw := merged.WorkPeriod[len(merged.WorkPeriod)-1]
			if len(pw.EndTime) == 0 {
				dm, err := buildlog.SameDayKitchenTimeDiff(entry.WorkPeriod[0].StartTime, pw.StartTime)
				if err != nil {
					return nil, err
				}

				return nil, fmt.Errorf(
					"Work period already started at %s (%d min ago), run 'stop' to finish",
					pw.StartTime,
					uint32(math.Ceil(dm.Minutes())))
			}
			merged.WorkPeriod = append(merged.WorkPeriod, entry.WorkPeriod[0])
		} else {
			if len(entry.Assembly) == 0 {
				if len(logs) == 0 {
					return nil, errors.New("Assembly not specified but also no previous log entry exists from which to inherit")
				}
				// assume unchanged from the prior log entry
				entry.Assembly = logs[len(logs)-1].Assembly
			}
			logs = append(logs, entry)
		}
		return logs, nil
	}
}

func executeStart(args *startArgs) error {
	now := time.Now()
	entry := protos.BuildLogEntry{
		Assembly: args.assembly,
		Date:     buildlog.FormatDateForLog(now),
		WorkPeriod: []*protos.TimePeriod{
			{
				StartTime:   now.Format(time.Kitchen),
				EndTime:     "",
				DurationMin: 0,
			}},
		DetailsFile: buildlog.LogDetailsFile([]string{}, now),
	}

	if err := buildlog.UpdateLogMetadataFile(buildlog.LogsPath, StartLogUpdater(&entry)); err != nil {
		return err
	}
	fmt.Printf("Started work period at %s:   %s\n", now.Format(time.Kitchen), buildlog.LogsPath)
	if f, err := buildlog.CreateLogDetailsFile(args.assembly, now, false); err != nil {
		return err
	} else {
		fmt.Printf("Details:  %s\n", f)
		return buildlog.LaunchEditor(f)
	}
}

func parseEdit(name string, argv []string) (*any, error) {
	return nil, nil
}

func executeEdit(_ *any) error {
	now := time.Now()
	logs, err := buildlog.ReadLogs(buildlog.LogsPath)
	if err != nil {
		return err
	}

	exists, _ := buildlog.LogExists(now, logs.LogEntry)
	if !exists {
		return fmt.Errorf("Log entry for %s has not been created or started yet", buildlog.FormatDateForLog(now))
	}
	df := buildlog.LogDetailsFile([]string{buildlog.LogsDir}, now)
	fmt.Printf("Editing details:  %s\n", df)
	return buildlog.LaunchEditor(df)
}

func parseStop(name string, argv []string) (*any, error) {
	return nil, nil
}

func StopLogUpdater(now time.Time) buildlog.LogUpdater {
	return func(logs []*protos.BuildLogEntry) ([]*protos.BuildLogEntry, error) {
		exists, index := buildlog.LogExists(now, logs)
		if !exists {
			return nil, fmt.Errorf("No log entry exists for today (%s)", now.Format(time.Kitchen))
		}
		merged := logs[index]
		if len(merged.WorkPeriod) == 0 {
			return nil, fmt.Errorf("Log entry exists for today but there are no work periods (weird). Run 'start' to begin working.")
		}
		pw := merged.WorkPeriod[len(merged.WorkPeriod)-1]
		if len(pw.EndTime) != 0 {
			return nil, fmt.Errorf("Work period already stopped at %s, run 'start' to begin working again", pw.StartTime)
		}
		pw.EndTime = now.Format(time.Kitchen)
		dm, err := buildlog.SameDayKitchenTimeDiff(pw.EndTime, pw.StartTime)
		if err != nil {
			return nil, err
		}
		pw.DurationMin = uint32(math.Ceil(dm.Minutes()))
		fmt.Printf("Stopped work period, %s to %s (%d min):   %s\n", pw.StartTime, pw.EndTime, pw.DurationMin, buildlog.LogsPath)
		return logs, nil
	}
}

func executeStop(_ *any) error {
	now := time.Now()

	if err := buildlog.UpdateLogMetadataFile(buildlog.LogsPath, StopLogUpdater(now)); err != nil {
		return err
	}
	return nil
}
