package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type Station struct {
	ID  int
	Lat float64
	Lon float64
	Key uint64 // Morton/Hilbert key
}

var (
	totalSat   = flag.Int("total-sat", 20000, "总卫星数量")
	totalTerm  = flag.Int("total-term", 1000, "总终端数量")
	nodeCount  = flag.Int("nodes", 4, "节点数量（node3~node6）")
	redisAddr  = flag.String("redis-addr", "localhost:6380", "Redis地址")
	epochCount = flag.Int("epochs", 5, "总epoch轮数")
)

// 经纬映射到 [0,N) 网格
func normalize(lon, lat float64, Nx, Ny uint32) (uint32, uint32) {
	x := (lon + 180.0) / 360.0
	y := (lat + 90.0) / 180.0
	xi := uint32(x * float64(Nx))
	yi := uint32(y * float64(Ny))
	return xi, yi
}

func morton(x, y uint32) uint64 {
	var ans uint64
	for i := 0; i < 32; i++ {
		ans |= ((uint64(x) >> i) & 1) << (2 * i)
		ans |= ((uint64(y) >> i) & 1) << (2*i + 1)
	}
	return ans
}

func hilbertXYToIndex(n, x, y int) int {
	index := 0
	s := n / 2
	for s > 0 {
		rx := 0
		ry := 0
		if (x & s) > 0 {
			rx = 1
		}
		if (y & s) > 0 {
			ry = 1
		}
		index += s * s * ((3 * rx) ^ ry)
		x, y = rot(s, x, y, rx, ry)
		s /= 2
	}
	return index
}

func rot(n, x, y, rx, ry int) (int, int) {
	if ry == 0 {
		if rx == 1 {
			x = n - 1 - x
			y = n - 1 - y
		}
		x, y = y, x
	}
	return x, y
}

type Point struct {
	x, y float64
	idx  int // 原始段编号
}

func kmeans(points []Point, k, maxIter int) []int {
	rand.Seed(time.Now().UnixNano())
	assignments := make([]int, len(points))
	centroids := make([]Point, k)

	// 随机初始化中心
	for i := 0; i < k; i++ {
		centroids[i] = points[rand.Intn(len(points))]
	}

	for iter := 0; iter < maxIter; iter++ {
		changed := false
		// 分配
		for i, p := range points {
			best := 0
			bestDist := math.MaxFloat64
			for j, c := range centroids {
				d := (p.x-c.x)*(p.x-c.x) + (p.y-c.y)*(p.y-c.y)
				if d < bestDist {
					bestDist = d
					best = j
				}
			}
			if assignments[i] != best {
				assignments[i] = best
				changed = true
			}
		}
		// 更新
		counts := make([]int, k)
		newCentroids := make([]Point, k)
		for i, a := range assignments {
			newCentroids[a].x += points[i].x
			newCentroids[a].y += points[i].y
			counts[a]++
		}
		for j := 0; j < k; j++ {
			if counts[j] > 0 {
				newCentroids[j].x /= float64(counts[j])
				newCentroids[j].y /= float64(counts[j])
				centroids[j] = newCentroids[j]
			}
		}
		if !changed {
			break
		}
	}
	return assignments
}

func rebalanceGroups(groups [][]Station, target int) [][]Station {
	total := 0
	for _, g := range groups {
		total += len(g)
	}
	avg := total / target
	maxSize := int(float64(avg) * 1.5) // 上限
	minSize := int(float64(avg) * 0.5) // 下限

	// flatten 所有点，按组顺序排
	type item struct {
		sta Station
		gid int
	}
	all := []item{}
	for gid, g := range groups {
		for _, s := range g {
			all = append(all, item{sta: s, gid: gid})
		}
	}

	// 遍历 groups，如果某组太大，就切出去分给小组
	for {
		changed := false
		for gi := range groups {
			if len(groups[gi]) > maxSize {
				extra := groups[gi][maxSize:]
				groups[gi] = groups[gi][:maxSize]
				// 找最小的组来接收
				minIdx := -1
				for j := range groups {
					if len(groups[j]) < minSize {
						if minIdx == -1 || len(groups[j]) < len(groups[minIdx]) {
							minIdx = j
						}
					}
				}
				if minIdx != -1 {
					groups[minIdx] = append(groups[minIdx], extra...)
					changed = true
				} else {
					// 实在没有小组，就均匀丢给其它组
					for _, s := range extra {
						best := 0
						for j := 1; j < len(groups); j++ {
							if len(groups[j]) < len(groups[best]) {
								best = j
							}
						}
						groups[best] = append(groups[best], s)
					}
					changed = true
				}
			}
		}
		if !changed {
			break
		}
	}
	return groups
}

