package go_Weather_ITUR

type WeatherSystem struct {
	BasicSystem
	weatherIndexs map[uint64]*WeatherIndexEntity
}

func (s *WeatherSystem) GetEntity() []Identifier {
	var entities []Identifier
	for _, weatherIndex := range s.weatherIndexs {
		entities = append(entities, weatherIndex)
	}
	return entities
}

func (s *WeatherSystem) Add(weatherIndex *WeatherIndexEntity) {
	if s.weatherIndexs == nil {
		s.weatherIndexs = make(map[uint64]*WeatherIndexEntity)
	}
	s.weatherIndexs[weatherIndex.GetBasicEntity().id] = weatherIndex
}

func (s *WeatherSystem) Remove(weatherIndex *WeatherIndexEntity) {
	if s.weatherIndexs != nil {
		delete(s.weatherIndexs, weatherIndex.GetBasicEntity().id)
	}
}

func (s *WeatherSystem) Update(dt int64) {
	s.AddElapsed(dt)
	if s.ShouldUpdate(s.elapsed) {
		println("weather index update")
		// 具体的update
	}
}



