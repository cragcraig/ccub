package buildlog

import (
	"fmt"
	"github.com/cragcraig/ccub/protos"
	"github.com/golang/protobuf/proto"
	"os"
	"sort"
	"strings"
	"time"
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

func LogExists(date time.Time, logs *protos.BuildLogs) (exists bool, index int) {
	if logs.LogEntry == nil {
		return false, -1
	}
	d := FormatDateForLog(date)
	for i, v := range logs.LogEntry {
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

func UpdateLogMetadataFile(f string, entry *protos.BuildLogEntry, overwrite bool) error {
	logs, err := ReadLogs(f)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Could not open logs metadata from %s\n", f, err.Error())
	}

	date, err := ParseDateOfLog(entry)
	if err != nil {
		return err
	}
	exists, index := LogExists(date, logs)
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
