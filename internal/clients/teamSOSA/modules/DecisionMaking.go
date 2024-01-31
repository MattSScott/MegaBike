package modules

import (
	"github.com/google/uuid"
)

// DecisionInputs - Inputs for making decisions
type DecisionInputs struct {
	AgentParameters *AgentParameters
	Environment     *EnvironmentModule
	AgentID         uuid.UUID
}

// DecisionOutputs - Struct for outputs of different decision types
// To be fair, it is currently not used. Can be used or not used. Up to you.
type DecisionOutputs struct {
	KickAgentID      uuid.UUID
	ShouldChangeBike bool
	BikeID           uuid.UUID
	GovernanceID     int
}

// DecisionModule - Module for handling various decisions
type DecisionModule struct{}

// NewDecisionModule - Constructor for DecisionModule
func NewDecisionModule() *DecisionModule {
	return &DecisionModule{}
}

// Based on social capital, decide which agent to kick through minimum capital
func (dm *DecisionModule) MakeKickDecision(inputs DecisionInputs) uuid.UUID {
	minAgentStruct := inputs.AgentParameters.GetMinimumTrust()
	return minAgentStruct.ID
}

// Accept based on larger than accept threshold
func (dm *DecisionModule) MakeAcceptAgentDecision(inputs DecisionInputs) bool {
	socialCapitalScore := inputs.AgentParameters.TrustNetwork[inputs.AgentID]
	return socialCapitalScore > AcceptThreshold
}

func (dm *DecisionModule) MakeBikeChangeDecision(inputs DecisionInputs) (bool, uuid.UUID) {
	// Logic to decide on bike change
	shouldChangeBike := false
	bikeID := uuid.Nil
	if inputs.AgentParameters.GetAverageTrust() < LeaveBikeThreshold {
		shouldChangeBike = true
		bikeID = inputs.Environment.GetBikeWithMaximumSocialCapital(inputs.AgentParameters)
	}
	return shouldChangeBike, bikeID
}

// Decide on governance
func (dm *DecisionModule) MakeGovernanceDecision(inputs DecisionInputs) int {
	return 1
}
