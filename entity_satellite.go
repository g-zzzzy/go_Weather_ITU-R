package go_Weather_ITUR

import "github.com/joshuaferrara/go-satellite"

type SatelliteEntity struct {
	BasicEntity
	// position  *SatellitePositionComponent
	// velocity  *SatelliteVelocityComponent
	// movement  *SatelliteMovementComponent
	satellite satellite.Satellite
	TLE       TLEComponent
}

func (e *SatelliteEntity) ID() uint64 {
	return e.GetBasicEntity().ID()
}

func (e *SatelliteEntity) GetSatelliteEntity() *SatelliteEntity {
	return e
}
