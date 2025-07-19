package go_Weather_ITUR

type ComponentManager struct {
	TLEComponents           map[EntityID]TLEComponent
	SatelliteSGP4Components map[EntityID]SatelliteSGP4Component

	SatelliteMovementComponents []SatelliteMovementComponent
	MovementEntityToIndex       map[EntityID]int

	StationPositionComponents []StationPositionComponent
	StationEntityToIndex      map[EntityID]int

	LinkComponents []LinkComponent
}
