package server

import (
	"SOMAS2023/internal/common/objects"
	"math/rand"
	"slices"

	"github.com/google/uuid"
)

func (s *Server) GetMegaBikes() map[uuid.UUID]objects.IMegaBike {
	return s.megaBikes
}

func (s *Server) GetLootBoxes() map[uuid.UUID]objects.ILootBox {
	return s.lootBoxes
}

func (s *Server) GetAwdi() objects.IAwdi {
	return s.awdi
}

// get a map of megaBikeIDs mapping to the ids of all Bikers that are trying to join it
func (s *Server) GetJoiningRequests(inLimbo []uuid.UUID) map[uuid.UUID][]uuid.UUID {
	// iterate over all agents, if their onBike is false add to the map their id in correspondance of that of their desired bike
	bikeRequests := make(map[uuid.UUID][]uuid.UUID)

	for agentID, agent := range s.GetAgentMap() {
		// don't process joining requests of agents in first round of limbo (ie the ones that have just left the bike)
		if !agent.GetBikeStatus() && !slices.Contains(inLimbo, agentID) {
			bike := agent.GetBike()
			if bike == uuid.Nil {
				continue
			}
			if ids, ok := bikeRequests[bike]; ok {
				bikeRequests[bike] = append(ids, agentID)
			} else {
				bikeRequests[bike] = []uuid.UUID{agentID}
			}
		}
	}
	return bikeRequests
}

// GetRandomBikeId returns the ID of a random bike.
func (s *Server) GetRandomBikeId() uuid.UUID {
	i, targetI := 0, rand.Intn(len(s.GetMegaBikes()))
	// Go doesn't have a sensible way to do this...
	for id := range s.GetMegaBikes() {
		if i == targetI {
			return id
		}
		i++
	}
	panic("no bikes")
}
