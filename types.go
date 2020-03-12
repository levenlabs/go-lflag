package lflag

import (
	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

// All the built-int ParamTypes
const (
	ParamTypeString   = "string"
	ParamTypeInt      = "int"
	ParamTypeInt64    = "int64"
	ParamTypeBool     = "bool"
	ParamTypeDuration = "duration"
	ParamTypeJSON     = "json"
)

// ParseFunc is a function that takes a string from the Source and converts it
// into a value necessary to store in the sent pointer.
type ParseFunc func(string, interface{}) error

// JSONStringFunc takes a JSON value and converts it into an appropriate string
// that will be eventually sent to ParseFunc. JSONStringAsIs and
// JSONStringUnmarshal are both JSONStringFunc's.
type JSONStringFunc func(json.RawMessage) (string, error)

var paramTypeParsers = map[string]ParseFunc{
	ParamTypeString:   parseParamTypeString,
	ParamTypeInt:      parseParamTypeInt,
	ParamTypeInt64:    parseParamTypeInt64,
	ParamTypeBool:     parseParamTypeBool,
	ParamTypeDuration: parseParamTypeDuration,
	ParamTypeJSON:     parseParamTypeJSON,
}

func parseParamTypeString(val string, ptr interface{}) error {
	*(ptr.(*string)) = val
	return nil
}

func parseParamTypeInt(val string, ptr interface{}) error {
	vali, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	*(ptr.(*int)) = vali
	return nil
}

func parseParamTypeInt64(val string, ptr interface{}) error {
	vali, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}
	*(ptr.(*int64)) = vali
	return nil
}

func parseParamTypeBool(val string, ptr interface{}) error {
	*(ptr.(*bool)) = (val != "" && val != "false")
	return nil
}

func parseParamTypeDuration(val string, ptr interface{}) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	*(ptr.(*time.Duration)) = d
	return nil
}

func parseParamTypeJSON(val string, ptr interface{}) error {
	return json.Unmarshal([]byte(val), ptr)
}

// JSONStringAsIs is a JSONStringFunc that just casts the json.RawMessage as a
// string. It's used for non-quoted values in JSON like numbers and booleans.
func JSONStringAsIs(j json.RawMessage) (string, error) {
	return string(j), nil
}

// JSONStringUnmarshal is a JSONStringFunc that unmarshal's the json.RawMessage
// into a string. It's used for string values in JSON.
func JSONStringUnmarshal(j json.RawMessage) (string, error) {
	var str string
	err := json.Unmarshal(j, &str)
	if err != nil {
		return "", err
	}
	return str, nil
}

// functions which will take json marshaled value for a ParamType and convert it
// to a string which would be expected from a non-json source
//
// For example, a String param from the CLI would come in as the go string
// "foo", whereas its json form is "\"foo\"", so that would need to be decoded.
//
// On the other hand, an Int param from the CLI would come in as the go string
// "10", and its json form is also "10" (once cast from json.RawMessage ->
// string)
var paramTypeJSONStringers = map[string]JSONStringFunc{
	ParamTypeString:   JSONStringUnmarshal,
	ParamTypeInt:      JSONStringAsIs,
	ParamTypeBool:     JSONStringAsIs,
	ParamTypeDuration: JSONStringUnmarshal,
	ParamTypeJSON:     JSONStringAsIs,
}

var customParamTypeTypes = map[string]reflect.Type{}

// CustomParamType defines a new ParamType that can be used by calling Custom
// or RequiredCustom. The examplePtr should be an example value that your
// ParseFunc expects. For instance, if you expect *time.Time then you'd send a
// new(time.Time) as the examplePtr.
func CustomParamType(name string, examplePtr interface{}, p ParseFunc, jp JSONStringFunc) {
	if _, ok := paramTypeParsers[name]; ok {
		panic("lflag: paramType already defined: " + name)
	}
	paramTypeParsers[name] = p
	paramTypeJSONStringers[name] = jp
	customParamTypeTypes[name] = reflect.TypeOf(examplePtr).Elem()
}
