package command

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/jmoiron/sqlx"
)

// Debug shows debug messages in functions like Run.
var Debug bool

// == Errors
var (
	errEnvVar      = errors.New("the format of the variable has to be VAR=value")
	errNoCmdInPipe = errors.New("no command around of pipe")
)

func InitEnv() {
	if err := syscall.Access(DEFAULT_PREFIX, 0755); err != nil {
		syscall.Mkdir(DEFAULT_PREFIX, 0755)
	}

	datafile := DEFAULT_PREFIX + DATAFILE

	if err := syscall.Access(datafile, 0644); err != nil {
		db := initDb()
		defer db.Close()

		db.Exec(SETS_TABLE)
		//db.Exec(STARTS_TABLE)
	} else {
		log.Fatalf("Fail to create database file (%s)\n", err)
	}
}

func initDb() *sqlx.DB {
	datafile := DEFAULT_PREFIX + DATAFILE
	engine, err := sqlx.Connect("sqlite3", datafile)
	if err != nil {
		panic(err)
	}

	return engine
}

func Unlock() {
	lockfile := DEFAULT_PREFIX + LOCKFILE

	if err := syscall.Access(lockfile, 0644); err == nil {
		os.Remove(lockfile)
		fmt.Printf("unlock: %s\n", lockfile)
	}
}

func splitWithSpaceAndQuote(arg string) (s []string) {
	tmp := arg
	for {
		idx := strings.Index(tmp, " ")
		if idx < 0 {
			s = append(s, tmp)
			break
		}

		buf := tmp[:idx]
		if buf[0] == '`' {
			idx2 := strings.Index(tmp[idx+1:], "`")
			if idx < 0 {
				panic("no matched backquote")
			}
			idx = idx + 1 + idx2 + 1
			buf = tmp[:idx]
			s = append(s, buf)
			if idx < len(tmp) {
				tmp = tmp[idx+1:]
			} else {
				break
			}
		} else {
			s = append(s, buf)
			tmp = tmp[idx+1:]
		}
	}
	return
}

func extractValue(arg string) (pkg, key, value string) {
	arg = strings.TrimSpace(arg)
	idx1 := strings.Index(arg, ".")
	if idx1 > 0 {
		pkg = arg[:idx1]
		tmp := arg[idx1+1:]
		idx2 := strings.Index(tmp, "=")
		if idx2 > 0 {
			key = tmp[:idx2]
			value = tmp[idx2+1:]
		} else {
			key = tmp
		}
	} else {
		pkg = arg
	}
	return
}

type extraCmdError string

func (e extraCmdError) Error() string {
	return "command not added to " + string(e)
}

type runError struct {
	cmd     string
	debug   string
	errType string
	err     error
}

func (e runError) Error() string {
	if Debug {
		if e.debug != "" {
			e.debug = "\n## DEBUG\n" + e.debug + "\n"
		}
		return fmt.Sprintf("Command line: `%s`\n%s\n## %s\n%s", e.cmd, e.debug, e.errType, e.err)
	}
	return fmt.Sprintf("\n%s", e.err)
}

// Run executes external commands just like RunWithMatch, but does not return
// the boolean `match`.
func Run(command string) (output []byte, err error) {
	output, _, err = RunWithMatch(command)
	return
}

// RunWithMatch executes external commands with access to shell features such as
// filename wildcards, shell pipes, environment variables, and expansion of the
// shortcut character "~" to home directory.
//
// This function avoids to have execute commands through a shell since an
// unsanitized input from an untrusted source makes a program vulnerable to
// shell injection, a serious security flaw which can result in arbitrary
// command execution.
//
// The most of commands return a text in output or an error if any.
// `match` is used in commands like *grep*, *find*, or *cmp* to indicate if the
// serach is matched.

