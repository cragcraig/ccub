package cli

import (
	"errors"
	"flag"
	"fmt"
)

const helpCmdName = "help"

type Command interface {
	Execute() error
}

type CommandMetadata struct {
	Description string
}

type CommandFactory interface {
	Metadata() CommandMetadata
	Create(name string, args []string) (Command, error)
}

func NewCommandFactory(metadata CommandMetadata, factoryFn func(name string, args []string) (Command, error)) CommandFactory {
	return &staticCommandFactory{
		metadata:  metadata,
		factoryFn: factoryFn,
	}
}

type staticCommandFactory struct {
	metadata  CommandMetadata
	factoryFn func(name string, args []string) (Command, error)
}

func (m *staticCommandFactory) Metadata() CommandMetadata {
	return m.metadata
}

func (m *staticCommandFactory) Create(name string, args []string) (Command, error) {
	return m.factoryFn(name, args)
}

func Exec(commands map[string]CommandFactory, cliName string, cmdName string, argv []string) error {
	if cmdName == helpCmdName || cmdName == "" {
		// Help
		return help(commands, cliName, argv)
	} else if factory, exists := commands[cmdName]; exists {
		// All other commands
		if cmd, err := factory.Create(cmdName, argv); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		} else {
			return cmd.Execute()
		}
	} else {
		// Unrecognized
		return fmt.Errorf("Unrecognized command \"%s\", try \"help\"", cmdName)
	}
}

func printVersion() {
	fmt.Println("Carbon Cub Build Log, version 0.19")
}

func help(commands map[string]CommandFactory, cliName string, argv []string) error {
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
