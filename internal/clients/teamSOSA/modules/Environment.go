package modules

import (
	objects "SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"math"
	"math/rand"

	"github.com/google/uuid"
)

const (
	AwdiRange = 10
)

type EnvironmentModule struct {
	AgentId   uuid.UUID
	GameState objects.IGameState
	BikeId    uuid.UUID
}

// ----- Lootboxes -----

// returns: the lootbox map for the agent to use
func (e *EnvironmentModule) GetLootBoxes() map[uuid.UUID]objects.ILootBox {
	return e.GameState.GetLootBoxes()
}

// returns: a lootbox object given an id
func (e *EnvironmentModule) GetLootBoxById(lootboxId uuid.UUID) objects.ILootBox {
	return e.GetLootBoxes()[lootboxId]
}

// returns: the coordinates of a lootbox given its id
func (e *EnvironmentModule) GetLootboxPos(lootboxId uuid.UUID) utils.Coordinates {
	return e.GetLootBoxById(lootboxId).GetPosition()
}

// returns: the uuid of a random lootbox
func (e *EnvironmentModule) GetRandomLootbox() uuid.UUID {
	// iterating over a map so will be a random first one
	for _, lootbox := range e.GetLootBoxes() {
		return lootbox.GetID()
	}
	panic("No lootboxes found.")
}

// returns: a filtered lootbox map containing only those of a given colour.
func (e *EnvironmentModule) GetLootBoxesByColour(colour utils.Colour) map[uuid.UUID]objects.ILootBox {
	lootboxes := e.GetLootBoxes()
	lootboxesFiltered := make(map[uuid.UUID]objects.ILootBox)
	for lootboxId, lootbox := range lootboxes {
		if lootbox.GetColour() == colour {
			lootboxesFiltered[lootboxId] = lootbox
		}
	}
	return lootboxesFiltered
}

// returns: uuid of the nearest lootbox
func (e *EnvironmentModule) GetNearestLootbox() uuid.UUID {

	nearestLootbox := uuid.Nil
	minDist := math.MaxFloat64
	// iterate over lootbox map
	for _, lootbox := range e.GetLootBoxes() {
		if e.IsLootboxNearAwdi(lootbox.GetID()) {
			continue
		}
		dist := e.GetDistanceToLootbox(lootbox.GetID())
		if dist < minDist {
			minDist = dist
			nearestLootbox = lootbox.GetID()
		}
	}

	if nearestLootbox == uuid.Nil {
		nearestLootbox = e.GetRandomLootbox()
	}

	return nearestLootbox
}

// returns: uuid of the nearest lootbox chosen from a subset of lootboxes
func (e *EnvironmentModule) GetNearestLootboxFromSubset(subset map[uuid.UUID]objects.ILootBox) uuid.UUID {
	nearest := uuid.Nil
	nDist := math.MaxFloat64

	for id := range subset {
		dist := e.GetDistanceToLootbox(id)
		if dist < nDist {
			nDist = dist
			nearest = id
		}
	}

	return nearest
}

// returns: uuid of the nearest lootbox of a given colour. e.g. nearest blue lootbox
func (e *EnvironmentModule) GetNearestLootboxByColour(colour utils.Colour) uuid.UUID {

	nearestLootbox := e.GetNearestLootbox() // Defaults to nearest lootbox
	minDist := math.MaxFloat64
	for _, lootbox := range e.GetLootBoxesByColour(colour) {
		if e.IsLootboxNearAwdi(lootbox.GetID()) {
			continue
		}
		dist := e.GetDistanceToLootbox(lootbox.GetID())
		if dist < minDist {
			minDist = dist
			nearestLootbox = lootbox.GetID()
		}
	}

	if nearestLootbox == uuid.Nil {
		nearestLootbox = e.GetRandomLootbox()
	}
	return nearestLootbox
}

