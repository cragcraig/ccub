package cli

import (
	"fmt"
    "io/ioutil"
    "os"
    "github.com/golang/protobuf/proto"
    "time"
    "strings"
    "github.com/cragcraig/ccub/logprotos"
)

const MonthLayout = "2006-Jan"
const DateLayout = "2006-Jan-02"

func readLogs(filename string) (*logprotos.BuildLogs, error) {
    fp, err := os.Open(filename)
    if err != nil {
        return &logprotos.BuildLogs{}, err
    }
    defer fp.Close()
    data, err := ioutil.ReadAll(fp)
    if err != nil {
        return &logprotos.BuildLogs{}, err
    }
    text := string(data)
    entries := logprotos.BuildLogs{}
    err = proto.UnmarshalText(text, &entries)
    return &entries, err
}

func writeLogs(filename string, logs *logprotos.BuildLogs) error {
    fp, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer fp.Close()
    return proto.MarshalText(fp, logs)
}

func parseDateArg(arg string) (time.Time, error) {
    if arg == "today" {
        return time.Now(), nil
    }
    return time.Parse(DateLayout, arg)
}

func NewLogCmd(cmd CommandEntry, argv []string) error {
	if len(argv) != 3 {
		return cmd.getUsageError()
	}
    date, err := parseDateArg(argv[0])
    if err != nil {
        return err
    }

    assembly := argv[1]
    tags := strings.Split(argv[2], ",")
    
    entry := logprotos.BuildLogEntryMetadata{
        Assembly: assembly,
        Date: date.Format(DateLayout),
        LogFile: date.Format(DateLayout) + ".md",
        Tags: tags,
    }

    fmt.Printf("%v\n", entry)

    logfile := "logs/" + date.Format(MonthLayout) + "/logs.textproto"
    /*
    logs, err := readLogs(logfile)
    if err != nil {
        fmt.Printf("Could not open %s: %s\n", logfile, err.Error())
    }
    */
    logs := &logprotos.BuildLogs{}
    logs.LogEntry = append(logs.LogEntry, &entry)
    return writeLogs(logfile, logs)
}

func LogCmd(cmd CommandEntry, argv []string) error {
	if len(argv) != 1 {
		return cmd.getUsageError()
	}
	fmt.Println("TODO: Log")
	return nil
}

func EndCmd(cmd CommandEntry, argv []string) error {
	if len(argv) != 1 {
		return cmd.getUsageError()
	}
	fmt.Println("TODO: End")
	return nil
}
