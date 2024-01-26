package teamSOSA

import (
	"SOMAS2023/internal/clients/teamSOSA/agent"
	"SOMAS2023/internal/common/objects"
)

// this function is going to be called by the server to instantiate bikers in the MVP
func GetBiker(baseBiker *objects.BaseBiker) objects.IBaseBiker {
	baseBiker.GroupID = 2
	return agent.NewAgentSOSA(baseBiker)
}
