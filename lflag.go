// Package lflag handles defining configuration providers as well as
// coordinating initialization code.
//
// # Configuration
//
// Parameters are defined much like they are in the standard flag package. A
// pointer is returned from a method like String, which will then be filled once
// lflag.Parse is called. All parameters must be defined before lflag.Parse is
// called.
//
//	addr := lflag.String("db-addr", ":666", "Address the database is listening on")
//	poolSize := lflag.Int("db-pool-size", 10, "Number of connections to the database to make")
//	lflag.Parse(lflag.NewSourceCLI())
//	db, err := NewDB(*addr, *poolSize)
//
// # Sources
//
// lflag supports multiple Sources, which are places from which configuration
// information is obtained. For example, NewSourceCLI returns a Source which
// will read configuration from command-line parameters, NewSourceEnv read
// configuration from environment parameters, and so on. Sources can be combined
// together using the Sources type, which handles merging the different Sources
// into a single one.
//
// # Initialization
//
// In addition to handling configuration parameters lflag also manages
// initialization based on their values, via the Do function. When lflag.Parse
// is called, after all parameters have been filled in, all functions passed
// into Do are called, in the order in which they were passed in.
//
// # Miscellaneous
//
// lflag handles a couple of other behaviors which are related to initialization
// and runtime.
//
// Build information can be compiled into binaries using lflag using the build
// variables like BuildCommit. See their doc string for how exactly to do that.
package lflag

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	llog "github.com/levenlabs/go-llog"
)

var (
	m = map[string]param{}
	l sync.Mutex

	// queueCh is used to queue up future Do's
	queueCh = make(chan func())

	// doneCh is closed once spin is done and all queued up Do's have been called
	// this is used to signal to future Do's that they shouldn't queue
	doneCh = make(chan bool)

	// callCh is used by Parse to get all queued up Do's
	callCh = make(chan func())

	// spinWg is only needed to keep track of spin for Reset to know when its
	// safe to reset
	spinWg sync.WaitGroup
)

func init() {
	spinWg.Add(1)
	go spin()

	logLevel := String("log-level", "info", "Log level to run with. Available levels are: debug, info, warn, error, fatal")
	Do(func() {
		err := llog.SetLevelFromString(*logLevel)
		if err != nil {
			llog.Fatal("error setting log level", llog.ErrKV(err))
		}
	})
}

func spin() {
	defer spinWg.Done()
	dos := []func(){
		// seed this with an empty one so the for loop doesn't immediately break
		func() {},
	}
loop:
	for len(dos) > 0 {
		select {
		case fn, ok := <-queueCh:
			// only Reset closes queueCh and so we should bail immediately
			if !ok {
				break loop
			}
			dos = append(dos, fn)
		case callCh <- dos[0]:
			dos = dos[1:]
		}
	}
	// trigger all waiting/future Do calls to immediately happen
	close(doneCh)
	// tell Parse to stop looping
	close(callCh)
}

func printfAndExit(str string, args ...interface{}) {
	fmt.Printf(str, args...)
	os.Stdout.Sync()
	os.Stderr.Sync()
	os.Exit(0)
}

// Prefixed is basically strings.Join except it ignores empty strings
func Prefixed(strs ...string) string {
	var nonemptys []string
	for _, s := range strs {
		if s == "" {
			continue
		}
		nonemptys = append(nonemptys, s)
	}
	return strings.Join(nonemptys, "-")
}

type param struct {
	Param
	ptr interface{}
}

func newParam(p Param, ptr interface{}) interface{} {
	l.Lock()
	defer l.Unlock()

	pp, ok := m[p.Name]
	if ok && pp.Param != p {
		panic(fmt.Sprintf("param named %q already exists and differs from this new one", p.Name))
	} else if !ok {
		pp.Param = p
		pp.ptr = ptr
	}
	m[p.Name] = pp
	return pp.ptr
}

// Do registers the given function to be performed after Parse is called an all
// param pointers have been filled in. Multiple functions may be registered
// using Do, though the order they are called is guaranteed to be in calling
// order. Once Parse has been called and ALL queued Do functions have returned,
// sent function is immediately invoked. If Do is called within another Do, then
// the sent function is added to the queue at the end and will be envoked once all
// queued up Do's are called.
func Do(fn func()) {
	select {
	// if doneCh is closed then we immediately call, otherwise we add it to the queue
	case <-doneCh:
		fn()
	case queueCh <- fn:
	}
}

// Reset removes any configured flags and resets everything back to before
// Parse was called. This should ONLY be used in tests
func Reset() {
	// closing queueCh will immediately kill spin
	close(queueCh)
	spinWg.Wait()
	m = map[string]param{}
	queueCh = make(chan func())
	doneCh = make(chan bool)
	callCh = make(chan func())

	spinWg.Add(1)
	go spin()
}

// String takes in the name of a config param, a default value, and a string
// describing the usage for the param, and returns a pointer which will be
// filled when Parse is called
func String(name, value, usage string) *string {
	p := Param{
		ParamType: ParamTypeString,
		Name:      name,
		Default:   value,
		Usage:     usage,
	}
	ptr := new(string)
	return newParam(p, ptr).(*string)
}

// RequiredString is like String, but it has no default value and must e set
func RequiredString(name, usage string) *string {
	p := Param{
		ParamType: ParamTypeString,
		Name:      name,
		Usage:     usage,
		Required:  true,
	}
	ptr := new(string)
	return newParam(p, ptr).(*string)
}

// Int takes in the name of a config param, a default value, and a string
// describing the usage for the param, and returns a pointer which will be
// filled when Parse is called
func Int(name string, value int, usage string) *int {
	p := Param{
		ParamType: ParamTypeInt,
		Name:      name,
		Default:   strconv.Itoa(value),
		Usage:     usage,
	}
	ptr := new(int)
	return newParam(p, ptr).(*int)
}

