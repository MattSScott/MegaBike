package agent

import (
	"SOMAS2023/internal/clients/teamSOSA/modules"
	"SOMAS2023/internal/common/objects"

	"github.com/google/uuid"
)

type AgentModules struct {
	Environment     *modules.EnvironmentModule
	AgentParameters *modules.AgentParameters
	Utils           *modules.UtilsModule
	VotedDirection  uuid.UUID
}

type IAgentSOSA interface {
	objects.IBaseBiker
	GetPreferenceForEquality() float64
}

// core agentSosa struct
type AgentSOSA struct {
	*objects.BaseBiker // Embedding the BaseBiker
	Modules            AgentModules
}

func (sosa *AgentSOSA) GetPreferenceForEquality() float64 {
	return sosa.Modules.AgentParameters.PreferenceForEquality
}


// returns: a pointer to an AgentSOSA object
func NewAgentSOSA(baseBiker *objects.BaseBiker) *AgentSOSA {
	return &AgentSOSA{
		BaseBiker: baseBiker,
		Modules: AgentModules{
			Environment:     modules.GetEnvironmentModule(baseBiker.GetID(), baseBiker.GetGameState(), baseBiker.GetBike()),
			AgentParameters: modules.NewAgentParameters(),
			Utils:           modules.NewUtilsModule(),
			VotedDirection:  uuid.Nil,
		},
	}
}
