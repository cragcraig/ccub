package cli

import (
	"errors"
	"flag"
	"fmt"
)

const cliCmdName = "ccub"
const helpCmdName = "help"

type CommandMetadata struct {
    Name string
    Description string
}

type CommandFactory interface {
    Metadata() CommandMetadata
    Create(name string) Command
}

type Command interface {
    Parse(args []string) error
    Execute() error
}

var commands = map[string]CommandFactory{
	"log": &logCmdFactory{},
/*
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
		eg:    []string{"", ""},
	},
    */
	// Normal commands
    /*
	"render": CommandEntry{
		name:  "render",
		cmd:   RenderCmd,
		desc:  "Render logs to Markdown using the provided template",
		usage: "TEMPLATE_FILE",
		eg: []string{
			"template.md",
		},
	},
    */
}
/*
func (cmd CommandEntry) getUsageError() error {
	msg := []string{fmt.Sprintf("Usage:  %s %s %s", cliCmdName, cmd.name, cmd.usage)}
	prefix := " e.g.,"
	for _, eg := range cmd.eg {
		msg = append(msg, fmt.Sprintf("%s  %s %s %s", prefix, cliCmdName, cmd.name, eg))
		prefix = "      "
	}
	return errors.New(strings.Join(msg, "\n"))
}
*/
func Exec(cmdName string, argv []string) error {
	if cmdName == helpCmdName || cmdName == "" {
		// Help
		return help(commands, argv)
	} else if factory, exists := commands[cmdName]; exists {
		// All other commands
        cmd := factory.Create(cmdName)
        if err := cmd.Parse(argv); err != nil {
            if errors.Is(err, flag.ErrHelp) {
                return nil
            }
            return err
        }
        return cmd.Execute()
	} else {
		// Unrecognized
		return fmt.Errorf("Unrecognized command \"%s\", try \"help\"", cmdName)
	}
}

func printVersion() {
	fmt.Println("Carbon Cub Build Log, version 0.19")
}

func help(commands map[string]CommandFactory, argv []string) error {
	if len(argv) == 0 {
		printVersion()
		fmt.Println("")
		fmt.Printf("Usage:  %s COMMAND [-flag1 value] [-flag2 value] ...\n", cliCmdName)
		fmt.Printf(" e.g.,  %s log -help\n", cliCmdName)
		fmt.Printf("        %s log -assembly \"left wing\" -date today -time 1pm-3:15pm\n", cliCmdName)
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
		for name, factory := range commands {
			fmt.Printf("  %-*s  %s\n", max, name, factory.Metadata().Description)
		}
		return nil
	} else {
		cmdName := argv[0]
		if _, exists := commands[cmdName]; exists {
			return fmt.Errorf("Try '%s -help'", cmdName)
		}
		return errors.New(fmt.Sprintf("Unable: '%s' is not a supported command", cmdName))
	}
}
