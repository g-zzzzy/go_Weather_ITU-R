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
	for entityID, movementComponent := range cm.SatelliteMovementComponents {
		sat, exists := cm.SatelliteSGP4Components[entityID]
		if !exists {
			continue
		}

		currentTime := time.Now().UTC()
		year, month, day := currentTime.Date()
		hour, min, sec := currentTime.Clock()
		p, v := satellite.Propagate(sat.Satrec, year, int(month), day, hour, min, sec)
		// log.Printf("Entity %d - Propagate at %d-%02d-%02d %02d:%02d:%02d", entityID, year, month, day, hour, min, sec)

		// 转换到地理坐标
		gmst := satellite.GSTimeFromDate(year, int(month),
			day, hour, min, sec)
		alt, _, lla := satellite.ECIToLLA(p, gmst)

		// 单位转换
		latitudeDeg := lla.Latitude * 180 / math.Pi
		longitudeDeg := lla.Longitude * 180 / math.Pi

		longitudeDeg = math.Mod(longitudeDeg+180+360, 360)

		altitudeMeters := alt * 1000

		movementComponent.Position.X, movementComponent.Position.Y, movementComponent.Position.Z = latitudeDeg, longitudeDeg, altitudeMeters
		movementComponent.Velocity.X, movementComponent.Velocity.Y, movementComponent.Velocity.Z = v.X, v.Y, v.Z
		cm.SatelliteMovementComponents[entityID] = movementComponent

		// fmt.Printf("[SatelliteSystem] Entity %d: Lat=%.2f Lon=%.2f Alt=%.2f\n",
		// 	entityID, latitudeDeg, longitudeDeg, altitudeMeters)
	}

}
