package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

type CleanCommand struct {
	UI cli.Ui
}

func (c *CleanCommand) Help() string {
	helpText := `
Usage: rogi clean

`
	return strings.TrimSpace(helpText)
}

// Run will return integer as true or false.
func (c *CleanCommand) Run(args []string) int {
	return 0
}

func (c *CleanCommand) Synopsis() string {
	return "env"
}
