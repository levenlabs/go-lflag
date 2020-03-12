package lflag

import (
	"bytes"
	. "testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceJSON(t *T) {
	pp := []Param{
		{ParamType: ParamTypeString, Name: "str"},
		{ParamType: ParamTypeString, Name: "str2"},
		{ParamType: ParamTypeInt, Name: "int"},
		{ParamType: ParamTypeBool, Name: "bool"},
		{ParamType: ParamTypeDuration, Name: "dur"},
		{ParamType: ParamTypeJSON, Name: "json"},
	}

	ts := SourceStub{
		"str2": "bar", // this should overwrite the one in the jsonFile
	}

	jsonFile := bytes.NewBufferString(`{
		"str": "foo\nsomething",
		"str2": "broken",
		"int":  1,
		"bool": true,
		"dur":  "30s",
		"json": {"foo":"bar"}
	}`)

	m, err := sourceJSON{innerSrc: ts, testJSONFile: jsonFile}.Parse(pp)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"str":  "foo\nsomething",
		"str2": "bar",
		"int":  "1",
		"bool": "true",
		"dur":  "30s",
		"json": `{"foo":"bar"}`,
	}, m)
}
