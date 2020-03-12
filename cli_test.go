package lflag

import (
	. "testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI(t *T) {
	found, err := parseCLI([]string{
		"--foo", "bats", "--bar=butts", "--flag1",
		"--flag2", "false",
		"something",         // should be ignored
		"something=else",    // so should this
		"something", "else", // and these
		"--unk=wat", // and whatever this is
	}, testParams)
	require.Nil(t, err)
	assert.Equal(t,
		map[string]string{
			"foo":   "bats",
			"bar":   "butts",
			"flag1": "true",
			"flag2": "false",
		},
		found,
	)
}
