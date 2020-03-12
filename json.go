package lflag

import (
	"encoding/json"
	"io"
	"os"
	"strings"
)

type sourceJSON struct {
	innerSrc     Source
	testJSONFile io.Reader // used as a fake json file in tests
}

// NewSourceJSON wraps an existing Source and adds support for reading a json
// file to source parameter values. The values coming from the inner Source will
// overwrite any which are found in the json file
func NewSourceJSON(inner Source) Source {
	return sourceJSON{innerSrc: inner}
}

func (sj sourceJSON) Parse(pp []Param) (map[string]string, error) {
	const paramName = "config-json-file"
	pp = append(pp, Param{
		ParamType: ParamTypeString,
		Name:      paramName,
		Usage:     "Name of json file to parse config object out of. Environment and CLI params overwrite json ones",
	})

	m, err := sj.innerSrc.Parse(pp)
	if err != nil {
		return nil, err
	}

	// if the parsed m contains a config file set, or a test one is given, make
	// a json decoder out of that. otherwise return the m we have
	var dec *json.Decoder
	if jsonConfigFile := m[paramName]; jsonConfigFile != "" {
		f, err := os.Open(jsonConfigFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		dec = json.NewDecoder(f)
	} else if sj.testJSONFile != nil {
		dec = json.NewDecoder(sj.testJSONFile)
	} else {
		return m, nil
	}

	// parse into a json map
	var jm map[string]json.RawMessage
	if err := dec.Decode(&jm); err != nil {
		return nil, err
	}

	// now transform the map[string]json.RawMessage into a map[string]string
	// using the json stringers
	out := make(map[string]string, len(jm))
	for _, p := range pp {
		// we treat null and unset as the same thing
		j, ok := jm[strings.ToLower(p.Name)]
		if !ok {
			continue
		}
		if string(j) == "null" {
			continue
		}

		str, err := paramTypeJSONStringers[p.ParamType](j)
		if err != nil {
			return nil, err
		}
		out[p.Name] = str
	}

	// merge m into out (so the inner source values overwrite this ones') and
	// return that
	for k, v := range m {
		out[k] = v
	}

	return out, nil
}
