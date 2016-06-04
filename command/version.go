package command

import (
	"bytes"
	"fmt"

	"github.com/mitchellh/cli"
)

type VersionCommand struct {
	Version  string
	Revision string
	Ui       cli.Ui
}

func (c *VersionCommand) Help() string {
	return ""
}

func (c *VersionCommand) Run(_ []string) int {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "rogi v%s", c.Version)

	c.Ui.Output(versionString.String())

	return 0
}

func (c *VersionCommand) Synopsis() string {
	return "version"
}
