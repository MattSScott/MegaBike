package modules

import (
	"math/rand"

	"github.com/google/uuid"
)

type AgentParameters struct {
	PlatonicTendency float64               // hardwired value between [0, 1] reflecting the agents personality, where a higher value signifies greater preference for equal and fair distribution.
	TrustNetwork    map[uuid.UUID]float64 		// mapping of uuid -> trust score. trust score ranges from 0 to 1
}


// returns: uuid of agent with minimum trust in your *whole* network, along with their trust score
func (ap *AgentParameters) GetMinimumTrust() IDTrustPair {
	minTrust := 2.0
	minAgentId := uuid.Nil

	for agentId, trust := range ap.TrustNetwork {
		if trust < minTrust {
			minTrust = trust
			minAgentId = agentId
		}
	}
	return IDTrustPair{ID: minAgentId, Trust: minTrust}
}

// returns: uuid of agent with maximum trust in your *whole* network, along with their trust score
func (ap *AgentParameters) GetMaximumTrust() IDTrustPair {
	maxTrust := -2.0
	maxAgentId := uuid.Nil

	for agentId, value := range ap.TrustNetwork {
		if value > maxTrust {
			maxTrust = value
			maxAgentId = agentId
		}
	}
	return IDTrustPair{ID: maxAgentId, Trust: maxTrust}
}

// returns: float representing the sum of the trust in your *whole* network
func (ap *AgentParameters) GetSumOfTrust() float64 {
	var sum = 0.0
	for _, value := range ap.TrustNetwork {
		sum += value
	}
	return sum
}

// returns: float representing average trust in your *whole* network
func (ap *AgentParameters) GetAverageTrust() float64 {
	// Prevent divide
	if len(ap.TrustNetwork) == 0 {
		return 0.5
	}

	sum := ap.GetSumOfTrust()

	return sum / float64(len(ap.TrustNetwork))
}

// updates: an agents' trust value, using the eventValue and eventWeight
func (ap *AgentParameters) UpdateTrustValue(agentID uuid.UUID, eventValue float64) {
	if _, ok := ap.TrustNetwork[agentID]; !ok {
		// if agentID is not in our trust network, assign them our current average trust
		ap.TrustNetwork[agentID] = ap.GetAverageTrust()
		return
	}

	// else, alter their trust value and clamp to ensure it is between 0 and 1
	ap.TrustNetwork[agentID] += eventValue
	ap.TrustNetwork[agentID] = clamp(ap.TrustNetwork[agentID])
}

// returns: float that is 1 if input > 1, and 0 if input < 0
func clamp(value float64) float64 {
	if value > 1.0 {
		return 1.0
	}
	if value < 0.0 {
		return 0.0
	}
	return value
}


func NewAgentParameters() *AgentParameters {
	return &AgentParameters{
		PlatonicTendency: rand.Float64(),
		TrustNetwork:    make(map[uuid.UUID]float64),
	}
}
