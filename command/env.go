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

	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"
	"github.com/mitchellh/cli"
)

type EnvCommand struct {
	UI cli.Ui
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

// Run will return integer as true or false.
func (c *EnvCommand) Run(args []string) int {
	length := len(args)
	db := initDb()
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(SETTING_BUCKET))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		cur := b.Cursor()

		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			var x settings
			if _, err := toml.Decode(string(v), &x); err != nil {
				continue
			}

			for y, z := range x {
				env := fmt.Sprintf("_%s__%s=%s", strings.ToUpper(string(k)), strings.ToUpper(y), strings.ToUpper(z))
				_ENV = append(_ENV, env)
			}
		}

		return nil
	})

	switch length {
	case 2:
		out, err := Run(args[1])
		if err != nil {
			msg := fmt.Sprintf("Could not run (%s) error (%v)", args[1], err)
			c.UI.Warn(msg)
		}
		fmt.Printf("out: %s\n", out)
	default:
		for i := range _ENV {
			fmt.Fprintf(os.Stdout, "%s\n", _ENV[i])
		}
	}

	return 0
}

func (c *EnvCommand) Synopsis() string {
	return "env"
}
