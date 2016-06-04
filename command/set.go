package command

import (
	"fmt"
	"io"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/cli"
)

type SetCommand struct {
	Ui cli.Ui
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

func listSettings(db *sqlx.DB, w io.Writer, args ...string) {
	var rows *sqlx.Rows
	var query string
	var err error

	fmt.Printf("args: %v\n", args)
	switch len(args) {
	case 2:
		query = LIST_SETTING
	case 1:
		query = LIST_SETTINGS
	case 0:
		query = LIST_ALL_SETTINGS
	}

	rows, err = db.Queryx(query, args[0], args[1])
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	fmt.Printf("%s\n", query)
	for rows.Next() {
		var so SettingObject
		rows.StructScan(so)
		fmt.Fprintf(w, "%s.%s=%s\n", so.Package, so.Key, so.Value)
	}
}

func changeSetting(db *sqlx.DB, pkg, key, value string) {
	var cc int32
	var query string

	row := db.QueryRow(COUNT_SETTING, pkg, key)
	row.Scan(&cc)
	if cc == 1 {
		query = CHANGE_SETTING
	} else {
		query = INSERT_SETTING
	}
	_, err := db.Exec(query, value, pkg, key)
	if err != nil {
		panic(err)
	}
}
