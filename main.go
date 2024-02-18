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

// func main() {
// 	flag.Parse()
// 	initialiseFlagConstants()
// 	fmt.Println("Hello Agents")
// 	s := server.GenerateServer()
// 	s.Initialize(100)
// 	s.Start()
// }

func main() {
	flag.Parse()
	initialiseFlagConstants()

	s := &server.Server{}
	s.Initialize(1)
	s.FoundingInstitutions()
	iters := 10000
	agents := len(s.GetAgentMap())
	s.TimeRuleEval(true, iters, agents)
	s.TimeRuleEval(false, iters, agents)
	fmt.Printf("Run for %d rules for %d agents.\n", iters, agents)
}
