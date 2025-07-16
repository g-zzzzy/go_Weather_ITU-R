package go_Weather_ITUR

type ComponentManager struct {
	TLEComponents           map[EntityID]TLEComponent
	SatelliteSGP4Components map[EntityID]SatelliteSGP4Component

	SatelliteMovementComponents []SatelliteMovementComponent
	MovementEntityToIndex       map[EntityID]int

	StationPositionComponents []StationPositionComponent
	StationEntityToIndex      map[EntityID]int

	WeatherComponents    []WeatherComponent
	WeatherEntityToIndex map[EntityID]int

	AttenuationInputComponents    []AttenuationInputComponent
	AttenuationInputEntityToIndex map[EntityID]int

	// AttenuationOutputComponents    []AttenuationOutputComponent
	// AttenuationOutputEntityToIndex map[EntityID]int

	LinkComponents map[LinkKey]LinkComponent
}

type LinkKey struct {
	SourceID EntityID
	TargetID EntityID
}
