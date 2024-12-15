package go_Weather_ITUR

import "log"

type StationSystem struct {
	BasicSystem
	stations map[uint64]*StationEntity
}

func (s *StationSystem) GetEntity() []Identifier {
	var entities []Identifier
	for _, station := range s.stations {
		entities = append(entities, station)
	}
	return entities
}

func (s *StationSystem) Add(station *StationEntity) {
	if s.stations == nil {
		s.stations = make(map[uint64]*StationEntity)
	}
	s.stations[station.GetBasicEntity().id] = station
}

func (s *StationSystem) Remove(station *StationEntity) {
	if s.stations != nil {
		delete(s.stations, station.GetBasicEntity().id)
	}
}

func (s *StationSystem) Update(dt int64) {
	log.Printf("station of position is fixed")
}
