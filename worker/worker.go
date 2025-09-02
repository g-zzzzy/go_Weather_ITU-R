package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	go_Weather_ITUR "go_Weather_ITUR/internal" // 替换为你的实际包路径

	"github.com/go-redis/redis/v8"
)

var (
	nodeID    = flag.Int("node-id", 3, "当前节点ID（3~6）")
	totalSat  = flag.Int("total-sat", 20000, "总卫星数量")
	totalTerm = flag.Int("total-term", 2000, "总终端数量")
	redisAddr = flag.String("redis-addr", "localhost:6379", "Redis地址")
	dataDir   = flag.String("data-dir", "data", "数据文件目录")
)

// 全局缓存所有卫星位置（用于链路计算）
var globalSatPositions = make(map[int]go_Weather_ITUR.SatelliteMovementComponent)

func main() {
	flag.Parse()
	ctx := context.Background()

	// 初始化Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: *redisAddr,
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

	// 2. 初始化本地世界（只加载分配的卫星和终端）
	world := go_Weather_ITUR.NewWorld(
		1, // 块数（单节点无需分块）
		4, // 工作线程数
		0, // 并行模式
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

	// 加载本地卫星（只加载分配的范围）
	world.InitSatelliteRange(startSat, endSat, satSystem)
	// 加载本地终端（只加载分配的范围）
	world.InitStationRange(startTerm, endTerm, stationSystem)

	// 3. 订阅必要的Redis频道
	pubsub := rdb.Subscribe(ctx, "epoch-start", "region-global")
	defer pubsub.Close()

	// 启动单独的goroutine处理region-global消息（接收其他节点的卫星位置）
	go func() {
		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("接收消息失败: %v", err)
				continue
			}
			if msg.Channel == "region-global" {
				// 解析卫星位置（格式：sat,ID,x,y,z）
				parts := strings.Split(msg.Payload, ",")
				if len(parts) != 5 || parts[0] != "sat" {
					continue
				}
				satID, _ := strconv.Atoi(parts[1])
				x, _ := strconv.ParseFloat(parts[2], 64)
				y, _ := strconv.ParseFloat(parts[3], 64)
				z, _ := strconv.ParseFloat(parts[4], 64)

				// 更新全局卫星位置缓存
				globalSatPositions[satID] = go_Weather_ITUR.SatelliteMovementComponent{
					EntityID: go_Weather_ITUR.EntityID(satID),
					PosX:     x,
					PosY:     y,
					PosZ:     z,
				}
			}
		}
	}()

	// 4. 等待并处理epoch-start信号
	log.Printf("节点%d准备就绪，等待epoch-start...", *nodeID)
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("接收消息失败: %v", err)
			continue
		}

		if msg.Channel == "epoch-start" {
			epoch, _ := strconv.Atoi(msg.Payload)
			log.Printf("节点%d收到epoch %d 启动信号", *nodeID, epoch)
			startTime := time.Now()

			// 5. 执行本地更新逻辑
			// a. 更新卫星位置
			world.Systems[go_Weather_ITUR.SatelliteSystemType].Update(
				1000, world.Components, world, time.Now())

			// b. 发布本地卫星位置到region-global
			for satID := startSat; satID < endSat; satID++ {
				pos := world.Components.SatelliteMovementComponents[satID]
				rdb.Publish(ctx, "region-global",
					fmt.Sprintf("sat,%d,%f,%f,%f", satID, pos.PosX, pos.PosY, pos.PosZ))
			}

			// c. 等待所有卫星位置同步（简单延迟，实际可优化为计数等待）
			time.Sleep(200 * time.Millisecond)

			// d. 更新终端天气
			world.Systems[go_Weather_ITUR.StationSystemType].Update(
				1000, world.Components, world, time.Now())

			// e. 生成全连接链路（使用全局卫星位置缓存）
			world.Systems[go_Weather_ITUR.TopoSystemType].Update(
				1000, world.Components, world, time.Now())

			// f. 计算链路衰减
			world.Systems[go_Weather_ITUR.AttenuationSystemType].Update(
				1000, world.Components, world, time.Now())

			// 6. 发布完成信号
			elapsed := time.Since(startTime).Milliseconds()
			rdb.Publish(ctx, "epoch-done",
				fmt.Sprintf("node:%d,elapsed:%d", *nodeID, elapsed))
			log.Printf("节点%d完成epoch %d，耗时%dms", *nodeID, epoch, elapsed)
		}
	}
}
