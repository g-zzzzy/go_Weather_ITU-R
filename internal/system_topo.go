package go_Weather_ITUR

import (
	"fmt"
	"log"
	"math"
	"time"
)

type TopoSystem struct {
	BasicSystem
}

func NewTopoSystem(interval int64) *TopoSystem {
	return &TopoSystem{
		BasicSystem{
			name:     "TopoSystem",
			interval: interval,
		},
	}
}

func isVisible(satComp SatelliteMovementComponent, stationComp StationPositionComponent) bool {
	// const threshold = 5.0 // 阈值：经纬度差值都小于5度则可见（可按需调整）
	// 纬度差值绝对值
	latDiff := math.Abs(satComp.PosX - stationComp.Lat)
	// 经度差值绝对值
	lonDiff := math.Abs(satComp.PosY - stationComp.Lon)
	// 同时满足则可见
	// log.Printf("satPosX=%.2f, satPosY=%.2f, staLat=%.2f, staLon=%.2f", satComp.PosX, satComp.PosY, stationComp.Lat, stationComp.Lon)
	// log.Printf("latDiff=%.2f, lonDiff=%.2f", latDiff, lonDiff)
	return latDiff < 5.0 && lonDiff < 5.0
}

func (s *TopoSystem) Update(dt int64, cm *ComponentManager, w *World, t time.Time) {
	log.Printf("TopoSystem update...")
	startTime := time.Now()

	// satelliteIDs, err := w.GetSystemEntityIDs("SatelliteSystem")
	// if err != nil {
	// 	fmt.Println("[TopoSystem] Error getting satellites:", err)
	// 	return
	// }
	satelliteIDs := w.TargetIDs
	log.Println("[TopoSystem] Satellite nums:", len(satelliteIDs))
	stationIDs, err := w.GetSystemEntityIDs("StationSystem")
	if err != nil {
		fmt.Println("[TopoSystem] Error getting stations:", err)
		return
	}

	// 清空现有的 Links 切片（在逻辑最前面添加）
	cm.Links = make([][]Link, w.numBlocks)

	// 简单实现：全连接
	cnt := 0
	blockSize := (len(satelliteIDs) + w.numBlocks - 1) / w.numBlocks // 每个块的卫星数量

	// blockSize := (len(cm.GlobalSatPositions) + w.numBlocks - 1) / w.numBlocks // 每个块的卫星数量
	for i := 0; i < w.numBlocks; i++ {
		start := i * blockSize
		end := (i + 1) * blockSize
		if end > len(satelliteIDs) {
			end = len(satelliteIDs)
		}
		blockSatellites := satelliteIDs[start:end]
		// links := make([]Link, 0, len(blockSatellites)*len(stationIDs))
		links := make([]Link, 0)
		for _, sourceID := range blockSatellites {

			for _, staID := range stationIDs {
				if !isVisible(cm.TargetSatellites[sourceID], cm.StationPositionComponents[staID]) {
					continue
				}

				links = append(links, Link{
					SourceID: sourceID,
					TargetID: staID,
				})
				cnt++
			}
		}
		cm.Links[i] = links
	}
	cm.LinksNum = cnt
	log.Printf("TopoSystem: Link count: %d", cnt)
	log.Printf("TopoSystem: Link size: %d", len(cm.Links))
	log.Printf("TopoSystem update time: %v", time.Since(startTime))
}
