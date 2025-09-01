package agent

import (
	"MegabikeFYPVersion/internal/agents/FYPAgent/modules"
	"MegabikeFYPVersion/internal/common/objects"
)

// core FYP agent struct
type FYPAgent struct {
	*objects.BaseBiker              // embed the BaseBiker
	Modules            AgentModules // include the agent modules
}

type AgentModules struct {
	Environment     *modules.EnvironmentModule // for environmental understanding
	AgentParameters *modules.AgentParameters   // additional agent fields
}

// constructor
func NewFYPAgent(baseBiker *objects.BaseBiker, tendency float64) *FYPAgent {
	return &FYPAgent{
		BaseBiker: baseBiker,
		Modules: AgentModules{
			Environment:     modules.NewEnvironmentModule(baseBiker.GetID(), baseBiker.GetGameState(), baseBiker.GetBike()),
			AgentParameters: modules.NewAgentParameters(tendency),
		},
	}
}
