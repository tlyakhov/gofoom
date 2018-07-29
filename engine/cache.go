package engine

var cache map[string]*interface{}

// Cache gets a cached model or creates it.
func Cache(id string, constructor func(id string) *interface{}) *interface{} {
	if cache == nil {
		cache = make(map[string]*interface{})
	}
	if model, ok := cache[id]; ok {
		return model
	}
	model := constructor(id)
	cache[id] = model
	return model
}