func partirionStations(stations []Station, groups int, mode string) [][]Station {
	switch mode {
	case "morton":
		// 1. Morton 编码
		Nx, Ny := uint32(2048), uint32(1024)
		for i := range stations {
			xi, yi := normalize(stations[i].Lon, stations[i].Lat, Nx, Ny)
			stations[i].Key = morton(xi, yi)
		}
	case "hilbert":
		// 原先的单次 Hilbert
		for i := range stations {
			x, y := normalize(stations[i].Lon, stations[i].Lat, 2048, 2048)
			stations[i].Key = uint64(hilbertXYToIndex(2048, int(x), int(y)))
		}
	case "hilbert2":
		// 新的 Hilbert + 二次聚类
		// 1. Hilbert 编码
		for i := range stations {
			x, y := normalize(stations[i].Lon, stations[i].Lat, 2048, 2048)
			stations[i].Key = uint64(hilbertXYToIndex(2048, int(x), int(y)))
		}

		// 2. 按 Hilbert key 排序
		sort.Slice(stations, func(i, j int) bool {
			return stations[i].Key < stations[j].Key
		})

		// 3. 切成比目标组更多的小段（比如 4 倍）
		subSegments := groups * 4
		if subSegments > len(stations) {
			subSegments = len(stations)
		}
		size := (len(stations) + subSegments - 1) / subSegments
		segments := make([][]Station, 0, subSegments)
		centroids := []Point{}
		for i := 0; i < len(stations); i += size {
			end := i + size
			if end > len(stations) {
				end = len(stations)
			}
			seg := stations[i:end]
			segments = append(segments, seg)

			// 质心
			var sx, sy float64
			for _, s := range seg {
				sx += s.Lon
				sy += s.Lat
			}
			cx := sx / float64(len(seg))
			cy := sy / float64(len(seg))
			centroids = append(centroids, Point{x: cx, y: cy, idx: len(segments) - 1})
		}

		// 4. 对质心做 KMeans
		assignments := kmeans(centroids, groups, 50)

		// 5. 把段归簇
		result := make([][]Station, groups)
		for i, seg := range segments {
			c := assignments[i]
			result[c] = append(result[c], seg...)
		}
		result = rebalanceGroups(result, groups)
		return result
	default:
		log.Fatalf("Unknown partitioning mode: %s", mode)
	}

	// 2. 排序（对 morton/hilbert 情况）
	sort.Slice(stations, func(i, j int) bool {
		return stations[i].Key < stations[j].Key
	})

	// 3. 均匀切分
	result := make([][]Station, groups)
	n := len(stations)
	size := (n + groups - 1) / groups
	for g := 0; g < groups; g++ {
		start := g * size
		end := (g + 1) * size
		if end > n {
			end = n
		}
		result[g] = stations[start:end]
	}
	return result
}

type RegionRect struct {
	NodeID int     `json:"node_id"`
	MinLat float64 `json:"min_lat"`
	MaxLat float64 `json:"max_lat"`
	MinLon float64 `json:"min_lon"` // [-180,180)
	MaxLon float64 `json:"max_lon"` // [-180,180)
	Wrap   bool    `json:"wrap"`    // true 表示经度区间跨 180/-180 边界（需要按 "wrap" 处理）
}

