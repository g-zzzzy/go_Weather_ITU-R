package go_Weather_ITUR

type SatelliteEntity struct {
	BasicEntity
	position SatellitePositionComponent
}

func (e *SatelliteEntity) ID() uint64 {
	return e.GetBasicEntity().ID()
}

func (e *SatelliteEntity) GetSatelliteEntity() *SatelliteEntity {
	return e
}
