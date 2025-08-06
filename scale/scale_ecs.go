package main

import (
	"fmt"
	internal "go_Weather_ITUR/internal"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
	"unsafe"
)

func main() {
	if len(os.Args) != 7 {
		log.Fatalf("Usage: %s <station_num> <satellite_nums> <round> <numWGs> <numBlocks> <parallelFlag>", os.Args[0])
	}
	// 解析参数
	stationCount, err1 := strconv.Atoi(os.Args[1])
	satelliteCount, err2 := strconv.Atoi(os.Args[2])
	round, err3 := strconv.Atoi(os.Args[3])
	numWGs, err4 := strconv.Atoi(os.Args[4])
	numBlocks, err5 := strconv.Atoi(os.Args[5])
	parallelFlag, err6 := strconv.Atoi(os.Args[6])
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		log.Fatalf("Invalid arguments: %v, %v, %v", err1, err2, err3)
	}
	fmt.Printf("ECS running with %d stations, %d satellites, %d numWGs and %d numBlocks for %d rounds in %d parallel pattern\n", stationCount, satelliteCount, numWGs, numBlocks, round, parallelFlag)

	startTime := time.Now()
	runtime.GOMAXPROCS(runtime.NumCPU())

	world := internal.NewWorld(numBlocks, numWGs, parallelFlag)

	satelliteSystem := internal.NewSatelliteSystem(5)
	topoSystem := internal.NewTopoSystem(10)
	stationSystem := internal.NewStationSystem(10)
	attenuationSystem := internal.NewAttenuationSystem(10)

	world.AddSystem(internal.SatelliteSystemType, satelliteSystem)
	world.AddSystem(internal.StationSystemType, stationSystem)
	world.AddSystem(internal.TopoSystemType, topoSystem)
	world.AddSystem(internal.AttenuationSystemType, attenuationSystem)

	world.InitSatellite(satelliteCount, satelliteSystem)
	world.InitStation(stationCount, stationSystem)

	// for i := 0; i < 100; i++ {

	for i := 0; i < round; i++ {
		world.Update(1)
		log.Printf("Run %d complete\n", i+1)
	}

	endTime := time.Now()
	log.Println("Total update time: ", endTime.Sub(startTime))
	log.Println("linkCache size:", unsafe.Sizeof(internal.Link{}))
}
