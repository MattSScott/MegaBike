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
	s := server.GenerateServer()
	s.Initialize(100)
	s.Start()
}

// func main() {
// 	flag.Parse()
// 	initialiseFlagConstants()

// 	s := &server.Server{}
// 	s.Initialize(1)
// 	// s.FoundingInstitutions()
// 	iters := 10000
// 	agents := len(s.GetAgentMap())
// 	s.TimeRuleEval(true, iters, agents)
// 	s.TimeRuleEval(false, iters, agents)
// 	fmt.Printf("Run for %d rules for %d agents.\n", iters, agents)
// }
