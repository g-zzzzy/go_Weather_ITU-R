package go_Weather_ITUR

// import "log"

import (
	"log"

	"github.com/joshuaferrara/go-satellite"
)

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

func (s *SatelliteSystem) Add(sat *SatelliteEntity) {
	if s.satellites == nil {
		s.satellites = make(map[uint64]*SatelliteEntity)
	}
	line1 := sat.TLE.line1
	line2 := sat.TLE.line2
	grav := sat.TLE.gravConst
	satrec := satellite.TLEToSat(line1, line2, grav)
	sat.satellite = satrec
	s.satellites[sat.GetBasicEntity().id] = sat
	log.Printf("卫星：%d,初始时的位置是 %f, %f, %f", sat.GetBasicEntity().id, sat.position.position.X, sat.position.position.Y, sat.position.position.Z)
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
		for _, sat := range s.satellites {
			p, v := satellite.Propagate(sat.satellite, 2023, 12, 30, 1+int(s.elapsed), 14, int(dt))
			log.Printf("time: %d", 13+s.elapsed)
			log.Printf("P: X = %.2f, Y = %.2f, Z = %.2f\n", p.X, p.Y, p.Z)
			log.Printf("V: X = %.2f, Y = %.2f, Z = %.2f\n", v.X, v.Y, v.Z)
			// log.Printf("卫星：%d,在2023年12月30日%d时14分的位置是 %f, %f, %f", sat.GetBasicEntity().id, s.elapsed+13, sat.position.position.X, sat.position.position.Y, sat.position.position.Z)
		}
	}
}
