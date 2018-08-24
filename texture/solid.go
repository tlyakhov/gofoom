package texture

import (
	"fmt"
	"image/color"

	// Decoders for common image types
	_ "image/jpeg"
	_ "image/png"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/registry"
)

type Solid struct {
	concepts.Base
	Diffuse color.NRGBA
}

func init() {
	registry.Instance().Register(Solid{})
}

func (s *Solid) Initialize() {
	s.Base = concepts.Base{}
	s.Base.Initialize()
}

func (s *Solid) Sample(x, y float64, scale float64) uint32 {
	return concepts.NRGBAToInt32(s.Diffuse)
}

func (s *Solid) Deserialize(data map[string]interface{}) {
	s.Initialize()
	s.Base.Deserialize(data)
	if v, ok := data["Diffuse"]; ok {
		var err error
		s.Diffuse, err = concepts.ParseHexColor(v.(string))
		if err != nil {
			fmt.Printf("Error deserializing Solid texture: %v\n", err)
		}
	}
}

func (s *Solid) Serialize() map[string]interface{} {
	result := s.Base.Serialize()
	result["Type"] = "texture.Solid"
	if s.Diffuse.A == 255 {
		result["Diffuse"] = fmt.Sprintf("#%02X%02X%02X", s.Diffuse.R, s.Diffuse.G, s.Diffuse.B)
	} else {
		result["Diffuse"] = fmt.Sprintf("#%02X%02X%02X%02X", s.Diffuse.R, s.Diffuse.G, s.Diffuse.B, s.Diffuse.A)
	}
	return result
}
