package go_Weather_ITUR

import "log"

//import "github.com/joshuaferrara/go-satellite"

type SatelliteSystem struct {
	BasicSystem
	satellites map[uint64]*SatelliteEntity
}

func (s *SatelliteSystem) GetEntity() []Identifier {
	var entities []Identifier
	for _, satellite := range s.satellites {
		entities = append(entities, satellite)
	}
	return entities
}

func (s *SatelliteSystem) Add(satellite *SatelliteEntity) {
	if s.satellites == nil {
		s.satellites = make(map[uint64]*SatelliteEntity)
	}
	s.satellites[satellite.GetBasicEntity().id] = satellite
}

func (s *SatelliteSystem) Remove(satellite *SatelliteEntity) {
	if s.satellites != nil {
		delete(s.satellites, satellite.GetBasicEntity().id)
	}
}

func (s *SatelliteSystem) Update(dt int64) {
	s.AddElapsed(dt)
	if s.ShouldUpdate(s.elapsed) {
		println("satellite position update")
		// 具体的update
		for _, satellite := range s.satellites {
			satellite.position.lat += 10
			satellite.position.lon += 10
			satellite.position.h += 1
			log.Printf("卫星 %d 的位置更新为：纬度 %.2f，经度 %.2f，高度 %.2f\n", satellite.GetBasicEntity().id, satellite.position.lat, satellite.position.lon, satellite.position.h)
		}
	}
}
