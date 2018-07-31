package mapping

var (
	ValidSectorTypes = map[string]interface{}{
		"mapping.Sector":      Sector{},
		"mapping.ToxicSector": ToxicSector{},
	}
	ValidEntityTypes = map[string]interface{}{
		"mapping.Player": Player{},
	}
)
