package teamSOSA

import (
	"SOMAS2023/internal/clients/teamSOSA/agent"
	"SOMAS2023/internal/common/objects"
)

// SOSA agent init function
func GetBiker(baseBiker *objects.BaseBiker) objects.IBaseBiker {
	return agent.NewAgentSOSA(baseBiker)
}
