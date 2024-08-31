package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cragcraig/ccub/cli"
	"github.com/cragcraig/ccub/cmds"
)

const cliName = "ccub"

var commands = map[string]cli.Command{
	"log":    cmds.LogCmd,
	"start":  cmds.StartCmd,
	"stop":   cmds.StopCmd,
	"status": cmds.StatusCmd,
	"edit":   cmds.EditCmd,
	"render": cmds.RenderCmd,
}

func main() {
	var cmdName string
	var args []string

	if len(os.Args) > 1 {
		cmdName = os.Args[1]
		args = os.Args[2:]
	}
	if err := cli.Exec(commands, cliName, strings.ToLower(cmdName), args); err != nil {
		fmt.Println(err)
	}
}
