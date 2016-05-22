package rogi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var cmdTest = &Command{
	Run:       runTest,
	UsageLine: "test: usage line",
	Short:     "short test",
	Long:      "long test",
}

func runTest(cmd *Command, args []string) {
}

func TestCommandName(t *testing.T) {
	assert.Equal(t, cmdTest.Name(), "test")
}

func TestCommandRunable(t *testing.T) {
	assert.Equal(t, cmdTest.Runable(), true)
}

func TestCommandUsage(t *testing.T) {
	assert.Equal(t, cmdTest.UsageLine, "test: usage line")
}

func TestCommandShort(t *testing.T) {
	assert.Equal(t, cmdTest.Short, "short test")
}

func TestCommandLong(t *testing.T) {
	assert.Equal(t, cmdTest.Long, "long test")
}
