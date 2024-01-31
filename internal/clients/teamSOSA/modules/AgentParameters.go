package modules

import (
	"github.com/google/uuid"
)

type AgentParameters struct {
	Trustworthiness float64               // intrinsic value for P(conform to action)
	TrustNetwork    map[uuid.UUID]float64 // mapping of social network to trust score
}

func (ap *AgentParameters) GetSumOfTrust() float64 {
	var sum = 0.0
	for _, value := range ap.TrustNetwork {
		sum += value
	}
	return sum
}

func (ap *AgentParameters) GetAverageTrust() float64 {
	// Prevent divide
	if len(ap.TrustNetwork) == 0 {
		return 0.5
	}

	sum := ap.GetSumOfTrust()

	return sum / float64(len(ap.TrustNetwork))
}

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

func clamp(value float64) float64 {
	if value > 1.0 {
		return 1.0
	}
	if value < 0.0 {
		return 0.0
	}
	return value
}

func (ap *AgentParameters) UpdateTrustValue(agentID uuid.UUID, eventValue, eventWeight float64) {
	if _, ok := ap.TrustNetwork[agentID]; !ok {
		ap.TrustNetwork[agentID] = ap.GetAverageTrust()
		return
	}

	ap.TrustNetwork[agentID] += eventValue * eventWeight
	ap.TrustNetwork[agentID] = clamp(ap.TrustNetwork[agentID])
}

// func (sc *SocialCapital) UpdateSocialCapital() {
// 	// fmt.Printf("[UpdateSocialCapital] Social Capital Before: %v\n", sc.SocialCapital)

// 	for id := range sc.SocialNetwork { // Assumes all maps have the same keys.
// 		// Add to Forgiveness Counters.
// 		if _, ok := sc.forgivenessCounter[id]; !ok {
// 			sc.forgivenessCounter[id] = 0.0
// 		}

// 		// Update Forgiveness Counter.
// 		newSocialCapital := ReputationWeight*sc.Reputation[id] + InstitutionWeight*sc.Institution[id] + NetworkWeight*sc.SocialNetwork[id]

// 		if sc.SocialCapital[id] < newSocialCapital {
// 			sc.forgivenessCounter[id] = 0
// 		}

// 		if sc.SocialCapital[id] > newSocialCapital && sc.forgivenessCounter[id] <= 3 {
// 			// Forgive if forgiveness counter is less than 3 and new social capital is less.
// 			sc.forgivenessCounter[id]++
// 			sc.SocialCapital[id] = newSocialCapital + forgivenessFactor*(sc.SocialCapital[id]-newSocialCapital)
// 		} else {
// 			sc.SocialCapital[id] = newSocialCapital
// 		}
// 	}
// 	// fmt.Printf("[UpdateSocialCapital] Social Capital After: %v\n", sc.SocialCapital)
// }

func NewAgentParameters() *AgentParameters {
	return &AgentParameters{
		Trustworthiness: 0.5,
		TrustNetwork:    make(map[uuid.UUID]float64),
	}
}
