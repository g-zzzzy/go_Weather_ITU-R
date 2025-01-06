package go_Weather_ITUR

// import "log"

import (
	"log"
	"time"

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
	// w.componentManager.AddComponent(EntityID(sat.id), SatellitePositionComponent{satellite.Vector3{}})
	// w.componentManager.AddComponent(EntityID(sat.id), SatelliteVelocityComponent{satellite.Vector3{}})
	w.componentManager.AddComponent(EntityID(sat.id), SatelliteMovementComponent{})
}

func (s *SatelliteSystem) Update(dt int64, cm *ComponentManager, w *World) {
	s.AddElapsed(dt)
	if s.ShouldUpdate(s.elapsed) {
		for i := 0; i < len(cm.satelliteMovementComponents); i++ {
			// positionComponent := &cm.satellitePositionComponents[i]
			// velocityComponent := &cm.satelliteVelocityComponents[i]
			movementComponent := &cm.satelliteMovementComponents[i]

			if sat, exists := s.satellites[uint64(i)]; exists {
				// get current time
				currentTime := time.Now().In(time.FixedZone("CST", 8*60*60))
				year, month, day := currentTime.Date()
				hours, minutes, seconds := currentTime.Clock()
				p, v := satellite.Propagate(sat.satellite, year, int(month), day, hours, minutes, seconds)
				log.Printf("year: %d, month: %d, day: %d, hours: %d, minutes: %d, seconds: %d", year, month, day, hours, minutes, seconds)
				movementComponent.position.X = p.X
				movementComponent.position.Y = p.Y
				movementComponent.position.Z = p.Z
				movementComponent.velocity.X = v.X
				movementComponent.velocity.Y = v.Y
				movementComponent.velocity.Z = v.Z
				log.Printf("卫星 %d, X: %.2f, Y: %.2f, Z: %.2f", i, cm.satelliteMovementComponents[i].position.X, cm.satelliteMovementComponents[i].position.Y, cm.satelliteMovementComponents[i].position.Z)
			}
		}
	}
}

func (s *SatelliteSystem) GetEntityIDs() []uint64 {
	var ids []uint64
	for id := range s.satellites {
		ids = append(ids, id)
	}
	return ids
}
