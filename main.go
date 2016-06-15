package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/mitchellh/cli"
	"github.com/vanillahsu/rogi/command"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	log.SetOutput(ioutil.Discard)

	args := os.Args[1:]

	for _, arg := range args {
		if arg == "-v" || arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs
			break
		}
	}

	command.InitEnv()

	cli := &cli.CLI{
		Args:     args,
		Commands: Commands,
		HelpFunc: cli.BasicHelpFunc("rogi"),
	}

	exitCode, err := cli.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}
