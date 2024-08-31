package cmds

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cragcraig/ccub/buildlog"
	"github.com/cragcraig/ccub/cli"
	"github.com/cragcraig/ccub/protos"
)

const humanReadableDate = "Mon Jan 02, 2006"

type startArgs struct {
	assembly string
}

var StartCmd = cli.ConstructCommand(
	cli.CommandMetadata{
		Description: "Start working",
	},
	parseStart,
	executeStart)

var StatusCmd = cli.ConstructCommand(
	cli.CommandMetadata{
		Description: "Status of work",
	},
	parseStatus,
	executeStatus)

var EditCmd = cli.ConstructCommand(
	cli.CommandMetadata{
		Description: "Edit log details",
	},
	parseEdit,
	executeEdit)

var StopCmd = cli.ConstructCommand(
	cli.CommandMetadata{
		Description: "Stop working",
	},
	parseStop,
	executeStop)

func durationMinToString(minutes int) string {
	return strings.TrimSuffix((time.Duration(minutes) * time.Minute).String(), "0s")
}

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
			if len(merged.WorkPeriod) > 0 {
				pw := merged.WorkPeriod[len(merged.WorkPeriod)-1]
				if len(pw.EndTime) == 0 {
					dur, err := buildlog.SameDayKitchenTimeDiff(entry.WorkPeriod[0].StartTime, pw.StartTime)
					if err != nil {
						return nil, err
					}

					return nil, fmt.Errorf(
						"Work period already ongoing, started at %s (%d minutes ago). Run 'stop' to end this work period.",
						pw.StartTime,
						uint32(math.Ceil(dur.Minutes())))
				}
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
		DetailsFile: buildlog.LogDetailsFile(now),
	}

	if err := buildlog.UpdateLogMetadataFile(buildlog.LogsPath, StartLogUpdater(&entry)); err != nil {
		return err
	}
	fmt.Printf("Started a new work period at %s\n\n", now.Format(time.Kitchen))
	fmt.Printf("Updated log file:   %s\n", buildlog.LogsPath)
	if f, err := buildlog.CreateLogDetailsFile(args.assembly, now, false); err != nil {
		return err
	} else {
		fmt.Printf("Details file:  %s\n", f)
		return buildlog.LaunchEditor(f)
	}
}

type statusArgs struct {
	date time.Time
}

func parseStatus(name string, argv []string) (*statusArgs, error) {
	args := &statusArgs{}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	// Raw flags
	date := flags.String("date", "today", "Date of work. Supported forms: "+strings.Join(buildlog.ValidDateFormats(), ", "))
	// Parse
	if err := flags.Parse(argv); err != nil {
		return nil, err
	}
	// Date
	if d, err := buildlog.ParseDateArg(*date); err != nil {
		return nil, err
	} else {
		args.date = d
	}
	return args, nil
}

func executeStatus(args *statusArgs) error {
	logs, err := buildlog.ReadLogs(buildlog.LogsPath)
	if err != nil {
		return err
	}

	exists, index := buildlog.LogExists(args.date, logs.LogEntry)
	if !exists {
		return fmt.Errorf("No log entry found for %s\n\nCreate a log entry using 'log' or 'start'\n", args.date.Format(humanReadableDate))
	}
	fmt.Printf("%s:  ", args.date.Format(humanReadableDate))
	entry := logs.LogEntry[index]
	if len(entry.WorkPeriod) > 0 {
		pw := entry.WorkPeriod[len(entry.WorkPeriod)-1]
		if len(pw.EndTime) == 0 {
			fmt.Printf("Ongoing work period started at %s\n\nRun 'stop' to end this work period\n", pw.StartTime)
		} else {
			total := 0
			for _, wp := range entry.WorkPeriod {
				total += int(wp.DurationMin)
			}
			fmt.Printf("Total logged work %s\n", durationMinToString(total))
		}
	} else {
		fmt.Printf("Log entry exists but without any work periods\n")
	}
	return nil
}

type editArgs struct {
	date time.Time
}

func parseEdit(name string, argv []string) (*editArgs, error) {
	args := &editArgs{}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	// Raw flags
	date := flags.String("date", "today", "Date of work. Supported forms: "+strings.Join(buildlog.ValidDateFormats(), ", "))
	// Parse
	if err := flags.Parse(argv); err != nil {
		return nil, err
	}
	// Date
	if d, err := buildlog.ParseDateArg(*date); err != nil {
		return nil, err
	} else {
		args.date = d
	}
	return args, nil
}

func executeEdit(args *editArgs) error {
	logs, err := buildlog.ReadLogs(buildlog.LogsPath)
	if err != nil {
		return err
	}

	exists, _ := buildlog.LogExists(args.date, logs.LogEntry)
	if !exists {
		return fmt.Errorf("No log entry found for %s. Create a log entry using 'log' or 'start'.", args.date.Format(humanReadableDate))
	}
	df := buildlog.LogDetailsFile(args.date)
	fmt.Printf("Editing details for %s\n\nDetails file:  %s\n", args.date.Format(humanReadableDate), df)
	return buildlog.LaunchEditor(df)
}

func parseStop(name string, argv []string) (*any, error) {
	return nil, nil
}

func StopLogUpdater(now time.Time) buildlog.LogUpdater {
	return func(logs []*protos.BuildLogEntry) ([]*protos.BuildLogEntry, error) {
		exists, index := buildlog.LogExists(now, logs)
		if !exists {
			return nil, fmt.Errorf("No log entry exists for today (%s)", now.Format(humanReadableDate))
		}
		merged := logs[index]
		if len(merged.WorkPeriod) == 0 {
			return nil, fmt.Errorf("Log entry exists for today but there are no work periods (weird), run 'start' to begin working")
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
		fmt.Printf("Stopped work period, %s to %s (%d minutes)\n", pw.StartTime, pw.EndTime, pw.DurationMin)
		total := 0
		for _, wp := range merged.WorkPeriod {
			total += int(wp.DurationMin)
		}
		fmt.Printf("Total time worked on %s:  %s\n", now.Format(humanReadableDate), durationMinToString(total))
		return logs, nil
	}
}

func executeStop(_ *any) error {
	now := time.Now()

	if err := buildlog.UpdateLogMetadataFile(buildlog.LogsPath, StopLogUpdater(now)); err != nil {
		return err
	}
	fmt.Printf("\nUpdated log entry for %s in log file:   %s\n", now.Format(humanReadableDate), buildlog.LogsPath)
	return nil
}
