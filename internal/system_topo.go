package go_Weather_ITUR

import (
	"fmt"
	"log"
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

func (s *TopoSystem) Update(dt int64, cm *ComponentManager, w *World, t time.Time) {
	log.Printf("TopoSystem update...")
	startTime := time.Now()
	satelliteIDs, err := w.GetSystemEntityIDs("SatelliteSystem")
	if err != nil {
		fmt.Println("[TopoSystem] Error getting satellites:", err)
		return
	}
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
	for i := 0; i < w.numBlocks; i++ {
		start := i * blockSize
		end := (i + 1) * blockSize
		if end > len(satelliteIDs) {
			end = len(satelliteIDs)
		}
		blockSatellites := satelliteIDs[start:end]
		links := make([]Link, 0, len(blockSatellites)*len(stationIDs))
		for _, sourceID := range blockSatellites {
			for _, staID := range stationIDs {
				links = append(links, Link{
					SourceID: sourceID,
					TargetID: staID,
				})
				cnt++
			}
		}
		cm.Links[i] = links
	}

	log.Printf("TopoSystem: Link count: %d", cnt)
	log.Printf("TopoSystem update time: %v", time.Since(startTime))
}
