package lflag

import (
	"bytes"
	"fmt"
)

// HelpPrefix may be set as a prefix to be printed out by sources which can
// print out help strings (e.g. NewSourceCLI). If the string doesn't end in a
// newline one will be added
var HelpPrefix string

// Build variables which can be set during the go build command. E.g.:
// -ldflags "-X 'gerrit.levenlabs.com/llib/lflag.BuildCommit commitHash'"
// These fields will be used to construct the string printed out when the
// --version flag is used.
var (
	BuildCommit    string
	BuildDate      string
	BuildNumber    string // TeamCity (or other build system) build number
	BuildGoVersion string
)

// Version compiles the build strings into a string which will be printed out
// when --version or -V is used, but is exposed so it may be used other places
// too.
func Version() string {
	orStr := func(s, alt string) string {
		if s == "" {
			return alt
		}
		return s
	}
	b := new(bytes.Buffer)
	fmt.Fprintf(b, "BuildCommit: %s\n", orStr(BuildCommit, "<unset>"))
	fmt.Fprintf(b, "BuildDate: %s\n", orStr(BuildDate, "<unset>"))
	fmt.Fprintf(b, "BuildNumber: %s\n", orStr(BuildNumber, "<unset>"))
	fmt.Fprintf(b, "BuildGoVersion: %s\n", orStr(BuildGoVersion, "<unset>"))
	return b.String()
}
