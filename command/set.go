package command

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"
	"github.com/mitchellh/cli"
)

type SetCommand struct {
	UI cli.Ui
}

func (c *SetCommand) Help() string {
	return ""
}

func (c *SetCommand) Run(args []string) int {
	/*
	 * usage: pkg.key = value
	 *      empty string  => show all settings
	 *      pkg           => show all settings of pkg
	 *      pkg.key       => show setting of pkg with key
	 *      pkg.key=value => change setting of pkg.key
	 */
	db := initDb()
	defer db.Close()

	if len(args) == 0 {
		listSettings(db, os.Stdout)
	} else {
		for _, arg := range args {
			pkg, key, value := extractValue(arg)
			if value != "" {
				changeSetting(db, pkg, key, value)
			} else {
				listSettings(db, os.Stdout, pkg, key)
			}
		}

	}

	return 0
}

func (c *SetCommand) Synopsis() string {
	return "set"
}

func listAllSettings(db *bolt.DB, w io.Writer) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(SETTING_BUCKET))

		if b == nil {
			return fmt.Errorf("bucket %s not exist", SETTING_BUCKET)
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var x settings

			if _, err := toml.Decode(string(v), &x); err != nil {
				return fmt.Errorf("toml.Decode: %s", err)
			}

			for y, z := range x {
				fmt.Fprintf(w, "%s.%s=%s\n", string(k), y, z)
			}
		}

		return nil
	})
}

func listPkgSettings(db *bolt.DB, w io.Writer, pkg string) error {
	return db.View(func(tx *bolt.Tx) error {
		var x settings

		b := tx.Bucket([]byte(SETTING_BUCKET))

		if b == nil {
			return fmt.Errorf("bucket %s not exist", SETTING_BUCKET)
		}

		c := b.Get([]byte(pkg))

		if _, err := toml.Decode(string(c), &x); err != nil {
			return fmt.Errorf("toml.Decode: %s", err)
		}

		for y, z := range x {
			fmt.Fprintf(w, "%s.%s=%s\n", pkg, y, z)
		}

		return nil
	})
}

func listOneSetting(db *bolt.DB, w io.Writer, pkg, key string) error {
	return db.View(func(tx *bolt.Tx) error {
		var x settings

		b := tx.Bucket([]byte(SETTING_BUCKET))

		if b == nil {
			return fmt.Errorf("bucket %s not exist", SETTING_BUCKET)
		}

		c := b.Get([]byte(pkg))

		if _, err := toml.Decode(string(c), &x); err != nil {
			return fmt.Errorf("toml.Decode: %s", err)
		}

		if _, ok := x[key]; ok {
			fmt.Fprintf(w, "%s.%s=%s\n", pkg, key, x[key])
		} else {
			fmt.Fprintf(w, "Could not found %s.%s", pkg, key)
		}

		return nil
	})
}

func listSettings(db *bolt.DB, w io.Writer, args ...string) {
	if len(args) > 0 {
		if args[0] != "" && args[1] != "" {
			listOneSetting(db, w, args[0], args[1])
		} else if args[0] != "" && args[1] == "" {
			listPkgSettings(db, w, args[0])
		}
	} else {
		listAllSettings(db, w)
	}
}

func changeSetting(db *bolt.DB, pkg, key, value string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(SETTING_BUCKET))
		if err != nil {
			return fmt.Errorf("create bucket (%s)", err)
		}

		c := b.Get([]byte(pkg))

		var x settings
		if _, err := toml.Decode(string(c), &x); err != nil {
			return fmt.Errorf("toml.Decode: %s", err)
		}

		x[key] = value

		var buffer bytes.Buffer
		e := toml.NewEncoder(&buffer)
		err = e.Encode(x)

		if err != nil {
			return fmt.Errorf("toml.Encode: %s", err)
		}

		err = b.Put([]byte(pkg), buffer.Bytes())

		return err
	})
}
