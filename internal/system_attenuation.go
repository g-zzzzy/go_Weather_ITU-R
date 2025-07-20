package go_Weather_ITUR

import (
	"go_Weather_ITUR/internal/itur"
	"go_Weather_ITUR/internal/utils"
	"log"
	"time"
)

type AttenuationSystem struct {
	BasicSystem
}

func NewAttenuationSystem(interval int64) *AttenuationSystem {
	return &AttenuationSystem{
		BasicSystem{
			name:     "AttenuationSystem",
			interval: interval,
		},
	}
}

func CalculateSatelliteLink(link *Link, satMovement *SatelliteMovementComponent, stationPos *StationPositionComponent, pre float64) {

	latGS, lonGS := stationPos.Lat, stationPos.Lon
	el := utils.Elevation_angle(satMovement.PosZ, satMovement.PosX, satMovement.PosY, latGS, lonGS)
	f := 22.5 // GHz
	p := 0.1
	hs := 0.1
	R001 := pre // mm/h
	tau := 45.0 // dB
	var Ls float64
	link.Ar = itur.RainAttenuation(latGS, lonGS, f, el, hs, p, R001, tau, Ls)
}

func (s *AttenuationSystem) Update(dt int64, cm *ComponentManager, w *World, t time.Time) {
	log.Printf("AttenuationSystem update...")
	startTime := time.Now()
	cnt := 0
	linkIdx := 0
	// 外层循环：遍历所有站点
	for staIdx := range cm.StationPositionComponents {
		sta := &cm.StationPositionComponents[staIdx]
		pre := cm.StationWeather[staIdx].Precipitation // 当前站点的降水数据

		// 内层循环：批处理该站点的所有卫星链接（匹配 TargetID 为当前站点ID）
		for linkIdx < len(cm.Links) && cm.Links[linkIdx].TargetID == staIdx {
			link := &cm.Links[linkIdx]
			sat := &cm.SatelliteMovementComponents[link.SourceID] // 链接对应的卫星

			// 计算衰减（参数顺序保持不变，仅数据来源调整）
			CalculateSatelliteLink(link, sat, sta, pre)

			linkIdx++
			cnt++
		}
	}

	log.Printf("Attenuation computed count: %d", cnt)
	log.Printf("AttenuationSystem update time: %v", time.Since(startTime))

}
