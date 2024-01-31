package agent

import (
	"SOMAS2023/internal/clients/teamSOSA/modules"
	"SOMAS2023/internal/common/objects"

	"github.com/google/uuid"
)

type AgentModules struct {
	Environment     *modules.EnvironmentModule
	AgentParameters *modules.AgentParameters
	Decision        *modules.DecisionModule
	Utils           *modules.UtilsModule
	VotedDirection  uuid.UUID
}

type AgentSOSA struct {
	*objects.BaseBiker // Embedding the BaseBiker
	Modules            AgentModules
}

func NewAgentSOSA(baseBiker *objects.BaseBiker) *AgentSOSA {
	baseBiker.GroupID = 2
	return &AgentSOSA{
		BaseBiker: baseBiker,
		Modules: AgentModules{
			Environment:     modules.GetEnvironmentModule(baseBiker.GetID(), baseBiker.GetGameState(), baseBiker.GetBike()),
			AgentParameters: modules.NewAgentParameters(),
			Decision:        modules.NewDecisionModule(),
			Utils:           modules.NewUtilsModule(),
			VotedDirection:  uuid.Nil,
		},
	}
}
