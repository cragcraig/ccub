package cli

import (
	"fmt"
    "io/ioutil"
    "os"
    "github.com/golang/protobuf/proto"
    "time"
    "strings"
    ccprotos "github.com/cragcraig/carboncub-build-log/protos"
)

const MonthLayout = "2006-Jan"
const DateLayout = "2006-Jan-02"

func readLogs(string filename) (ccproto.BuildLogs, error) {
    fp, err := os.Open(logfn)
    if err != nil {
        return ccproto.BuildLogs{}, err
    }
    defer fp.Close()
    data, err := ioutil.ReadAll(fp)
    if err != nil {
        return ccproto.BuildLogs{}, err
    }
    text := string(data)
    entries := ccproto.BuildLogs{}
    err = proto.UnmarshallText(text, &entries)
    return entries, err
}

func writeLogs(string filename, ccproto.BuildLogs) error {
    fp, err := os.Create("logs/" + date.Format(MonthLayout) + "/logs-new.textproto")
    if err != nil {
        err
    }
    defer fp.Close()
    return proto.MarshalText(fp, &ex)
}

func parseDateArg(string arg) (time.Time, error) {
    if arg == "today" {
        return time.Now(), nil
    }
    return time.Parse(DateLayout, arg)
}

func NewLogCmd(cmd CommandEntry, argv []string) error {
	if len(argv) != 2 {
		return cmd.getUsageError()
	}
    d, err := parseDateArg(argv[1])
    if err != nil {
        return err
    }

    date := d.Date()
    assembly := argv[2]
    tags := strings.Split(argv[3], ",")
    
    entry := ccproto.BuildLogEntryMetadata{
        assembly: assembly
        date: date.Format(DateLayout)
        log_filename: date.Format(DateLayout) + ".md"
        tags: tags
    }

    fmt.Printf("%v\n", entry)

    logfile := "logs/" + date.Format(MonthLayout) + "/logs.textproto"
    logs, err := readLogs(logfile)
    if err != nil {
        return err
    }
    logs.log_entry = append(prev.log_entry, entry)
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
