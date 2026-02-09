package dynamic

type Timer struct {
	Start    int64
	End      int64
	OnExpiry func()
}

func (t *Timer) Attach(sim *Simulation) {
	sim.Timers[t] = struct{}{}
}

func (t *Timer) Detach(sim *Simulation) {
	delete(sim.Timers, t)
}

func (t *Timer) Expired() bool {
	return false
}
