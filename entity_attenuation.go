package go_Weather_ITUR

type AttenuationEntity struct {
	BasicEntity
	weatherIndex      *WeatherIndexComponent
	attenuation       *AttenuationComponent
	stationPosition   *StationPositionComponent
	satelliteMovement *SatelliteMovementComponent
}
