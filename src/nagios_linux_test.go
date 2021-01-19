// +build linux
package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_runCommand_InvalidCommandError(t *testing.T) {
	stdout, stderr, exit := runCommand("jdijfs")
	assert.Equal(t, -1, exit)
	assert.Equal(t, "", stdout)
	assert.NotEmpty(t, stderr)
}

func Test_runCommand_returns1(t *testing.T) {
	stdout, stderr, exit := runCommand("/bin/sh", "test/returns2.sh")
	assert.Equal(t, 2, exit)
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
}