func Region(terminalGroup [][]Station, nodeIDs []int) []RegionRect {
	// 把 lon 转到 [0,360)
	mod360 := func(a float64) float64 {
		x := math.Mod(a, 360.0)
		if x < 0 {
			x += 360.0
		}
		return x
	}

	// 计算一组 station 的最小经度覆盖弧（返回 minLon,maxLon 都在 [-180,180) 范围内；wrap=true 表示覆盖跨 dateline）
	computeLonBounds := func(sts []Station) (minLon, maxLon float64, wrap bool) {
		n := len(sts)
		if n == 0 {
			return -180.0, 180.0, false
		}
		lons := make([]float64, 0, n)
		for _, s := range sts {
			lon360 := mod360(s.Lon) // 0..360
			lons = append(lons, lon360)
		}
		sort.Float64s(lons)
		// 找到最大间隙
		maxGap := -1.0
		maxGapIdx := 0
		for i := 0; i < len(lons)-1; i++ {
			g := lons[i+1] - lons[i]
			if g > maxGap {
				maxGap = g
				maxGapIdx = i
			}
		}
		// wrap gap
		wg := lons[0] + 360.0 - lons[len(lons)-1]
		if wg > maxGap {
			maxGap = wg
			maxGapIdx = len(lons) - 1
		}
		start := (maxGapIdx + 1) % len(lons)
		end := maxGapIdx
		min360 := lons[start]
		max360 := lons[end]
		// 判定是否 wrap（如果 start > end 则表示跨界）
		if start <= end {
			wrap = false
		} else {
			wrap = true
		}
		// 转回 [-180,180)
		if min360 >= 180.0 {
			minLon = min360 - 360.0
		} else {
			minLon = min360
		}
		if max360 >= 180.0 {
			maxLon = max360 - 360.0
		} else {
			maxLon = max360
		}
		return minLon, maxLon, wrap
	}
	rects := make([]RegionRect, 0, len(terminalGroup))
	for i, grp := range terminalGroup {
		// lat bounds
		minLat := 90.0
		maxLat := -90.0
		for _, s := range grp {
			if s.Lat < minLat {
				minLat = s.Lat
			}
			if s.Lat > maxLat {
				maxLat = s.Lat
			}
		}
		// 若该组为空，置为全域（或按需要特殊处理）
		if len(grp) == 0 {
			minLat = -90.0
			maxLat = 90.0
		}

		// lon bounds（处理 dateline）
		minLon, maxLon, wrap := computeLonBounds(grp)

		if minLat < -90 {
			minLat = -90
		}
		if maxLat > 90 {
			maxLat = 90
		}

		// 对经度，用 mod360 扩展然后转换回 [-180,180)
		min360 := mod360(minLon)
		max360 := mod360(maxLon)
		min360 = mod360(min360)
		max360 = mod360(max360)
		if min360 >= 180 {
			minLon = min360 - 360
		} else {
			minLon = min360
		}
		if max360 >= 180 {
			maxLon = max360 - 360
		} else {
			maxLon = max360
		}

		rects = append(rects, RegionRect{
			NodeID: nodeIDs[i],
			MinLat: minLat,
			MaxLat: maxLat,
			MinLon: minLon,
			MaxLon: maxLon,
			Wrap:   wrap,
		})
	}
	for _, rect := range rects {
		log.Printf("节点%d的范围：lat: %f-%f, lon: %f-%f", rect.NodeID, rect.MaxLat, rect.MinLat, rect.MinLon, rect.MaxLon)
	}
	return rects

}

