package command

/*
 * convert package setting to env.
 * ex:
 * root.setting => _ROOT__SETTING_
 * if 1st argument == root, will set root.setting to _SETTING_ too.
 */

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
)

type EnvCommand struct {
	Ui cli.Ui
}

func init() {
	_ENV = os.Environ()
	_HOME = os.Getenv("HOME")
}

func (c *EnvCommand) Help() string {
	helpText := `
Usage: rogi env pkg [filename]

`
	return strings.TrimSpace(helpText)
}

func (c *EnvCommand) Run(args []string) int {
	length := len(args)
	db := initDb()
	defer db.Close()

	rows, err := db.Queryx(LIST_ALL_SETTINGS)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var so SettingObject
		err := rows.StructScan(&so)

		if err != nil {
			continue
		}

		env := fmt.Sprintf("_%s__%s_=%s", strings.ToUpper(so.Package), strings.ToUpper(so.Key), so.Value)
		_ENV = append(_ENV, env)
		if length > 0 && args[0] == so.Package {
			env = fmt.Sprintf("_%s_=%s", strings.ToUpper(so.Key), so.Value)
			_ENV = append(_ENV, env)
		}
	}

	switch length {
	case 2:
		out, err := Run(args[1])
		if err != nil {
			log.Fatalf("Could not run (%s) error (%v)", args[1], err)
		}
		fmt.Printf("out: %s\n", out)
	default:
		for i, _ := range _ENV {
			fmt.Fprintf(os.Stdout, "%s\n", _ENV[i])
		}
	}

	return 0
}

func (c *EnvCommand) Synopsis() string {
	return "env"
}
