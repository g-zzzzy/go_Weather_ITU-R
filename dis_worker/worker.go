package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	satMutex sync.RWMutex
)

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
	startTerm, _ := strconv.Atoi(taskAlloc["start_term"])
	endTerm, _ := strconv.Atoi(taskAlloc["end_term"])
	log.Printf("节点%d任务: 卫星[%d-%d], 终端[%d-%d]",
		*nodeID, startSat, endSat, startTerm, endTerm)

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
	world.InitStationRange(startTerm, endTerm, stationSystem)

	// 3. 订阅Redis频道（分离不同频道的处理逻辑）
	epochSub := rdb.Subscribe(ctx, "epoch-start")
	defer epochSub.Close()

	regionSub := rdb.Subscribe(ctx, "region-global")
	defer regionSub.Close()

	syncSub := rdb.Subscribe(ctx, "sat-sync-complete")
	defer syncSub.Close()

	// 启动goroutine处理region-global消息（接收其他节点的卫星位置）
	go func() {
		for {
			msg, err := regionSub.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("接收消息失败: %v", err)
				continue
			}

			parts := strings.Split(msg.Payload, ",")
			if len(parts) != 5 || parts[0] != "sat" {
				continue
			}
			satID, _ := strconv.Atoi(parts[1])
			x, _ := strconv.ParseFloat(parts[2], 64)
			y, _ := strconv.ParseFloat(parts[3], 64)
			z, _ := strconv.ParseFloat(parts[4], 64)

			// 加锁更新全局缓存，避免并发写入冲突
			satMutex.Lock()
			world.Components.GlobalSatPositions[satID] = go_Weather_ITUR.SatelliteMovementComponent{
				EntityID: go_Weather_ITUR.EntityID(satID),
				PosX:     x,
				PosY:     y,
				PosZ:     z,
			}
			world.GlobalIDs = append(world.GlobalIDs, go_Weather_ITUR.EntityID(satID))
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

		epoch, err := strconv.Atoi(msg.Payload)
		if err != nil {
			log.Printf("解析epoch编号失败: %v", err)
			continue
		}
		log.Printf("节点%d收到epoch %d 启动信号", *nodeID, epoch)
		startTime := time.Now()

		// 5. 执行本地更新逻辑
		// a. 更新本地卫星位置
		world.Systems[go_Weather_ITUR.SatelliteSystemType].Update(
			1000, world.Components, world, time.Now())

		// b. 批量发布本地卫星位置（控制速率，避免Redis消息风暴）
		publishBatchSize := 100 // 每批发布100个卫星位置
		for i := startSat; i < endSat; i += publishBatchSize {
			end := i + publishBatchSize
			if end > endSat {
				end = endSat
			}
			// 批量发布减少Redis请求次数
			for satID := i; satID < end; satID++ {
				pos := world.Components.SatelliteMovementComponents[satID]
				rdb.Publish(ctx, "region-global",
					fmt.Sprintf("sat,%d,%.6f,%.6f,%.6f", satID, pos.PosX, pos.PosY, pos.PosZ))
			}
			time.Sleep(10 * time.Millisecond) // 轻微延迟，避免拥塞
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
		log.Printf("节点%d：共缓存%d/%d颗卫星位置", *nodeID, len(world.Components.GlobalSatPositions), *totalSat)
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
			fmt.Sprintf("node:%d,elapsed:%d", *nodeID, elapsed))
		log.Printf("节点%d完成epoch %d，耗时%dms", *nodeID, epoch, elapsed)

		// 清理本轮缓存，准备下一轮
		// satMutex.Lock()
		// globalSatPositions = make(map[int]go_Weather_ITUR.SatelliteMovementComponent)
		// satMutex.Unlock()

	}
}
