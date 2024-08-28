package cli

import (
	"errors"
	"flag"
	"fmt"
)

const helpCmdName = "help"

type CommandMetadata struct {
	Description string
}

type Command interface {
	Metadata() CommandMetadata
	ParseArgsAndExecute(name string, argv []string) error
}

func ConstructCommand[T any](metadata CommandMetadata, parseArgs func(name string, args []string) (T, error), execute func(args T) error) Command {
	return commandTmpl[T]{
		metadata:  metadata,
		parseArgs: parseArgs,
		execute:   execute,
	}
}

type commandTmpl[T any] struct {
	metadata  CommandMetadata
	parseArgs func(name string, argv []string) (T, error)
	execute   func(args T) error
}

func (cmd commandTmpl[T]) Metadata() CommandMetadata {
	return cmd.metadata
}

func (cmd commandTmpl[T]) ParseArgsAndExecute(name string, argv []string) error {
	if args, err := cmd.parseArgs(name, argv); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	} else {
		return cmd.execute(args)
	}
}

func Exec(commands map[string]Command, cliName string, cmdName string, argv []string) error {
	if cmdName == helpCmdName || cmdName == "" {
		// Help
		return help(commands, cliName, argv)
	} else if cmd, exists := commands[cmdName]; exists {
		// All other commands
		return cmd.ParseArgsAndExecute(cmdName, argv)
	} else {
		// Unrecognized
		return fmt.Errorf("Unrecognized command \"%s\", try \"help\"", cmdName)
	}
}

func printVersion() {
	fmt.Println("Carbon Cub Build Log, version 0.20")
}

func help(commands map[string]Command, cliName string, argv []string) error {
	if len(argv) == 0 {
		printVersion()
		fmt.Println("")
		fmt.Printf("Usage:  %s COMMAND [-flag1 value] [-flag2 value] ...\n", cliName)
		fmt.Printf(" e.g.,  %s log -help\n", cliName)
		fmt.Printf("        %s log -assembly \"left wing\" -date today -time 1pm-3:15pm\n", cliName)
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
		for name, cmd := range commands {
			fmt.Printf("  %-*s  %s\n", max, name, cmd.Metadata().Description)
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
