package material

type Painful struct {
	Hurt float64 `editable:"Hurt" edit_type:"float"`
}

func (m *Painful) ActOnEntity(e Entity) {
	if m.Hurt == 0 {
		return
	}

	if ae, ok := e.(*AliveEntity); ok && ae.HurtTime == 0 {
		ae.Hurt(m.Hurt)
	}
}

type PainfulSampled struct {
	Sampled
	Painful
}
