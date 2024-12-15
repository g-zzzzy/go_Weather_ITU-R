package go_Weather_ITUR

type WeatherIndexEntity struct{
	BasicEntity
	indexs WeatherIndexComponent
}

func (e *WeatherIndexEntity) ID() uint64 {
	return e.GetBasicEntity().ID()
}

func (e *WeatherIndexEntity) GetWeatherEntity() *WeatherIndexEntity {
	return e
}