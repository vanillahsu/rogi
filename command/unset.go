package command

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"
	"github.com/mitchellh/cli"
)

type UnsetCommand struct {
	UI cli.Ui
}

func (c *UnsetCommand) Help() string {
	helpText := `
Usage: rogi unset pkg.setting

`
	return strings.TrimSpace(helpText)
}

// Run will return integer as true or false.
func (c *UnsetCommand) Run(args []string) int {
	db := initDb()
	defer db.Close()

	if len(args) != 0 {
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(SETTING_BUCKET))

			if b == nil {
				return fmt.Errorf("bucket %s not exist", SETTING_BUCKET)
			}

			for _, arg := range args {
				pkg, key, _ := extractValue(arg)

				d := b.Get([]byte(pkg))
				if d == nil {
					continue
				}

				var x settings
				if _, err := toml.Decode(string(d), &x); err != nil {
					continue
				}

				delete(x, key)

				var buffer bytes.Buffer
				e := toml.NewEncoder(&buffer)
				err := e.Encode(x)

				if err != nil {
					continue
				}

				b.Put([]byte(pkg), buffer.Bytes())
			}

			return nil
		})
	}

	return 0
}

func (c *UnsetCommand) Synopsis() string {
	return "unset"
}
