package voting

import (

	"github.com/google/uuid"
)

// BikerID:share
type Allocation map[uuid.UUID]float64

// aggregates the allocations from multiple agents into one, assuming all the agents provide a map with a suggested share for each agent, and these add up to 1
func AggregateAllocations(voters map[uuid.UUID]Allocation, weights map[uuid.UUID]float64) map[uuid.UUID]float64 {

	if len(voters) == 0 {
		panic("no votes provided")
	}

	aggregatedAllocations := make(map[uuid.UUID]float64)

	// initialise allocations to 0.0
	for voter := range voters {
		aggregatedAllocations[voter] = 0.0
	}

	for agentID, allocation := range voters {
		weight := weights[agentID]
		for id, share := range allocation {
			aggregatedAllocations[id] += weight * share
		}
	}

	normalizeFactor := 0.0
	for _, vote := range aggregatedAllocations {
		normalizeFactor += vote
	}
	if normalizeFactor == 0.0 {
		panic("all votes summed to zero")
	}
	// normalising step for all agents involved
	for agentId, vote := range aggregatedAllocations {
		aggregatedAllocations[agentId] = vote / normalizeFactor
	}
	return aggregatedAllocations
}
