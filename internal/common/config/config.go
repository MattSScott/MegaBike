package config

import (
	"flag"
	"math"
)

type Config struct {
	BikerAgentCount int
	LootBoxRatio float64
	GlobalRuleCount int
	StratifyRules bool
	GoodPlatonicTendency float64
	ProportionOfGoodAgents float64
	LootBoxCount int
	MegaBikeCount int
}

// NewConfig parses the command line and returns a populated Config
func NewConfig() Config {
	cfg := Config{}

	flag.IntVar(&cfg.BikerAgentCount, "agents", 24, "number of agents in simulator")
	flag.Float64Var(&cfg.LootBoxRatio, "loot", 2.5, "ratio of lootboxes to agents")
	flag.IntVar(&cfg.GlobalRuleCount, "rules", 0, "number of initial rules in global rule cache")
	flag.BoolVar(&cfg.StratifyRules, "s", true, "stratify rules by action")
	flag.Float64Var(&cfg.GoodPlatonicTendency, "goodpt", 0.75, "platonic tendency for the good agents")
	flag.Float64Var(&cfg.ProportionOfGoodAgents, "goodagproportion", 0.5, "proportion of agents with the good platonic tendency")

	flag.Parse()

	cfg.LootBoxCount = int(float64(cfg.BikerAgentCount) * cfg.LootBoxRatio)
	bikesNeeded := math.Ceil(float64(cfg.BikerAgentCount) / 8)
	cfg.MegaBikeCount = int(bikesNeeded)

	return cfg
}