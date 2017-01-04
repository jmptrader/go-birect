package birect

// Info allows for associating data with any given connection
type Info map[string]interface{}

func newInfo() Info {
	return Info{}
}

// Get returns the value of the given key
func (i Info) Get(key string) interface{} {
	return i[key]
}

// Set sets the value of the given key
func (i Info) Set(key string, val interface{}) {
	i[key] = val
}

// GetString returns the value of the given key as a string.
// If the value of key is not a string, GetString will panic.
func (i Info) GetString(key string) string {
	if val := i.Get(key); val != nil {
		return val.(string)
	}
	return ""
}

// MustGetString returns the value of the given key as a string,
// or panics if there is no value.
func (i Info) MustGetString(key string) (val string) {
	if val = i.GetString(key); val == "" {
		panic("Missing " + key)
	}
	return
}
