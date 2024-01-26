package server_test

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"SOMAS2023/internal/server"
	"math"
	"testing"

	"github.com/google/uuid"
)

func OnlySpawnBaseBikers(t *testing.T) {
	oldInitFunctions := server.AgentInitFunctions
	server.AgentInitFunctions = []server.AgentInitFunction{nil}
	t.Cleanup(func() {
		server.AgentInitFunctions = oldInitFunctions
	})
}

type NegativeAgent struct {
	*objects.BaseBiker
}

type INegativeAgent interface {
	objects.IBaseBiker
	FurthestLootbox() uuid.UUID
	DictateDirection() uuid.UUID
	ProposeDirection() uuid.UUID
	FinalDirectionVote(proposals map[uuid.UUID]uuid.UUID) voting.LootboxVoteMap
	DecideWeights(utils.Action) map[uuid.UUID]float64
}

func NewNegativeAgent(gameState objects.IGameState) *NegativeAgent {
	return &NegativeAgent{
		BaseBiker: objects.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), gameState),
	}
}

func (a *NegativeAgent) FurthestLootbox() uuid.UUID {
	currLocation := a.GetLocation()
	furthestDist := 0.0
	var nearestBox uuid.UUID
	var currDist float64
	i := 0
	for _, loot := range a.GetGameState().GetLootBoxes() {
		x, y := loot.GetPosition().X, loot.GetPosition().Y
		currDist = math.Sqrt(math.Pow(currLocation.X-x, 2) + math.Pow(currLocation.Y-y, 2))
		if currDist > furthestDist {
			nearestBox = loot.GetID()
			furthestDist = currDist
		}
		i++
	}
	return nearestBox
}

func (a *NegativeAgent) DictateDirection() uuid.UUID {
	return a.FurthestLootbox()
}

// used when trying a negative agent as the leader
func (a *NegativeAgent) ProposeDirection() uuid.UUID {
	return a.FurthestLootbox()
}

// only vote for own proposal
func (a *NegativeAgent) FinalDirectionVote(proposals map[uuid.UUID]uuid.UUID) voting.LootboxVoteMap {
	votes := make(voting.LootboxVoteMap)
	furthest := a.FurthestLootbox()
	for _, proposal := range proposals {
		if furthest == proposal {
			votes[proposal] = 1.0
		} else {
			votes[proposal] = 0.0
		}
	}
	return votes
}

func (a *NegativeAgent) DecideWeights(utils.Action) map[uuid.UUID]float64 {
	weights := make(map[uuid.UUID]float64)
	bike := a.GetGameState().GetMegaBikes()[a.GetBike()]
	agents := bike.GetAgents()
	for _, agent := range agents {
		if agent.GetID() == a.GetID() {
			weights[agent.GetID()] = 1.0
		} else {
			weights[agent.GetID()] = 0.0
		}
	}
	return weights
}

func (a *NegativeAgent) DecideDictatorAllocation() voting.IdVoteMap {
	fellowBikers := a.GetFellowBikers()
	allocation := make(voting.IdVoteMap)
	for _, biker := range fellowBikers {
		if biker.GetID() == a.GetID() {
			allocation[biker.GetID()] = 1.0
		} else {
			allocation[biker.GetID()] = 0.0
		}
	}
	return allocation
}
