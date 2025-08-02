package go_Weather_ITUR

import (
	"fmt"
	"log"
	"sort"
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
	cm.Links = cm.Links[:0]

	// 简单实现：全连接
	cnt := 0
	for _, sourceID := range satelliteIDs {
		for _, staID := range stationIDs {

			cm.Links = append(cm.Links, Link{
				SourceID: sourceID,
				TargetID: staID,
			})
			cnt++

		}
	}
	sort.Slice(cm.Links, func(i, j int) bool {
		return cm.Links[i].TargetID < cm.Links[j].TargetID
	})
	log.Printf("TopoSystem: Link count: %d", cnt)
	log.Printf("TopoSystem update time: %v", time.Since(startTime))
}
