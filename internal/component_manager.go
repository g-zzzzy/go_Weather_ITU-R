package go_Weather_ITUR

type ComponentManager struct {
	// TLEComponents           map[EntityID]TLEComponent
	SatelliteSGP4Components []SatelliteSGP4Component

	SatelliteMovementComponents []SatelliteMovementComponent
	// MovementEntityToIndex       map[EntityID]int

	StationPositionComponents []StationPositionComponent
	StationWeather            []EnvironmentIndex
	// StationEntityToIndex      map[EntityID]int

	// LinkComponents []LinkComponent
	Links            [][]Link
	TargetSatellites []SatelliteMovementComponent // Preallocate for all satellites
}
