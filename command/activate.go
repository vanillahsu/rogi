package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

type ActivateCommand struct {
	UI cli.Ui
}

func (c *ActivateCommand) Help() string {
	helpText := `
Usage: rogi activate [pkg name]

`
	return strings.TrimSpace(helpText)
}

// Run will return integer as true or false.
func (c *ActivateCommand) Run(args []string) int {
	return 0
}

func (c *ActivateCommand) Synopsis() string {
	return "activate"
}