// returns: uuid of the nearest lootbox of a given colour, from a subset of the total lootboxes.
func (e *EnvironmentModule) GetNearestLootboxByColourFromSubset(color utils.Colour, subset map[uuid.UUID]objects.ILootBox) uuid.UUID {
	nearestLootbox := e.GetNearestLootboxFromSubset(subset) // Defaults to nearest lootbox
	minDist := math.MaxFloat64
	for id := range e.GetLootBoxesByColour(color) {
		if _, ok := subset[id]; !ok {
			continue
		} 
		if e.IsLootboxNearAwdi(id) {
			continue
		}
		dist := e.GetDistanceToLootbox(id)
		if dist < minDist {
			minDist = dist
			nearestLootbox = id
		}

	}
	return nearestLootbox
}

// returns: the distance from the agents bike to a given lootbox id
func (e *EnvironmentModule) GetDistanceToLootbox(lootboxId uuid.UUID) float64 {
	bikePos, lootboxPos := e.GetBikeById(e.BikeId).GetPosition(), e.GetLootBoxById(lootboxId).GetPosition()

	return e.GetDistance(bikePos, lootboxPos)
}

// returns: uuid of the lootbox with the highest 'gain', where gain is the resources divided by the distance.
func (e *EnvironmentModule) GetHighestGainLootbox() uuid.UUID {
	bestGain := float64(0)
	bestLoot := uuid.Nil
	for _, lootboxId := range e.GetLootBoxes() {
		if e.IsLootboxNearAwdi(lootboxId.GetID()) {
			continue
		}
		gain := lootboxId.GetTotalResources() / e.GetDistanceToLootbox(lootboxId.GetID())
		if gain > bestGain {
			bestGain = gain
			bestLoot = lootboxId.GetID()
		}
	}

	if bestLoot == uuid.Nil {
		bestLoot = e.GetRandomLootbox()
	}

	return bestLoot
}

// returns: the uuid of the nearest lootbox that is 'away' from the awdi
func (e *EnvironmentModule) GetNearestLootboxAwayFromAwdi() uuid.UUID {
	// Find positions.
	bikePos := e.GetBikeById(e.BikeId).GetPosition()
	awdiPos := e.GetAwdi().GetPosition()

	// Find position away from awdi.
	deltaX := awdiPos.X - bikePos.X
	deltaY := awdiPos.Y - bikePos.Y

	awayX := bikePos.X - deltaX
	awayY := bikePos.Y - deltaY
	awayPos := utils.Coordinates{X: awayX, Y: awayY}

	// Find nearest lootbox away from awdi.
	minLoot := uuid.Nil
	minDist := math.MaxFloat64
	for id, lootbox := range e.GetLootBoxes() {
		dist := e.GetDistance(awayPos, lootbox.GetPosition())
		if dist < minDist {
			minDist = dist
			minLoot = id
		}
	}
	return minLoot
}

// ----- Bikes -----

// returns: the awdi object
func (e *EnvironmentModule) GetAwdi() objects.IAwdi {
	return e.GameState.GetAwdi()
}

// returns: the megabike map
func (e *EnvironmentModule) GetBikes() map[uuid.UUID]objects.IMegaBike {
	return e.GameState.GetMegaBikes()
}

// returns: the megabike object given a bike id
func (e *EnvironmentModule) GetBikeById(bikeId uuid.UUID) objects.IMegaBike {
	return e.GetBikes()[bikeId]
}

// returns: your own bike object
func (e *EnvironmentModule) GetBike() objects.IMegaBike {
	return e.GetBikeById(e.BikeId)
}

// returns: the orientation of the bike
func (e *EnvironmentModule) GetBikeOrientation() float64 {
	return e.GetBikeById(e.BikeId).GetOrientation()
}

// returns: the biker in the whole game with the maximum trust (not called anywhere)
func (e *EnvironmentModule) GetBikerWithMaxTrust(ap *AgentParameters) IDTrustPair {
	fellowBikers := e.GetBikerAgents()
	maxTrustAgentId := uuid.Nil
	maxTrust := -2.0
	for _, fellowBiker := range fellowBikers {
		if trust, ok := ap.TrustNetwork[e.AgentId]; ok {
			if trust >= maxTrust {
				maxTrustAgentId = fellowBiker.GetID()
				maxTrust = trust
			}
		}
	}
	return IDTrustPair{ID: maxTrustAgentId, Trust: maxTrust}
}

