package lflag

import (
	. "testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceEnv(t *T) {
	env := []string{
		"FOO=",
		"BAR=bar",
		"FLAG1=true",
		"HOME=whatever",
		"FOO_BAR=okthen",
	}

	out, err := parseEnv(env, testParams)
	require.Nil(t, err)
	assert.Equal(t, map[string]string{
		"foo":     "",
		"bar":     "bar",
		"flag1":   "true",
		"foo-bar": "okthen",
	}, out)

}
