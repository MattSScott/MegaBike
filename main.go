package main

import (
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/server"
	"flag"
	"math"
)

func initialiseFlagConstants() {
	globals.LootBoxCount = int(float64(*globals.BikerAgentCount) * *globals.LootBoxRatio)
	bikesNeeded := math.Ceil(float64(*globals.BikerAgentCount) / 8)
	globals.MegaBikeCount = int(bikesNeeded)
}

func main() {
	flag.Parse()
	initialiseFlagConstants()
	s := server.GenerateServer() // returns a zero-valued server object
	s.Initialize(5) //sets up the environment, parameter =num of iterations usually 100
	s.Start()
}