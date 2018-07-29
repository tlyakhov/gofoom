package logic

func (m *Map) Frame(lastFrameTime float64) {
	m.Player.Frame(lastFrameTime)

	for _, sector := range m.Sectors {
		// sector.ActOnEntity(player)
		for _, e := range sector.Entities {
			if !e.Active {
				continue
			}
			for _, pvs := range sector.PVSEntity {
				pvs.ActOnEntity(e)
			}
		}
		sector.Frame(lastFrameTime)
	}
}
