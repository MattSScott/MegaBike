package teamSOSA

import (
	"SOMAS2023/internal/clients/teamSOSA/agent"
	"SOMAS2023/internal/common/objects"
)

// init function for the SOSA agent
func GetBiker(baseBiker *objects.BaseBiker, tendency float64) objects.IBaseBiker {
	return agent.NewAgentSOSA(baseBiker, tendency)
}
