package globals

import "flag"

// const LootBoxCount = BikerAgentCount * 2.5 // 2.5 lootboxes available per Agent
// const MegaBikeCount = 11                   // Megabikes should have 8 riders
// const BikerAgentCount = 56                 // 56 agents in total

var BikerAgentCount = flag.Int("agents", 56, "number of agents in simulator")
var GlobalRuleCount = flag.Int("rules", 100, "number of initial rules in global rule cache")
var StratifyRules = flag.Bool("s", true, "stratify rules by action")

var LootBoxCount int = 140
var MegaBikeCount int = 10
