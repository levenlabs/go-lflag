package lflag

// Param describes everything a Source needs to know about a single
// configuration option which has been defined
type Param struct {
	//  is the
	ParamType string

	// Name is the name of the parameter, e.g. "listen-addr" or "img-file".
	Name string

	// Default is the default value of the param, as a string. If this param is
	// going to be parsed as something else, like an int, the default string
	// should also be parsable as an int.
	//
	// If this param is a ParamTypeBool this will be "true" or "false"
	Default string

	// Usage is a short description of the paramater's usage
	Usage string

	// Required should be true if the parameter must be provided by the caller
	Required bool
}

// Source describes an entity which actually provides the values for
// configuration, e.g. command line or environment variables.
type Source interface {

	// Parse returns all the configured keys/values. The passed in slice should
	// not be modified.
	//
	// NOTE the only valid true value for a ParamTypeBool is "true". All other
	// values will be considered false
	Parse([]Param) (map[string]string, error)
}

// SourceStub can be used for testing with configuration options
type SourceStub map[string]string

// Parse implements the Source interface by returning the SourceStub as-is
func (ss SourceStub) Parse([]Param) (map[string]string, error) {
	return ss, nil
}

// Sources encompasses multiple Source instances. When Parse is called the
// returned map from each will be combined together, with right-most Source
// values taking precedence over their lefthand neighbor
type Sources []Source

// Parse implements the Source interface. See Sources' doc
func (ss Sources) Parse(pp []Param) (map[string]string, error) {
	vals := map[string]string{}
	for _, s := range ss {
		sm, err := s.Parse(pp)
		if err != nil {
			return nil, err
		}
		for k, v := range sm {
			vals[k] = v
		}
	}
	return vals, nil
}
