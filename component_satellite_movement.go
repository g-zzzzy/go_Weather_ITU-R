package go_Weather_ITUR

import "github.com/joshuaferrara/go-satellite"

type SatelliteMovementComponent struct {
	position satellite.Vector3
	velocity satellite.Vector3
}
