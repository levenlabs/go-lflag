package lflag

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
)

type sourceCLI struct{}

// NewSourceCLI initializes  and returns a new Source which will pull from the
// command line arguments at runtime. It also handles --help and --version
// options
func NewSourceCLI() Source {
	return sourceCLI{}
}

func (sc sourceCLI) Parse(pp []Param) (map[string]string, error) {
	return parseCLI(os.Args[1:], pp)
}

// split out for testing
func parseCLI(args []string, pp []Param) (map[string]string, error) {
	cliM := map[string]Param{}
	for _, p := range pp {
		cliM["--"+p.Name] = p
	}

	var arg string
	found := map[string]string{}
	for {
		if len(args) == 0 {
			return found, nil
		}

		arg, args = args[0], args[1:]

		argParts := strings.SplitN(arg, "=", 2)
		argName := argParts[0]

		if argName == "-h" || argName == "--help" {
			printfAndExit(cliHelpStr(pp))
		} else if argName == "-V" || argName == "--version" {
			printfAndExit(Version())
		}

		var argVal string
		var argValOk bool
		if len(argParts) == 2 {
			argVal = argParts[1]
			argValOk = true
		}

		p, ok := cliM[argName]
		if !ok {
			continue
		}

		if p.ParamType == ParamTypeBool {
			// check for a true/false value
			if !argValOk && len(args) > 0 {
				if !strings.HasPrefix(args[0], "-") {
					argValOk = true
					argVal, args = args[0], args[1:]
				}
			}
			if argValOk {
				found[p.Name] = argVal
			} else if p.Default == "true" {
				found[p.Name] = ""
			} else {
				found[p.Name] = "true"
			}
			continue
		}

		if !argValOk && len(args) > 0 {
			argVal, args = args[0], args[1:]
		}

		found[p.Name] = argVal
	}
}

// returns string form of help message. newline will be appended already
func cliHelpStr(pp []Param) string {

	sort.Slice(pp, func(i, j int) bool {
		return pp[i].Name < pp[j].Name
	})

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	bufParam := func(p Param) {
		fmt.Fprintf(buf, "\t--%s", p.Name)
		if p.ParamType == ParamTypeBool {
			fmt.Fprintf(buf, " (flag)")
		}
		fmt.Fprintf(buf, "\n")

		if p.Usage != "" {
			fmt.Fprintf(buf, "\t\t%s\n", p.Usage)
		}

		if p.Default != "" {
			fmt.Fprintf(buf, "\t\tDefault: %q\n", p.Default)
		} else if p.Required {
			fmt.Fprintf(buf, "\t\t(Required)\n")
		} else {
			fmt.Fprintf(buf, "\t\t(Optional)\n")
		}

		// All parameters span multiple lines, so add a newline between each for
		// clarity
		fmt.Fprintf(buf, "\n")
	}

	if HelpPrefix != "" {
		fmt.Fprintf(buf, "\n%s", HelpPrefix)
		if HelpPrefix[len(HelpPrefix)-1] != '\n' {
			// ensure we always write at least one newline
			fmt.Fprint(buf, "\n")
		}
	}

	fmt.Fprint(buf, "\n")
	for _, p := range pp {
		bufParam(p)
	}
	bufParam(Param{
		ParamType: ParamTypeBool,
		Name:      "help",
		Usage:     "Show this help message and exit",
	})
	bufParam(Param{
		ParamType: ParamTypeBool,
		Name:      "version",
		Usage:     "Print out a build string and exit",
	})

	return buf.String()
}
