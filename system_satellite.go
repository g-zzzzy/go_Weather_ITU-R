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

func (s *SatelliteSystem) Add(sat *SatelliteEntity, w *World) {
	if s.satellites == nil {
		s.satellites = make(map[uint64]*SatelliteEntity)
	}
	line1 := sat.TLE.line1
	line2 := sat.TLE.line2
	grav := sat.TLE.gravConst
	satrec := satellite.TLEToSat(line1, line2, grav)
	sat.satellite = satrec
	s.satellites[sat.GetBasicEntity().id] = sat
	log.Printf("Adding component for satellite with ID %d\n", sat.id)
	w.componentManager.AddComponent(EntityID(sat.id), SatellitePositionComponent{satellite.Vector3{}})
	w.componentManager.AddComponent(EntityID(sat.id), SatelliteVelocityComponent{satellite.Vector3{}})
}

func (s *SatelliteSystem) Update(dt int64, cm *ComponentManager) {
	s.AddElapsed(dt)
	if s.ShouldUpdate(s.elapsed) {
		for i := 0; i < len(cm.satellitePositionComponents); i++ {
			positionComponent := &cm.satellitePositionComponents[i]
			velocityComponent := &cm.satelliteVelocityComponents[i]

			if sat, exists := s.satellites[uint64(i)]; exists {
				p, v := satellite.Propagate(sat.satellite, 2023, 12, 30, 1+int(s.elapsed), 14, int(dt))
				positionComponent.position.X = p.X
				positionComponent.position.Y = p.Y
				positionComponent.position.Z = p.Z
				velocityComponent.velocity.X = v.X
				velocityComponent.velocity.Y = v.Y
				velocityComponent.velocity.Z = v.Z
				log.Printf("卫星 %d, X: %.2f, Y: %.2f, Z: %.2f", i, cm.satellitePositionComponents[i].position.X, cm.satellitePositionComponents[i].position.Y, cm.satellitePositionComponents[i].position.Z)
			}
		}
	}
}