// 清理 Redis 中的数据
func cleanupRedis(ctx context.Context, rdb *redis.Client, nodeIDs []int) {
	// 删除任务分配的哈希
	for _, nodeID := range nodeIDs {
		key := fmt.Sprintf("task_alloc:%d", nodeID)
		if err := rdb.Del(ctx, key).Err(); err != nil {
			log.Printf("删除 %s 失败: %v", key, err)
		} else {
			log.Printf("已删除 %s", key)
		}
	}

	// 清理 Pub/Sub 频道消息（Redis pub/sub 本身不会存储消息，只有消费者在线时才能收到）
	// 如果用了 Stream 或 List，要 XDEL / DEL；但你现在只是 pub/sub，不需要清理。

	log.Println("Redis 清理完成")
}

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
	log.Println("成功连接Redis")

	// 读取终端位置数据
	stations := make([]Station, 0, *totalTerm)
	filename_station := "../data/terminal.txt"
	file, err := os.Open(filename_station)
	if err != nil {
		fmt.Println("Error Loading Station:", err)
	} else {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		readSta := 0
		for scanner.Scan() && readSta < *totalTerm {
			line := scanner.Text()
			parts := strings.Fields(line)
			if len(parts) != 2 {
				continue
			}
			lat, err1 := strconv.ParseFloat(parts[0], 64)
			lon, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 != nil || err2 != nil {
				continue
			}
			stations = append(stations, Station{
				ID:  readSta,
				Lat: lat,
				Lon: lon,
			})
			readSta++
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
		}
	}

	// 1. 任务分配：为每个节点分配卫星和终端范围
	terminalGroup := partirionStations(stations, 4, "hilbert")

	nodeIDs := []int{3, 4, 5, 6} // 节点ID列表

	// 区域划分
	regions := Region(terminalGroup, nodeIDs)
	regionJSON, _ := json.Marshal(regions)

	satPerNode := *totalSat / *nodeCount
	// termPerNode := *totalTerm / *nodeCount

	for i, nodeID := range nodeIDs {
		startSat := i * satPerNode
		endSat := startSat + satPerNode
		if i == *nodeCount-1 { // 最后一个节点处理剩余卫星
			endSat = *totalSat
		}

		groupData, _ := json.Marshal(terminalGroup[i])

		// 存储任务分配到Redis
		err := rdb.HSet(ctx, fmt.Sprintf("task_alloc:%d", nodeID), map[string]interface{}{
			"start_sat": startSat,
			"end_sat":   endSat,
			"stations":  groupData,
			"region":    string(regionJSON),
		}).Err()
		if err != nil {
			log.Fatalf("任务分配失败（node %d）: %v", nodeID, err)
		}
		log.Printf("节点%d分配: 卫星[%d-%d], 终端数量=[%d]",
			nodeID, startSat, endSat, len(terminalGroup[i]))
	}

	// 等待所有节点确认收到任务分配
	log.Printf("等待所有节点确认任务分配...")
	initSub := rdb.Subscribe(ctx, "worker-init-ok")
	defer initSub.Close()
	readyNodes := make(map[int]bool)
	for len(readyNodes) < *nodeCount {
		msg, err := initSub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("接收消息失败: %v", err)
			continue
		}
		nodeID, _ := strconv.Atoi(msg.Payload)
		readyNodes[nodeID] = true
		log.Printf("节点%d完成任务初始化", nodeID)
	}
	if err := initSub.Close(); err != nil {
		log.Printf("关闭initSub失败: %v", err)
	}
	log.Printf("所有节点均已确认任务分配")

	// 2. 启动epoch循环
	log.Printf("开始执行%d轮epoch...", *epochCount)
	for epoch := 1; epoch <= *epochCount; epoch++ {
		startTime := time.Now()

		// 发布epoch-start信号（格式：epoch编号）
		if err := rdb.Publish(ctx, "epoch-start", strconv.Itoa(epoch)).Err(); err != nil {
			log.Printf("发布epoch-start失败: %v", err)
			continue
		}
		log.Printf("已发布 epoch %d 启动信号", epoch)

		// 订阅epoch-done，等待所有节点完成
		pubsub := rdb.Subscribe(ctx, "epoch-done")
		defer pubsub.Close()

		doneNodes := make(map[int]bool)
		maxElapsed := int64(0)

		for len(doneNodes) < *nodeCount {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("接收消息失败: %v", err)
				continue
			}

			// 解析消息（格式：node:3,elapsed:123）
			var nodeID int
			var elapsed int64
			fmt.Sscanf(msg.Payload, "node:%d,elapsed:%d", &nodeID, &elapsed)
			doneNodes[nodeID] = true
			log.Printf("节点%d完成epoch %d，耗时%dms", nodeID, epoch, elapsed)

			if elapsed > maxElapsed {
				maxElapsed = elapsed
			}
		}
		if err := pubsub.Close(); err != nil {
			log.Printf("关闭pubsub失败（epoch %d）: %v", epoch, err)
		}

		// 记录本轮总耗时
		totalElapsed := time.Since(startTime).Milliseconds()
		log.Printf("epoch %d 完成，全局最大耗时%dms，总耗时%dms\n",
			epoch, maxElapsed, totalElapsed)

		// 控制epoch间隔（根据实际需求调整）
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("所有epoch完成，发送 finish 信号给所有节点")
	if err := rdb.Publish(ctx, "finish", "1").Err(); err != nil {
		log.Printf("发布finish信号失败: %v", err)
	}

	cleanupRedis(ctx, rdb, nodeIDs)
	log.Println("程序执行完毕，清理Redis数据并退出")

}