// RequiredInt is like Int, but it has no default value and must e set
func RequiredInt(name, usage string) *int {
	p := Param{
		ParamType: ParamTypeInt,
		Name:      name,
		Usage:     usage,
		Required:  true,
	}
	ptr := new(int)
	return newParam(p, ptr).(*int)
}

// Bool takes in the name of a config param, a default value, and a string
// describing the usage for the param, and returns a pointer which will be
// filled when Parse is called
func Bool(name string, value bool, usage string) *bool {
	var def string
	if value {
		def = "true"
	}
	p := Param{
		ParamType: ParamTypeBool,
		Name:      name,
		Default:   def,
		Usage:     usage,
	}
	ptr := new(bool)
	return newParam(p, ptr).(*bool)
}

// RequiredBool is like Bool, but it has no default value and must be set
func RequiredBool(name, usage string) *bool {
	p := Param{
		ParamType: ParamTypeBool,
		Name:      name,
		Usage:     usage,
		Required:  true,
	}
	ptr := new(bool)
	return newParam(p, ptr).(*bool)
}

// Duration takes in the name of a config param, a default value, a string
// describing the usage for the param, and returns a pointer which will be
// filled when Parse is called.
//
// The value given by a config for a Duration must be parsable by
// time.ParseDuration.
func Duration(name string, value time.Duration, usage string) *time.Duration {
	p := Param{
		ParamType: ParamTypeDuration,
		Name:      name,
		Default:   value.String(),
		Usage:     usage,
	}
	ptr := new(time.Duration)
	return newParam(p, ptr).(*time.Duration)
}

// RequiredDuration is like Duration, but it has no default value and must be
// set
func RequiredDuration(name, usage string) *time.Duration {
	p := Param{
		ParamType: ParamTypeDuration,
		Name:      name,
		Usage:     usage,
		Required:  true,
	}
	ptr := new(time.Duration)
	return newParam(p, ptr).(*time.Duration)
}

// Custom takes in a paramType, the name of a config param, a default value, a
// string describing the usage for the param, and returns a pointer which will
// be filled when Parse is called.
//
// The value must be the underlying type that your custom Parser expects as a
// pointer and we'll use fmt.Sprint() to convert it to a string.
//
// You can use type-assertion on the return value to get the pointer type that
// you expect. For instance, if your custom ParamType is for time.Time, this
// would return a *time.Time.
func Custom(paramType, name string, value interface{}, usage string) interface{} {
	p := Param{
		ParamType: paramType,
		Name:      name,
		Default:   fmt.Sprint(value),
		Usage:     usage,
	}
	typ, ok := customParamTypeTypes[paramType]
	if !ok {
		panic("lflag: ParamType not defined: " + paramType)
	}
	ptr := reflect.New(typ).Interface()
	return newParam(p, ptr)
}

// RequiredCustom is like Custom, but it has no default and must be set
func RequiredCustom(paramType, name, usage string) interface{} {
	p := Param{
		ParamType: paramType,
		Name:      name,
		Usage:     usage,
		Required:  true,
	}
	typ, ok := customParamTypeTypes[paramType]
	if !ok {
		panic("lflag: ParamType not defined: " + paramType)
	}
	ptr := reflect.New(typ).Interface()
	return newParam(p, ptr)
}

// JSON reads in the config param as a string and json.Unmarshals it into the
// given receiver pointer. value is a default value for the parameter, and usage
// describes the parameter.
//
// If value cannot be json.Marshaled (for help string purposes) this will panic
func JSON(rcv interface{}, name string, value interface{}, usage string) {
	jValue, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	p := Param{
		ParamType: ParamTypeJSON,
		Name:      name,
		Default:   string(jValue),
		Usage:     usage,
	}
	ptr := newParam(p, rcv)
	if ptr != rcv {
		panic(fmt.Sprintf("param named %q already exists and differs from this new one", p.Name))
	}
}

// RequiredJSON is like JSON, but it has no default and must be set
func RequiredJSON(rcv interface{}, name string, usage string) {
	p := Param{
		ParamType: ParamTypeJSON,
		Name:      name,
		Usage:     usage,
		Required:  true,
	}
	newParam(p, rcv)
}

// Parse goes through each Source and compiles a set of values for registered
// params (later Sources overwrite previous ones), fills in param pointers, then
// calls all functions registered using Do.
//
// At the end of Parse all of the lflag package's params are reset. Any calls
// to Do will immediately invoke the sent function after Parse is called.
func Parse(s Source) {
	l.Lock()
	defer l.Unlock()

	pp := make([]Param, 0, len(m))
	for _, p := range m {
		pp = append(pp, p.Param)
	}

	vals, err := s.Parse(pp)
	if err != nil {
		panic(err)
	}

	for _, p := range pp {
		val, valOk := vals[p.Name]
		if !valOk {
			if p.Required {
				panic(fmt.Sprintf("parameter %q required but not set", p.Name))
			}
			val = p.Default
		}

		err := paramTypeParsers[p.ParamType](val, m[p.Name].ptr)
		if err != nil {
			panic(fmt.Sprintf("error parsing parameter %s: %v", p.Name, err))
		}
	}

	for fn := range callCh {
		fn()
	}

	m = map[string]param{}
}

// Configure is a shortcut around Parse which uses our default sources (in order
// of least-to-most precedent: json-file, environment, cli).
//
// Additionally, lflag.NotifyReadyAfter is called  and closed after Parse is
// called letting any service manager know that we are now ready.
func Configure() {
	var s Source = Sources{NewSourceEnv(), NewSourceCLI()}
	s = NewSourceJSON(s)
	Parse(s)
}
