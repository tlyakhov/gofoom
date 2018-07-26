package engine

import (
	"fmt"
)

// TextureCache is a simple in-memory key-value store of Textures.
type TextureCache struct {
	cache map[string]*Texture
}

// Get a cached texture or load it.
func (tc *TextureCache) Get(src string, genMips, filter bool) *Texture {
	key := fmt.Sprintf("%s_%t_%t", src, genMips, filter)

	if t, ok := tc.cache[key]; ok {
		return t
	}

	t := &Texture{Source: src, GenerateMipMaps: genMips, Filter: filter}
	t.Load()
	tc.cache[key] = t
	return t
}
