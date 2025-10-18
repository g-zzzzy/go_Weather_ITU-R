package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	go_Weather_ITUR "go_Weather_ITUR/internal" // 替换为实际包路径

	"github.com/go-redis/redis/v8"
)

var (
	nodeID      = flag.Int("node-id", 3, "当前节点ID（3~6）")
	totalSat    = flag.Int("total-sat", 20000, "总卫星数量")
	totalTerm   = flag.Int("total-term", 2000, "总终端数量")
	redisAddr   = flag.String("redis-addr", "10.0.0.53:6380", "Redis地址")
	nodeCount   = flag.Int("node-count", 4, "总节点数量（3~6共4个）") // 新增：总节点数
	syncTimeOut = flag.Duration("sync-timeout", 10*time.Second, "卫星位置同步超时时间")
)

// 全局缓存所有卫星位置（带锁保证并发安全）
var (
	// globalSatPositions = make(map[int]go_Weather_ITUR.SatelliteMovementComponent)
	satMutex          sync.RWMutex
	regionMutex       sync.RWMutex
	localRegionRects  []go_Weather_ITUR.RegionRect
	globalLoadedCount int
)

func toLatLon(x, y, z float64) (lat, lon float64) {
	r := math.Sqrt(x*x + y*y + z*z)
	lat = math.Asin(z/r) * 180.0 / math.Pi
	lon = math.Atan2(y, x) * 180.0 / math.Pi
	return
}

func inRect(lat, lon float64, rect go_Weather_ITUR.RegionRect) bool {
	if lat < rect.MinLat || lat > rect.MaxLat {
		return false
	}
	if !rect.Wrap {
		return !(lon < rect.MinLon || lon > rect.MaxLon)
	}
	// wrap == true 表示跨日界，例如 minLon=170, maxLon=-170，
	// 判断方法为 lon >= minLon OR lon <= maxLon
	return lon >= rect.MinLon || lon <= rect.MaxLon
}

func Region(world *go_Weather_ITUR.World) map[int][]int {
	regionBuckets := make(map[int][]int)
	for satID, pos := range world.Components.SatelliteMovementComponents {
		regionMutex.RLock()
		lat, lon := pos.PosX, pos.PosY
		longitudeDeg := math.Mod(lon+180, 360) - 180
		// 粗略的判断，不涉及圆锥角的投影
		for _, rect := range localRegionRects {
			if inRect(lat, longitudeDeg, rect) {
				regionBuckets[rect.NodeID] = append(regionBuckets[rect.NodeID], satID)
			}
		}
		regionMutex.RUnlock()
	}
	return regionBuckets
}

