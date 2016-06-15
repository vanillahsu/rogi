package main

import (
	"os"
	"os/signal"

	"github.com/mitchellh/cli"
	"github.com/vanillahsu/rogi/command"
)

var Commands map[string]cli.CommandFactory

func init() {
	ui := &cli.BasicUi{Writer: os.Stdout}

	Commands = map[string]cli.CommandFactory{
		"env": func() (cli.Command, error) {
			return &command.EnvCommand{
				UI: ui,
			}, nil
		},

		"set": func() (cli.Command, error) {
			return &command.SetCommand{
				UI: ui,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Ui:       ui,
				Version:  Version,
				Revision: GitCommit,
			}, nil
		},
	}
}

func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}