func RunWithMatch(command string) (output []byte, match bool, err error) {
	var (
		cmds           []*exec.Cmd
		outPipes       []io.ReadCloser
		stdout, stderr bytes.Buffer
	)

	commands := strings.Split(command, "|")
	lastIdxCmd := len(commands) - 1

	// Check lonely pipes.
	for _, cmd := range commands {
		if strings.TrimSpace(cmd) == "" {
			err = runError{command, "", "ERR", errNoCmdInPipe}
			return
		}
	}

	for i, cmd := range commands {
		cmdEnv := _ENV // evironment variables for each command
		indexArgs := 1 // position where the arguments start
		fields := strings.Fields(cmd)
		lastIdxFields := len(fields) - 1
		for j, fCmd := range fields {
			if fCmd[len(fCmd)-1] == '=' || // VAR= foo
				(j < lastIdxFields && fields[j+1][0] == '=') { // VAR =foo
				err = runError{command, "", "ERR", errEnvVar}
				return
			}

			if strings.ContainsRune(fields[0], '=') {
				cmdEnv = append([]string{fields[0]}, _ENV...) // Insert the environment variable
				fields = fields[1:]                           // and it is removed from arguments
			} else {
				break
			}
		}
		// ==

		cmdPath, e := exec.LookPath(fields[0])
		if e != nil {
			err = runError{command, "", "ERR", e}
			return
		}

		// == Get the path of the next command, if any
		for j, fCmd := range fields {
			cmdBase := path.Base(fCmd)

			if cmdBase != "sudo" && cmdBase != "xargs" {
				break
			}
			// It should have an extra command.
			if j+1 == len(fields) {
				err = runError{command, "", "ERR", extraCmdError(cmdBase)}
				return
			}

			nextCmdPath, e := exec.LookPath(fields[j+1])
			if e != nil {
				err = runError{command, "", "ERR", e}
				return
			}

			if fields[j+1] != nextCmdPath {
				fields[j+1] = nextCmdPath
				indexArgs = j + 2
			}
			// == Expansion of arguments
		}
		expand := make(map[int][]string, len(fields))

		for j := indexArgs; j < len(fields); j++ {
			// Skip flags
			if fields[j][0] == '-' {
				continue
			}

			// Shortcut character "~"
			if fields[j] == "~" || strings.HasPrefix(fields[j], "~/") {
				fields[j] = strings.Replace(fields[j], "~", _HOME, 1)
			}

			// File name wildcards
			names, e := filepath.Glob(fields[j])
			if e != nil {
				err = runError{command, "", "ERR", e}
				return
			}
			if names != nil {
				expand[j] = names
			}
		}

		// Substitute the names generated for the pattern starting from last field.
		if len(expand) != 0 {
			for j := len(fields) - indexArgs; j >= indexArgs; j-- {
				if v, ok := expand[j]; ok {
					fields = append(fields[:j], append(v, fields[j+1:]...)...)
				}
			}
		}

		// == Handle arguments with quotes
		hasQuote := false
		needUpdate := false
		tmpFields := []string{}

		for j := indexArgs; j < len(fields); j++ {
			v := fields[j]
			lastChar := v[len(v)-1]

			if !hasQuote && (v[0] == '\'' || v[0] == '"') {
				if !needUpdate {
					needUpdate = true
				}

				v = v[1:] // skip quote

				if lastChar == '\'' || lastChar == '"' {
					v = v[:len(v)-1] // remove quote
				} else {
					hasQuote = true
				}

				tmpFields = append(tmpFields, v)
				continue
			}

			if hasQuote {
				if lastChar == '\'' || lastChar == '"' {
					v = v[:len(v)-1] // remove quote
					hasQuote = false
				}
				tmpFields[len(tmpFields)-1] += " " + v
				continue
			}

			tmpFields = append(tmpFields, v)
		}

		if needUpdate {
			fields = append(fields[:indexArgs], tmpFields...)
		}

		// == Create command
		c := &exec.Cmd{
			Path: cmdPath,
			Args: append([]string{fields[0]}, fields[1:]...),
			Env:  cmdEnv,
		}

		// == Connect pipes
		outPipe, e := c.StdoutPipe()
		if e != nil {
			err = runError{command, "", "ERR", e}
			return
		}

		if i == 0 {
			c.Stdin = os.Stdin
		} else {
			c.Stdin = outPipes[i-1] // anterior output
		}

		// == Buffers
		c.Stderr = &stderr

		// Only save the last output
		if i == lastIdxCmd {
			c.Stdout = &stdout
		}

		// == Start command
		if e := c.Start(); e != nil {
			err = runError{command,
				fmt.Sprintf("- Command: %s\n- Args: %s", c.Path, c.Args),
				"Start", fmt.Errorf("%s", c.Stderr)}
			return
		}

		//
		cmds = append(cmds, c)
		outPipes = append(outPipes, outPipe)
	}

	for _, c := range cmds {
		if e := c.Wait(); e != nil {
			_, isExitError := e.(*exec.ExitError)

			// Error type due I/O problems.
			if !isExitError {
				err = runError{command,
					fmt.Sprintf("- Command: %s\n- Args: %s", c.Path, c.Args),
					"Wait", fmt.Errorf("%s", c.Stderr)}
				return
			}

			if c.Stderr != nil {
				if stderr := fmt.Sprintf("%s", c.Stderr); stderr != "" {
					stderr = strings.TrimRight(stderr, "\n")
					err = runError{command,
						fmt.Sprintf("- Command: %s\n- Args: %s", c.Path, c.Args),
						"Stderr", fmt.Errorf("%s", stderr)}
					return
				}
			}
		} else {
			match = true
		}
	}

	return stdout.Bytes(), match, nil
}

// Runf is like Run, but formats its arguments according to the format.
// Analogous to Printf().
func Runf(format string, args ...interface{}) ([]byte, error) {
	return Run(fmt.Sprintf(format, args...))
}

// RunWithMatchf is like RunWithMatch, but formats its arguments according to
// the format. Analogous to Printf().
func RunWithMatchf(format string, args ...interface{}) ([]byte, bool, error) {
	return RunWithMatch(fmt.Sprintf(format, args...))
}
