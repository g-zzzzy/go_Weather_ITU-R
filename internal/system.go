package go_Weather_ITUR

type System interface {
	Update(dt int64, cm *ComponentManager, w *World)
	// Remove(e BasicEntity)
	GetInterval() int64
	ShouldUpdate(elapsed int64) bool
	AddElapsed(dt int64)
	GetElapsed() int64
	GetEntityIDs() []EntityID
	Name() string
}

type BasicSystem struct {
	name      string
	interval  int64
	elapsed   int64
	entityIDs []EntityID
}

func (s *BasicSystem) AddEntityID(eid EntityID) {
	s.entityIDs = append(s.entityIDs, eid)
}

func (s *BasicSystem) GetElapsed() int64 {
	return s.elapsed
}

func (s *BasicSystem) GetEntityIDs() []EntityID {
	return s.entityIDs
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

func (s *BasicSystem) Name() string {
	return s.name
}
