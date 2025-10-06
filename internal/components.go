package go_Weather_ITUR

import "github.com/joshuaferrara/go-satellite"

type TLEComponent struct {
	Line1     string
	Line2     string
	GravConst satellite.Gravity
}

type SatelliteSGP4Component struct {
	Satrec satellite.Satellite
}

type SatelliteMovementComponent struct {
	EntityID         EntityID
	PosX, PosY, PosZ float64
}

type StationPositionComponent struct {
	EntityID EntityID
	Lat, Lon float64
	Key      float64
}

type EnvironmentIndex struct {
	Temperature2m float64
	Precipitation float64
	Pressure      float64 // hPa
}

type Link struct {
	SourceID EntityID
	TargetID EntityID
	Ar       float64
	_        [16]byte
}
