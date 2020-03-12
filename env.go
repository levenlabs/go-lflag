package lflag

import (
	"fmt"
	"os"
	"strings"
)

type sourceEnv struct{}

// NewSourceEnv initializes and returns a new Source which will pull from the
// environment variables at runtime. All param names are completely uppercased
// and have '-' replaced with '_', e.g "listen-addr" becomes "LISTEN_ADDR"
func NewSourceEnv() Source {
	return sourceEnv{}
}

// Parse implements the Source method
func (se sourceEnv) Parse(pp []Param) (map[string]string, error) {
	return parseEnv(os.Environ(), pp)
}

// split out for testing
func parseEnv(ee []string, pp []Param) (map[string]string, error) {
	envM := map[string]Param{}
	for _, p := range pp {
		name := strings.ToUpper(p.Name)
		name = strings.Replace(name, "-", "_", -1)
		envM[name] = p
	}

	ret := map[string]string{}
	for _, e := range ee {
		envParts := strings.SplitN(e, "=", 2)
		if len(envParts) != 2 {
			return nil, fmt.Errorf("malformed environment variable: %q", e)
		}
		if p, ok := envM[envParts[0]]; ok {
			ret[p.Name] = envParts[1]
		}
	}
	return ret, nil
}
