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

func (s *TopoSystem) Update(dt int64, cm *ComponentManager, w *World) {
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

	// 简单实现：全连接
	cnt := 0
	for _, satID := range satelliteIDs {
		for _, staID := range stationIDs {
			linkKey := LinkKey{SourceID: satID, TargetID: staID}
			cm.LinkComponents[linkKey] = LinkComponent{
				Connected: true,
			}
			cnt++
			// 调试输出
			// fmt.Printf("[TopoSystem] Linked Sat %d <-> Sta %d\n", satID, staID)
		}
	}
	log.Printf("TopoSystem: Link count: %d", cnt)
	endTime := time.Now()
	log.Printf("TopoSystem update time: %v", endTime.Sub(startTime))
}
