package agent

import (
	"SOMAS2023/internal/clients/teamSOSA/modules"
	"SOMAS2023/internal/common/objects"
)

// core AgentSOSA struct
type AgentSOSA struct {
	*objects.BaseBiker // Embedding the BaseBiker
	Modules            AgentModules
}

type AgentModules struct {
	Environment     *modules.EnvironmentModule
	AgentParameters *modules.AgentParameters
	Utils           *modules.UtilsModule
}


// constructor
func NewAgentSOSA(baseBiker *objects.BaseBiker, tendency float64) *AgentSOSA {
	return &AgentSOSA{
		BaseBiker: baseBiker,
		Modules: AgentModules{
			Environment:     modules.NewEnvironmentModule(baseBiker.GetID(), baseBiker.GetGameState(), baseBiker.GetBike()),
			AgentParameters: modules.NewAgentParameters(tendency),
			Utils:           modules.NewUtilsModule(),
		},
	}
}
