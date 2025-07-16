package go_Weather_ITUR

// import "log"

import (
	"log"
	"math"
	"time"

	"github.com/joshuaferrara/go-satellite"
)

type SatelliteSystem struct {
	BasicSystem
}

func NewSatelliteSystem(interval int64) *SatelliteSystem {
	return &SatelliteSystem{
		BasicSystem{
			name:     "SatelliteSystem",
			interval: interval,
		},
	}
}

func (s *SatelliteSystem) Update(dt int64, cm *ComponentManager, w *World) {
	log.Printf("SatelliteSystem update...")
	startTime := time.Now()
	count := 0
	for i := range cm.SatelliteMovementComponents {
		count++
		movementComponent := &cm.SatelliteMovementComponents[i]
		entityID := movementComponent.EntityID
		sat, exists := cm.SatelliteSGP4Components[entityID]
		if !exists {
			continue
		}

		currentTime := time.Now().UTC()
		year, month, day := currentTime.Date()
		hour, min, sec := currentTime.Clock()
		p, v := satellite.Propagate(sat.Satrec, year, int(month), day, hour, min, sec)

		gmst := satellite.GSTimeFromDate(year, int(month), day, hour, min, sec)
		alt, _, lla := satellite.ECIToLLA(p, gmst)

		latitudeDeg := lla.Latitude * 180 / math.Pi
		longitudeDeg := lla.Longitude * 180 / math.Pi
		longitudeDeg = math.Mod(longitudeDeg+180+360, 360)
		altitudeMeters := alt * 1000

		movementComponent.PosX, movementComponent.PosY, movementComponent.PosZ = latitudeDeg, longitudeDeg, altitudeMeters
		movementComponent.VelX, movementComponent.VelY, movementComponent.VelZ = v.X, v.Y, v.Z
	}
	endTime := time.Now()
	log.Printf("satellite count: %d", count)
	log.Printf("SatelliteSystem update time: %v", endTime.Sub(startTime))

}
