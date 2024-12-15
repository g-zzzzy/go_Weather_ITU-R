package go_Weather_ITUR


type World struct {
	systems      []System
}

func (w *World) AddSystem(system System) {
	w.systems = append(w.systems, system)
}

func (w *World) GetSystem() []System {
	return w.systems
}

func (w *World) Update(dt int64) {
	for _, system := range w.GetSystem() {
		system.Update(dt)
	}
}
