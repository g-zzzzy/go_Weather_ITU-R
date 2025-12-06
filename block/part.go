package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const peanoN uint32 = 6561
const peanoLevel int = 8

type Station struct {
	ID  int
	Lat float64
	Lon float64
	Key uint64 // Morton/Hilbert key
}

// 经纬映射到 [0,N) 网格

// peanoKey 采用 3-进制分割而非 Hilbert 的 2-进制旋转方式
// Nx, Ny 应为 3^k 网格（否则边界自动拉伸）
func peanoKey(x, y uint32, level int) uint64 {
	var key uint64
	var pow uint64 = 1

	// 从最低层往上递归编码
	for i := 0; i < level; i++ {
		// 取当前层的网格编号（mod 3）
		ix := x % 3
		iy := y % 3

		// 转换成 0~8 的 Peano 扫描序
		cell := ix + 3*iy

		key += uint64(cell) * pow
		pow *= 9 // 每层 9 个子格

		// 下一层递归
		x /= 3
		y /= 3
	}
	return key
}

func normalize(lon, lat float64, Nx, Ny uint32) (uint32, uint32) {
	x := (lon + 180.0) / 360.0
	y := (lat + 90.0) / 180.0
	xi := uint32(x * float64(Nx))
	yi := uint32(y * float64(Ny))
	return xi, yi
}

