package main

import (
	"fmt"
	"github.com/cragcraig/ccub/cli"
	"github.com/cragcraig/ccub/cmds/log"
	"os"
	"strings"
)

const cliName = "ccub"

var commands = map[string]cli.CommandFactory{
	"log": log.NewFactory(),
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
