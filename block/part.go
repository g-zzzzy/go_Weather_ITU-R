package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Station struct {
	ID  int
	Lat float64
	Lon float64
	Key uint64 // Morton/Hilbert key
}

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

func partitionStations(stations []Station, groups int) [][]Station {
	// 1. Morton 编码
	Nx, Ny := uint32(2048), uint32(1024)
	for i := range stations {
		xi, yi := normalize(stations[i].Lon, stations[i].Lat, Nx, Ny)
		stations[i].Key = morton(xi, yi)
	}

	// 2. 排序
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

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <terminal_num><group_num>", os.Args[0])
	}
	terminalNum, err1 := strconv.Atoi(os.Args[1])
	groupNum, err2 := strconv.Atoi(os.Args[2])
	if err1 != nil {
		log.Fatalf("Invalid terminal_num: %v", err1)
	}
	if err2 != nil {
		log.Fatalf("Invalid group_num: %v", err2)
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
	terminalGouped := partitionStations(stations, groupNum)
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
