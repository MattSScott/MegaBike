package fypagent

import (
	"MegabikeFYPVersion/internal/agents/FYPAgent/agent"
	"MegabikeFYPVersion/internal/common/objects"
)

// init function for the SOSA agent
func GetBiker(baseBiker *objects.BaseBiker, tendency float64) objects.IBaseBiker {
	return agent.NewFYPAgent(baseBiker, tendency)
}
