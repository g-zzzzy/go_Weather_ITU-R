package go_Weather_ITUR

type System interface {
	Update(dt int64, cm *ComponentManager)
	// Remove(e BasicEntity)
	GetInterval() int64
	ShouldUpdate(elapsed int64) bool
	AddElapsed(dt int64)
}

type BasicSystem struct {
	interval int64
	elapsed  int64
}

func (s *BasicSystem) GetInterval() int64 {
	return s.interval
}

func (s *BasicSystem) AddElapsed(dt int64) {
	s.elapsed += dt
}

func (s *BasicSystem) ShouldUpdate(elapsed int64) bool {
	return s.elapsed%s.interval == 0
}