func main() {
	flag.Parse()
	ctx := context.Background()

	// 初始化Redis客户端（增加连接池配置，避免连接耗尽）
	rdb := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		PoolSize: 10, // 连接池大小
	})
	defer rdb.Close()

	// 检查Redis连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("无法连接Redis: %v", err)
	}
	log.Printf("节点%d成功连接Redis", *nodeID)

	// 1. 从Redis获取任务分配
	taskAlloc, err := rdb.HGetAll(ctx, fmt.Sprintf("task_alloc:%d", *nodeID)).Result()
	if err != nil {
		log.Fatalf("获取任务分配失败: %v", err)
	}

	// 解析任务范围
	startSat, _ := strconv.Atoi(taskAlloc["start_sat"])
	endSat, _ := strconv.Atoi(taskAlloc["end_sat"])

	var stations []go_Weather_ITUR.Station
	if taskAlloc["stations"] != "" {
		if err := json.Unmarshal([]byte(taskAlloc["stations"]), &stations); err != nil {
			log.Fatalf("解析终端分区失败: %v", err)
		}
	}

	// 解析所有区域范围
	if taskAlloc["region"] != "" {
		var rects []go_Weather_ITUR.RegionRect
		if err := json.Unmarshal([]byte(taskAlloc["region"]), &rects); err != nil {
			log.Fatalf("解析区域范围失败: %v", err)
		} else {
			regionMutex.Lock()
			localRegionRects = rects
			regionMutex.Unlock()
			log.Printf("节点%d 已加载 %d 个区域矩形", *nodeID, len(rects))
		}
	}

	// 2. 初始化本地世界
	world := go_Weather_ITUR.NewWorld(
		4, // 块数
		4, // 工作线程数
		0, // 并行模式
		*totalSat,
	)

	// 注册系统
	satSystem := go_Weather_ITUR.NewSatelliteSystem(1000)
	stationSystem := go_Weather_ITUR.NewStationSystem(1000)
	topoSystem := go_Weather_ITUR.NewTopoSystem(1000)
	attenSystem := go_Weather_ITUR.NewAttenuationSystem(1000)

	world.AddSystem(go_Weather_ITUR.SatelliteSystemType, satSystem)
	world.AddSystem(go_Weather_ITUR.StationSystemType, stationSystem)
	world.AddSystem(go_Weather_ITUR.TopoSystemType, topoSystem)
	world.AddSystem(go_Weather_ITUR.AttenuationSystemType, attenSystem)

	// 加载本地卫星和终端
	world.InitSatelliteRange(startSat, endSat, satSystem)
	world.InitStationFromList(stations, stationSystem)
	// world.InitStationRange(startTerm, endTerm, stationSystem)

	log.Printf("节点%d初始化完成: 卫星[%d-%d], 终端数量=[%d]",
		*nodeID, startSat, endSat, len(stations))

	if err := rdb.Publish(ctx, "worker-init-ok", strconv.Itoa(*nodeID)).Err(); err != nil {
		log.Fatalf("发布worker-init-ok失败: %v", err)
	}

	// 3. 订阅Redis频道（分离不同频道的处理逻辑）
	epochSub := rdb.Subscribe(ctx, "epoch-start", "finish")
	defer epochSub.Close()

	channelRegion := fmt.Sprintf("region-pos:%d", *nodeID)
	regionSub := rdb.Subscribe(ctx, channelRegion)
	defer regionSub.Close()

	syncSub := rdb.Subscribe(ctx, "sat-sync-complete")
	defer syncSub.Close()

	// 启动goroutine处理region-pos消息（接收其他节点的卫星位置）
	go func() {
		for {
			msg, err := regionSub.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("接收消息失败: %v", err)
				continue
			}

			var sats []struct {
				ID int     `json:"id"`
				X  float64 `json:"x"`
				Y  float64 `json:"y"`
				Z  float64 `json:"z"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &sats); err != nil {
				log.Printf("解析卫星位置JSON失败: %v", err)
				continue
			}

			globalLoadedCount++
			var newTargetIDs []go_Weather_ITUR.EntityID
			var newTargetSatellites []go_Weather_ITUR.SatelliteMovementComponent

			for _, sat := range sats {
				x := sat.X
				y := sat.Y
				z := sat.Z
				entityID := go_Weather_ITUR.EntityID(globalLoadedCount - 1)
				for len(newTargetSatellites) <= int(entityID) {
					newTargetSatellites = append(newTargetSatellites,
						go_Weather_ITUR.SatelliteMovementComponent{})
				}
				newTargetSatellites[entityID] = go_Weather_ITUR.SatelliteMovementComponent{
					EntityID: entityID,
					PosX:     x,
					PosY:     y,
					PosZ:     z,
				}
				newTargetIDs = append(newTargetIDs, entityID)
				globalLoadedCount++
			}

			satMutex.Lock()
			world.TargetIDs = append(world.TargetIDs, newTargetIDs...)
			world.Components.TargetSatellites = newTargetSatellites
			satMutex.Unlock()
		}

	}()

	syncDoneCh := make(chan struct{})
	go func() {
		doneNodes := make(map[int]bool)
		for {
			msg, err := syncSub.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("接收消息失败: %v", err)
				continue
			}
			nodeID, err := strconv.Atoi(msg.Payload)
			if err == nil {
				doneNodes[nodeID] = true
				log.Printf("已收到节点%d的卫星同步完成信号", nodeID)
				if len(doneNodes) >= *nodeCount {
					close(syncDoneCh) // 所有节点完成
					return
				}
			}
		}
	}()

	// 4. 等待并处理epoch-start信号（增加超时机制）
	log.Printf("节点%d准备就绪，等待epoch-start...", *nodeID)
	for {
		msg, err := epochSub.ReceiveMessage(ctx)

		if err != nil {
			log.Printf("接收失败: %v", err)
			continue
		}

		switch msg.Channel {
		case "finish":
			log.Printf("节点%d收到finish信号，退出程序", *nodeID)
			return
		case "epoch-start":
			epoch, err := strconv.Atoi(msg.Payload)
			if err != nil {
				log.Printf("解析epoch编号失败: %v", err)
				continue
			}
			log.Printf("节点%d收到epoch %d 启动信号", *nodeID, epoch)
			startTime := time.Now()
			globalLoadedCount = 0 // 重置全局加载计数
			world.TargetIDs = world.TargetIDs[:0]
			world.Components.TargetSatellites = world.Components.SatelliteMovementComponents[:0]

			// 5. 执行本地更新逻辑
			// a. 更新本地卫星位置
			world.Systems[go_Weather_ITUR.SatelliteSystemType].Update(
				1000, world.Components, world, time.Now())

			// b. 卫星分区
			log.Printf("节点%d开始处理卫星位置分区", *nodeID)

			regionBuckets := Region(world)

			// c. 批量发布本地卫星位置（控制速率，避免Redis消息风暴）
			for rid, satIDs := range regionBuckets {
				// batch publish, 避免逐条发太多小消息
				log.Printf("节点%d：发布区域%d的卫星位置，共%d颗", *nodeID, rid, len(satIDs))
				batchSize := 200
				for i := 0; i < len(satIDs); i += batchSize {
					j := i + batchSize
					if j > len(satIDs) {
						j = len(satIDs)
					}
					// 构造一个简短的 JSON：[{id:...,x:...,y:...,z:...}, ...]

					payload := make([]map[string]interface{}, 0, j-i)
					for _, sid := range satIDs[i:j] {
						p := world.Components.SatelliteMovementComponents[sid]
						payload = append(payload, map[string]interface{}{
							"id": sid, "x": p.PosX, "y": p.PosY, "z": p.PosZ,
						})
					}
					bs, _ := json.Marshal(payload)
					// 发布到 region-pos:<regionNodeID>
					_ = rdb.Publish(ctx, fmt.Sprintf("region-pos:%d", rid), string(bs)).Err()
				}
			}

			// 本节点发布同步完成信号
			rdb.Publish(ctx, "sat-sync-complete", strconv.Itoa(*nodeID))

			// 等待同步完成或超时
			select {
			case <-syncDoneCh:
				log.Printf("节点%d：所有卫星位置同步完成", *nodeID)
			case <-time.After(*syncTimeOut):
				log.Printf("节点%d：卫星位置同步超时，继续执行（可能数据不完整）", *nodeID)
			}

			// d. 验证是否获取了所有卫星位置
			satMutex.RLock()
			log.Printf("节点%d：共缓存%d颗卫星位置", *nodeID, len(world.Components.TargetSatellites))
			satMutex.RUnlock()

			// e. 执行后续计算（终端天气、链路生成、衰减计算）
			world.Systems[go_Weather_ITUR.StationSystemType].Update(
				1000, world.Components, world, time.Now())

			world.Systems[go_Weather_ITUR.TopoSystemType].Update(
				1000, world.Components, world, time.Now())

			world.Systems[go_Weather_ITUR.AttenuationSystemType].Update(
				1000, world.Components, world, time.Now())

			// 6. 发布完成信号
			elapsed := time.Since(startTime).Milliseconds()
			rdb.Publish(ctx, "epoch-done",
				fmt.Sprintf("node:%d,elapsed:%d,lilnks:%d", *nodeID, elapsed, world.Components.LinksNum))
			log.Printf("节点%d完成epoch %d，处理了%d条链路，耗时%dms", *nodeID, epoch, world.Components.LinksNum, elapsed)

			// 清理本轮缓存，准备下一轮
			// satMutex.Lock()
			// globalSatPositions = make(map[int]go_Weather_ITUR.SatelliteMovementComponent)
			// satMutex.Unlock()
		}

	}
}
