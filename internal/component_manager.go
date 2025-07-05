package go_Weather_ITUR

type ComponentManager struct {
	TLEComponents               map[EntityID]TLEComponent
	SatelliteSGP4Components     map[EntityID]SatelliteSGP4Component
	SatelliteMovementComponents map[EntityID]SatelliteMovementComponent
	StationPositionComponents   map[EntityID]StationPositionComponent
	WeatherIndexComponents      map[EntityID]WeatherIndexComponent
	AttenuationComponents       map[LinkKey]AttenuationComponent
	LinkComponents              map[LinkKey]LinkComponent
}

type LinkKey struct {
	SourceID EntityID
	TargetID EntityID
}
