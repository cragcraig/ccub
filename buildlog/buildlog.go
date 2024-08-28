package buildlog

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cragcraig/ccub/protos"
	"github.com/golang/protobuf/proto"
)

const (
	LogsDir            = "log"
	LogsFile           = "buildlog.textproto"
	LogsPath           = LogsDir + "/" + LogsFile
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

func LogExists(date time.Time, logs []*protos.BuildLogEntry) (exists bool, index int) {
	d := FormatDateForLog(date)
	for i, v := range logs {
		if v.Date == d {
			return true, i
		}
	}
	return false, -1
}

func ReadLogs(f string) (*protos.BuildLogs, error) {
	text, err := ReadFile(f)
	if err != nil {
		return &protos.BuildLogs{}, err
	}
	entries := protos.BuildLogs{}
	err = proto.UnmarshalText(text, &entries)
	return &entries, err
}

func WriteLogs(f string, logs *protos.BuildLogs) error {
	err := EnsureDirExists(LogsDir)
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

func LogDetailsDir(basedir []string, date time.Time) string {
	return strings.Join(append(basedir, date.Format(MonthLayout)), "/")
}

func LogDetailsFile(basedir []string, date time.Time) string {
	return strings.Join([]string{LogDetailsDir(basedir, date), date.Format(DateLayout) + ".md"}, "/")
}

func CreateLogDetailsFile(assembly string, date time.Time, overwrite bool) (string, error) {
	if err := EnsureDirExists(LogDetailsDir([]string{LogsDir}, date)); err != nil {
		return "", err
	}
	f := LogDetailsFile([]string{LogsDir}, date)
	if !overwrite {
		if exists, err := FileExists(f); err != nil {
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

type LogUpdater func(logs []*protos.BuildLogEntry) ([]*protos.BuildLogEntry, error)

func InsertLogUpdater(entry *protos.BuildLogEntry) LogUpdater {
	return func(logs []*protos.BuildLogEntry) ([]*protos.BuildLogEntry, error) {
		date, err := ParseDateOfLog(entry)
		if err != nil {
			return logs, err
		}
		exists, _ := LogExists(date, logs)
		if exists {
			return logs, fmt.Errorf("Log entry already exists for %s", entry.Date)
		}
		return append(logs, entry), nil
	}
}

func UpsertLogUpdater(entry *protos.BuildLogEntry) LogUpdater {
	return func(logs []*protos.BuildLogEntry) ([]*protos.BuildLogEntry, error) {
		date, err := ParseDateOfLog(entry)
		if err != nil {
			return logs, err
		}
		exists, index := LogExists(date, logs)
		if exists {
			logs[index] = entry
		} else {
			logs = append(logs, entry)
		}
		return logs, nil
	}
}

func UpdateLogMetadataFile(f string, update LogUpdater) error {
	logs, err := ReadLogs(f)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Could not open logs metadata from %s\n%s", f, err.Error())
	}

	logs.LogEntry, err = update(logs.LogEntry)
	if err != nil {
		return err
	}

	// Ensure logs are ordered by date
	sort.Slice(logs.LogEntry, func(i, j int) bool {
		it, _ := ParseDateOfLog(logs.LogEntry[i])
		jt, _ := ParseDateOfLog(logs.LogEntry[j])
		return jt.After(it)
	})

	return WriteLogs(f, logs)
}

func ParseAssemblyArg(arg string) (string, error) {
	if !containsString(validAssemblies, arg) {
		return "", fmt.Errorf("Assembly must be one of:\n  %s", strings.Join(validAssemblies, "\n  "))
	}
	return arg, nil
}
