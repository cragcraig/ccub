package cli

import (
	"fmt"
)

func InitCmd(cmd CommandEntry, argv []string) error {
	if len(argv) != 2 {
		return cmd.getUsageError()
	}
	fmt.Println("TODO: Init")
	return nil
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