// returns: the biker in the whole game with the minimum trust (n
func (e *EnvironmentModule) GetBikerWithMinTrust(ap *AgentParameters) IDTrustPair {
	fellowBikers := e.GetBikerAgents()
	minTrustAgentId := uuid.Nil
	minTrust := math.MaxFloat64
	for _, fellowBiker := range fellowBikers {
		if trust, ok := ap.TrustNetwork[e.AgentId]; ok {
			if trust < minTrust {
				minTrustAgentId = fellowBiker.GetID()
				minTrust = trust
			}
		}
	}

	if minTrustAgentId != uuid.Nil && minTrustAgentId != e.AgentId {
		// If minSC is nil or !us, then return the culprit.
		return IDTrustPair{ID: minTrustAgentId, Trust: minTrust}
	}
	// Otherwise, return a random agent.
	if len(fellowBikers) > 1 {
		i, targetI := 0, rand.Intn(len(fellowBikers))
		for id := range fellowBikers {
			if i == targetI {
				return IDTrustPair{ID: id, Trust: minTrust}
			}
			i++
		}
	}
	panic("No agents found to kick off.")

}

// returns: uuid of the bike with the maximum trust
func (e *EnvironmentModule) GetBikeWithMaximumTrust(ap *AgentParameters) uuid.UUID {
	maxAverage := float64(0)
	maxBikeId := uuid.Nil

	bikes := e.GetBikes()
	for bikeId, bike := range bikes {
		totalTrust := float64(0)
		agentCount := float64(len(bike.GetAgents()))

		// Sum up the trust of all agents on this bike
		for _, agent := range bike.GetAgents() {
			agentId := agent.GetID()
			totalTrust += ap.TrustNetwork[agentId]
		}

		// Calculate average trust for this bike, Assume we don't switch to a bike with 0 agents
		if agentCount > 0 {
			averageTrust := totalTrust / agentCount
			if averageTrust > maxAverage {
				maxAverage = averageTrust
				maxBikeId = bikeId
			}
		}
	}

	if maxBikeId != uuid.Nil || maxBikeId == e.BikeId {
		// If found, change to that bike.
		return maxBikeId
	}

	// Otherwise, change to a random bike.
	i, targetI := 0, rand.Intn(len(bikes))
	for id := range bikes {
		if i == targetI {
			return id
		}
		i++
	}
	panic("No bikes found to change to.")

}

// returns: bool reflecting whether a given lootbox is / isn't 'near' to the awdi
func (e *EnvironmentModule) IsLootboxNearAwdi(lootboxId uuid.UUID) bool {
	lootboxPos, awdiPos := e.GetLootBoxById(lootboxId).GetPosition(), e.GetAwdi().GetPosition()

	return e.GetDistance(lootboxPos, awdiPos) <= AwdiRange
}

// returns: distance from the bike to awdi
func (e *EnvironmentModule) GetDistanceToAwdi() float64 {
	bikePos, awdiPos := e.GetBikeById(e.BikeId).GetPosition(), e.GetAwdi().GetPosition()

	return e.GetDistance(bikePos, awdiPos)
}

// returns: bool reflecting whether the awdi is 'near'
func (e *EnvironmentModule) IsAwdiNear() bool {
	return e.GetDistanceToAwdi() <= AwdiRange
}

// returns: a map of uuid->biker objects, containing only agents are who are on a bike.
func (e *EnvironmentModule) GetBikerAgents() map[uuid.UUID]objects.IBaseBiker {
	bikes := e.GetBikes()
	bikerAgents := make(map[uuid.UUID]objects.IBaseBiker)
	for _, bike := range bikes {
		for _, biker := range bike.GetAgents() {
			bikerAgents[biker.GetID()] = biker
		}
	}
	return bikerAgents
}

// ----- Utils -----

// returns: the distance between two objects in the gameworld
func (e *EnvironmentModule) GetDistance(pos1, pos2 utils.Coordinates) float64 {

	return math.Sqrt(math.Pow(pos1.X-pos2.X, 2) + math.Pow(pos1.Y-pos2.Y, 2))
}



// getter

func GetEnvironmentModule(agentId uuid.UUID, gameState objects.IGameState, bikeId uuid.UUID) *EnvironmentModule {
	return &EnvironmentModule{
		AgentId:   agentId,
		GameState: gameState,
		BikeId:    bikeId,
	}
}
