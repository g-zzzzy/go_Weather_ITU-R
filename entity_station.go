package go_Weather_ITUR

type StationEntity struct {
	BasicEntity
	position StationPositionComponent
}

func (e *StationEntity) ID() uint64 {
	return e.GetBasicEntity().ID()
}

func (e *StationEntity) GetStationEntity() *StationEntity {
	return e
}
