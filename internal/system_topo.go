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
	log.Printf("[TopoSystem] Satellite count: %d", len(satelliteIDs))
	if err != nil {
		fmt.Println("[TopoSystem] Error getting satellites:", err)
		return
	}
	stationIDs, err := w.GetSystemEntityIDs("StationSystem")
	log.Printf("[TopoSystem] Station count: %d", len(stationIDs))
	if err != nil {
		fmt.Println("[TopoSystem] Error getting stations:", err)
		return
	}

	// 清空现有的 LinkComponents 切片（在逻辑最前面添加）
	cm.LinkComponents = cm.LinkComponents[:0]

	// 简单实现：全连接
	cnt := 0
	for _, satID := range satelliteIDs {
		sourceID := cm.MovementEntityToIndex[satID]
		for _, staID := range stationIDs {

			cm.LinkComponents = append(cm.LinkComponents, LinkComponent{
				SourceID: sourceID,
				TargetID: cm.StationEntityToIndex[staID],
			})
			cnt++

		}
	}
	log.Printf("TopoSystem: Link count: %d", cnt)
	log.Printf("TopoSystem update time: %v", time.Since(startTime))
}
