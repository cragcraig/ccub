package cli

import (
	"errors"
	"fmt"
	"strings"
)

const cliCmdName = "cclog"
const helpCmdName = "help"

type CommandFunc func(CommandEntry, []string) error

type CommandEntry struct {
	name  string
	cmd   CommandFunc
	desc  string
	usage string
	eg    []string
}

var commands = map[string]CommandEntry{
	// Meta commands
	"version": CommandEntry{
		name: "version",
		cmd: func(_ CommandEntry, _ []string) error {
			printVersion()
			return nil
		},
		desc:  "Report the version",
		usage: "",
	},
	helpCmdName: CommandEntry{
		name:  helpCmdName,
		cmd:   nil, // special case to avoid circular dep
		desc:  "Provide help documentation",
		usage: "[COMMAND]",
		eg:    []string{"", "log now riveting main spar"},
	},
	// Normal commands
	"new": CommandEntry{
		name:  "new",
		cmd:   NewLogCmd,
		desc:  "Initialize a new build log entry",
		usage: "today|DATE",
		eg:    []string{""},
	},
}

func (cmd CommandEntry) getUsageError() error {
	msg := []string{fmt.Sprintf("Usage:  %s %s %s", cliCmdName, cmd.name, cmd.usage)}
	prefix := " e.g.,"
	for _, eg := range cmd.eg {
		msg = append(msg, fmt.Sprintf("%s  %s %s %s", prefix, cliCmdName, cmd.name, eg))
		prefix = "      "
	}
	return errors.New(strings.Join(msg, "\n"))
}

func Exec(cmdName string, argv []string) error {
	if cmdName == helpCmdName || cmdName == "" {
		// Help
		return help(commands, argv)
	} else if c, exists := commands[cmdName]; exists {
		// All other commands
		return c.cmd(c, argv)
	} else {
		// Unrecognized
		return fmt.Errorf("Unrecognized command \"%s\", try \"help\"", cmdName)
	}
}

func printVersion() {
	fmt.Println("Carbon Cub Build Logger, version 0.1")
}

func help(commands map[string]CommandEntry, argv []string) error {
	if len(argv) == 0 {
		printVersion()
		fmt.Println("")
		fmt.Printf("Usage:  %s COMMAND [ARG1] [ARG2...]\n", cliCmdName)
		fmt.Printf(" e.g.,  %s help init\n", cliCmdName)
		fmt.Printf("        %s log now Riveting main spar\n", cliCmdName)
		fmt.Println("")
		fmt.Println("Commands:")
		// Get length of the longest command
		max := 0
		for k, _ := range commands {
			if l := len(k); l > max {
				max = l
			}
		}
		// Print all commands with descriptions
		for _, cmd := range commands {
			fmt.Printf("  %-*s  %s\n", max, cmd.name, cmd.desc)
		}
		return nil
	} else {
		cmdName := argv[0]
		if cmd, exists := commands[cmdName]; exists {
			fmt.Println(strings.ToUpper(cmd.name), "-", cmd.desc)
			fmt.Println(cmd.getUsageError())
			return nil
		}
		return errors.New(fmt.Sprintf("Unable: '%s' is not a supported command", cmdName))
	}
}