func normalizeSin(lon, lat float64, Nx, Ny uint32) (uint32, uint32) {
	x := (lon + 180.0) / 360.0
	y := (math.Sin(lat*math.Pi/180.0) + 1.0) / 2.0
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

// ----------------- 辅助：KMeans (用于对小段质心做二次聚类) -----------------
type centroid struct {
	Lon float64
	Lat float64
}

func euclidDist2(a, b centroid) float64 {
	dlon := a.Lon - b.Lon
	dlat := a.Lat - b.Lat
	return dlon*dlon + dlat*dlat
}

// kmeans++ 初始化质心（针对小样本，稳定性足够）
func kmeansPlusPlusInit(points []centroid, k int) []centroid {
	rand.Seed(time.Now().UnixNano())
	n := len(points)
	centers := make([]centroid, 0, k)
	// pick one at random
	centers = append(centers, points[rand.Intn(n)])
	dist := make([]float64, n)
	for len(centers) < k {
		var sum float64
		for i := 0; i < n; i++ {
			minD := math.MaxFloat64
			for _, c := range centers {
				d := euclidDist2(points[i], c)
				if d < minD {
					minD = d
				}
			}
			dist[i] = minD
			sum += minD
		}
		// pick new center proportionally to dist
		r := rand.Float64() * sum
		acc := 0.0
		chosen := 0
		for i := 0; i < n; i++ {
			acc += dist[i]
			if acc >= r {
				chosen = i
				break
			}
		}
		centers = append(centers, points[chosen])
	}
	return centers
}

// ----------------- Hilbert 二次聚类逻辑 -----------------
// nextPowerOfTwo 返回 >= x 的最小 2^p
func nextPowerOfTwo(x int) int {
	p := 1
	for p < x {
		p <<= 1
	}
	return p
}

type Point struct {
	x, y float64
	idx  int // 原始段编号
}

// rebalanceGroups: 再平衡，避免组大小差异过大
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

// ----------------- partitionStations 改动：加入 hilbert2 模式 -----------------
func partitionStations(stations []Station, groups int, mode string) [][]Station {
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
	case "hilbertSin":
		// 原先的单次 Hilbert
		for i := range stations {
			x, y := normalizeSin(stations[i].Lon, stations[i].Lat, 2048, 2048)
			stations[i].Key = uint64(hilbertXYToIndex(2048, int(x), int(y)))
		}
	case "peano":
		for i := range stations {
			x, y := normalize(stations[i].Lon, stations[i].Lat, peanoN, peanoN)
			stations[i].Key = peanoKey(x, y, peanoLevel)
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
	case "tree":
		// 四叉树划分（保持你之前实现）
		bbox := [4]float64{-180, 180, -90, 90}
		root := buildQuadtree(stations, bbox, (len(stations)+groups-1)/groups)
		var groupsList [][]Station
		collectLeafGroups(root, &groupsList)
		groupsList = adjustGroups(groupsList, groups)
		return groupsList
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

// ----------------- 以下为你已有的 Quadtree 函数（保持不变） -----------------
type QuadNode struct {
	bbox     [4]float64 // xmin, xmax, ymin, ymax
	points   []Station
	children [4]*QuadNode
}

func buildQuadtree(stations []Station, bbox [4]float64, maxPerNode int) *QuadNode {
	node := &QuadNode{bbox: bbox}
	if len(stations) <= maxPerNode {
		node.points = stations
		return node
	}
	// 划分四个象限
	xmid := (bbox[0] + bbox[1]) / 2
	ymid := (bbox[2] + bbox[3]) / 2
	quadrants := make([][]Station, 4)
	for _, s := range stations {
		switch {
		case s.Lon <= xmid && s.Lat <= ymid: // 左下
			quadrants[0] = append(quadrants[0], s)
		case s.Lon > xmid && s.Lat <= ymid: // 右下
			quadrants[1] = append(quadrants[1], s)
		case s.Lon <= xmid && s.Lat > ymid: // 左上
			quadrants[2] = append(quadrants[2], s)
		case s.Lon > xmid && s.Lat > ymid: // 右上
			quadrants[3] = append(quadrants[3], s)
		}
	}
	bboxes := [][4]float64{
		{bbox[0], xmid, bbox[2], ymid},
		{xmid, bbox[1], bbox[2], ymid},
		{bbox[0], xmid, ymid, bbox[3]},
		{xmid, bbox[1], ymid, bbox[3]},
	}
	for i := 0; i < 4; i++ {
		if len(quadrants[i]) > 0 {
			node.children[i] = buildQuadtree(quadrants[i], bboxes[i], maxPerNode)
		}
	}
	return node
}

func collectLeafGroups(node *QuadNode, groups *[][]Station) {
	if node == nil {
		return
	}
	isLeaf := true
	for i := 0; i < 4; i++ {
		if node.children[i] != nil {
			isLeaf = false
			collectLeafGroups(node.children[i], groups)
		}
	}
	if isLeaf {
		*groups = append(*groups, node.points)
	}
}

func adjustGroups(groups [][]Station, target int) [][]Station {
	// 如果太多，合并小组
	for len(groups) > target {
		// 找到两个最小的组，合并
		minIdx1, minIdx2 := 0, 1
		for i := 0; i < len(groups); i++ {
			for j := i + 1; j < len(groups); j++ {
				if len(groups[i])+len(groups[j]) < len(groups[minIdx1])+len(groups[minIdx2]) {
					minIdx1, minIdx2 = i, j
				}
			}
		}
		merged := append(groups[minIdx1], groups[minIdx2]...)
		// 删除原有两个组
		newGroups := [][]Station{}
		for k := 0; k < len(groups); k++ {
			if k != minIdx1 && k != minIdx2 {
				newGroups = append(newGroups, groups[k])
			}
		}
		newGroups = append(newGroups, merged)
		groups = newGroups
	}

	// 如果太少，拆分大组
	for len(groups) < target {
		// 找到最大的组，切成两半
		maxIdx := 0
		for i := 1; i < len(groups); i++ {
			if len(groups[i]) > len(groups[maxIdx]) {
				maxIdx = i
			}
		}
		group := groups[maxIdx]
		mid := len(group) / 2
		newGroups := [][]Station{}
		for k := 0; k < len(groups); k++ {
			if k == maxIdx {
				newGroups = append(newGroups, group[:mid])
				newGroups = append(newGroups, group[mid:])
			} else {
				newGroups = append(newGroups, groups[k])
			}
		}
		groups = newGroups
	}

	return groups
}

// ----------------- main (略，同你原来 main 保持一致) -----------------
func main() {
	if len(os.Args) != 4 {
		log.Fatalf("Usage: %s <terminal_num><group_num><morton|hilbert|hilbert2|tree>", os.Args[0])
	}
	terminalNum, err1 := strconv.Atoi(os.Args[1])
	groupNum, err2 := strconv.Atoi(os.Args[2])
	mode, err3 := os.Args[3], error(nil)
	if err1 != nil {
		log.Fatalf("Invalid terminal_num: %v", err1)
	}
	if err2 != nil {
		log.Fatalf("Invalid group_num: %v", err2)
	}
	if mode != "morton" && mode != "hilbert" && mode != "hilbert2" && mode != "hilbertSin" && mode != "tree" && mode != "peano" {
		err3 = fmt.Errorf("mode must be 'morton' or 'hilbert' or 'hilbert2' or 'hilbertSin' or 'tree'")
	}
	if err3 != nil {
		log.Fatalf("Invalid mode: %v", err3)
	}

	stations := make([]Station, 0, terminalNum)

	filename_station := "data/terminal.txt"
	file, err := os.Open(filename_station)
	if err != nil {
		fmt.Println("Error Loading Station:", err)
	} else {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		readSta := 0
		for scanner.Scan() && readSta < terminalNum {
			line := scanner.Text()
			parts := strings.Fields(line)
			if len(parts) != 2 {
				continue
			}
			lat, err1 := strconv.ParseFloat(parts[0], 64)
			lon, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 != nil || err2 != nil {
				fmt.Println("Error Lat: ", err1)
				fmt.Println("Error Lon: ", err2)
				continue
			} else {
				stations = append(stations, Station{
					ID:  readSta,
					Lat: lat,
					Lon: lon,
				})
				readSta++
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Scanner Error: ", err)
		}
	}
	terminalGouped := partitionStations(stations, groupNum, mode)
	for g, group := range terminalGouped {
		outFileName := fmt.Sprintf("data/terminals_group_%d.txt", g)
		outFile, err := os.Create(outFileName)
		if err != nil {
			fmt.Println("Error creating file:", err)
			continue
		}
		defer outFile.Close()
		writer := bufio.NewWriter(outFile)
		for _, sta := range group {
			line := fmt.Sprintf("%f %f\n", sta.Lat, sta.Lon)
			_, err := writer.WriteString(line)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				break
			}
		}
		writer.Flush()
		fmt.Printf("Group %d: %d stations written to %s\n", g, len(group), outFileName)
	}
}
