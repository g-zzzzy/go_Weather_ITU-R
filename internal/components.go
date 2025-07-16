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
	VelX, VelY, VelZ float64
}

type StationPositionComponent struct {
	EntityID EntityID
	Lat, Lon float64
}

type WeatherComponent struct {
	EntityID      EntityID
	Temperature2m float64
	Precipitation float64
	Pressure      float64
}

type AttenuationInputComponent struct {
	EntityID                  EntityID
	SatPosX, SatPosY, SatPosZ float64
	StationLat, StationLon    float64
	Precipitation             float64
}

// type AttenuationOutputComponent struct {
// 	EntityID    EntityID
// 	Attenuation float64
// }

type LinkComponent struct {
	Connected bool
	Ar        float64
}
