package main

import (
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/server"
	"flag"
	"fmt"
	"math"
)

func initialiseFlagConstants() {
	globals.LootBoxCount = int(float64(*globals.BikerAgentCount) * 2.5)
	bikesNeeded := math.Ceil(float64(*globals.BikerAgentCount) / 8)
	globals.MegaBikeCount = int(bikesNeeded)
}

func main() {
	flag.Parse()
	initialiseFlagConstants()
	fmt.Println("Hello Agents")
	s := server.GenerateServer()
	s.Initialize(100)
	s.Start()
}
