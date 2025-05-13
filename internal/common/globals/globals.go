package globals

import "flag"

var BikerAgentCount = flag.Int("agents", 24, "number of agents in simulator")
var LootBoxRatio = flag.Float64("loot", 2.5, "ratio of lootboxes to agents")
var GlobalRuleCount = flag.Int("rules", 0, "number of initial rules in global rule cache")
var StratifyRules = flag.Bool("s", true, "stratify rules by action")
var GoodPlatonicTendency = flag.Float64("goodpt", 1, "platonic tendency for the good agents")
var ProportionOfGoodAgents = flag.Float64("goodagproportion", 0.5, "proportion of agents with the good platonic tendency")

var LootBoxCount int = 140
var MegaBikeCount int = 3
