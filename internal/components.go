package go_Weather_ITUR

import "github.com/joshuaferrara/go-satellite"

type TLEComponent struct {
	Line1     string
	Line2     string
	GravConst satellite.Gravity
}

type StationPositionComponent struct {
	Lat float64
	Lon float64
}

type SatelliteMovementComponent struct {
	Position satellite.Vector3
	Velocity satellite.Vector3
}

type WeatherIndexComponent struct {
	T             float64 // 2m temperature (K)
	P             float64 // surface pressure	(hPa)
	V_t           float64 // total column water vapour (kg/m2)
	rho           float64 // surface water vapour density (g/m3)
	precipitation float64 // rain (mm/s)
	hr            float64 // rain height (km)
}

type AttenuationComponent struct {
	Attenuation float64
}

type SatelliteSGP4Component struct {
	Satrec satellite.Satellite
}

type LinkComponent struct {
	Connected bool
}
