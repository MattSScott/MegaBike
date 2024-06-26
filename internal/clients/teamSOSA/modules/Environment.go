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

///
/// Lootboxes
///

func (e *EnvironmentModule) GetLootBoxes() map[uuid.UUID]objects.ILootBox {
	return e.GameState.GetLootBoxes()
}

func (e *EnvironmentModule) GetLootBoxById(lootboxId uuid.UUID) objects.ILootBox {
	return e.GetLootBoxes()[lootboxId]
}

func (e *EnvironmentModule) GetLootboxPos(lootboxId uuid.UUID) utils.Coordinates {
	return e.GetLootBoxById(lootboxId).GetPosition()
}

func (e *EnvironmentModule) GetRandomLootbox() uuid.UUID {
	for _, lootbox := range e.GetLootBoxes() {
		return lootbox.GetID()
	}
	panic("No lootboxes found.")
}

func (e *EnvironmentModule) GetLootBoxesByColor(color utils.Colour) map[uuid.UUID]objects.ILootBox {
	lootboxes := e.GetLootBoxes()
	lootboxesFiltered := make(map[uuid.UUID]objects.ILootBox)
	for _, lootbox := range lootboxes {
		if lootbox.GetColour() == color {
			lootboxesFiltered[lootbox.GetID()] = lootbox
		}
	}
	return lootboxesFiltered
}

func (e *EnvironmentModule) GetNearestLootbox(agentId uuid.UUID) uuid.UUID {
	nearestLootbox := uuid.Nil
	minDist := math.MaxFloat64
	for _, lootbox := range e.GetLootBoxes() {
		if e.IsLootboxNearAwdi(lootbox.GetID()) {
			// fmt.Printf("[GetNearestLootbox] Lootbox %v is near awdi\n", lootbox.GetID())
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

func (e *EnvironmentModule) GetNearestLootboxFromSubset(agentId uuid.UUID, subset map[uuid.UUID]objects.ILootBox) uuid.UUID {
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

func (e *EnvironmentModule) GetNearestLootboxByColor(agentId uuid.UUID, color utils.Colour) uuid.UUID {
	nearestLootbox := e.GetNearestLootbox(agentId) // Defaults to nearest lootbox
	minDist := math.MaxFloat64
	for _, lootbox := range e.GetLootBoxesByColor(color) {
		if e.IsLootboxNearAwdi(lootbox.GetID()) {
			// fmt.Printf("[GetNearestLootboxByColor] Lootbox %v is near awdi\n", lootbox.GetID())
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


func (e *EnvironmentModule) GetNearestLootboxByColorFromSubset(agentId uuid.UUID, color utils.Colour, subset map[uuid.UUID]objects.ILootBox) uuid.UUID {
	// TODO: could be nil
	nearestLootbox := e.GetNearestLootboxFromSubset(agentId, subset) // Defaults to nearest lootbox
	minDist := math.MaxFloat64
	for id := range e.GetLootBoxesByColor(color) {
		if _, ok := subset[id]; !ok {
			continue
		} 
		if e.IsLootboxNearAwdi(id) {
			// fmt.Printf("[GetNearestLootboxByColor] Lootbox %v is near awdi\n", lootbox.GetID())
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

func (e *EnvironmentModule) GetDistanceToLootbox(lootboxId uuid.UUID) float64 {
	bikePos, agntPos := e.GetBikeById(e.BikeId).GetPosition(), e.GetLootBoxById(lootboxId).GetPosition()

	return e.GetDistance(bikePos, agntPos)
}

// Gets lootbox with the highest gain.
// We define gain as the distance to the lootbox divided by the total resources in the lootbox.
func (e *EnvironmentModule) GetHighestGainLootbox() uuid.UUID {
	bestGain := float64(0)
	bestLoot := uuid.Nil
	for _, lootboxId := range e.GetLootBoxes() {
		if e.IsLootboxNearAwdi(lootboxId.GetID()) {
			// fmt.Printf("[GetHighestGainLootbox] Lootbox %v is near awdi\n", lootboxId.GetID())
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

///
/// Bikes
///

func (e *EnvironmentModule) GetAwdi() objects.IAwdi {
	return e.GameState.GetAwdi()
}

func (e *EnvironmentModule) GetBikes() map[uuid.UUID]objects.IMegaBike {
	return e.GameState.GetMegaBikes()
}

func (e *EnvironmentModule) GetBikeById(bikeId uuid.UUID) objects.IMegaBike {
	return e.GetBikes()[bikeId]
}

func (e *EnvironmentModule) GetBike() objects.IMegaBike {
	return e.GetBikeById(e.BikeId)
}

func (e *EnvironmentModule) GetBikeOrientation() float64 {
	return e.GetBikeById(e.BikeId).GetOrientation()
}

func (e *EnvironmentModule) GetBikerWithMaxSocialCapital(ap *AgentParameters) IDTrustPair {
	fellowBikers := e.GetBikerAgents()
	maxSCAgentId := uuid.Nil
	maxSC := -2.0
	for _, fellowBiker := range fellowBikers {
		if sc, ok := ap.TrustNetwork[e.AgentId]; ok {
			if sc >= maxSC {
				maxSCAgentId = fellowBiker.GetID()
				maxSC = sc
			}
		}
	}
	return IDTrustPair{ID: maxSCAgentId, Trust: maxSC}
}

func (e *EnvironmentModule) GetBikerWithMinSocialCapital(ap *AgentParameters) IDTrustPair {
	fellowBikers := e.GetBikerAgents()
	minSCAgentId := uuid.Nil
	minSC := math.MaxFloat64
	for _, fellowBiker := range fellowBikers {
		if sc, ok := ap.TrustNetwork[e.AgentId]; ok {
			if sc < minSC {
				minSCAgentId = fellowBiker.GetID()
				minSC = sc
			}
		}
	}

	if minSCAgentId != uuid.Nil && minSCAgentId != e.AgentId {
		// If minSC is nil or !us, then return the culprit.
		return IDTrustPair{ID: minSCAgentId, Trust: minSC}
	}
	// Otherwise, return a random agent.
	if len(fellowBikers) > 1 {
		i, targetI := 0, rand.Intn(len(fellowBikers))
		for id := range fellowBikers {
			if i == targetI {
				return IDTrustPair{ID: id, Trust: minSC}
			}
			i++
		}
	}
	panic("No agents found to kick off.")
	// return IDTrustPair{ID: uuid.Nil, Trust: math.NaN()}

}

func (e *EnvironmentModule) GetBikeWithMaximumSocialCapital(ap *AgentParameters) uuid.UUID {
	maxAverage := float64(0)
	maxBikeId := uuid.Nil

	bikes := e.GetBikes()
	for bikeId, bike := range bikes {
		totalSocialCapital := float64(0)
		agentCount := float64(len(bike.GetAgents()))

		// Sum up the social capital of all agents on this bike
		for _, agent := range bike.GetAgents() {
			agentId := agent.GetID()
			totalSocialCapital += ap.TrustNetwork[agentId]
		}

		// Calculate average social capital for this bike, Assume we don't swtich to a bike with 0 agents
		if agentCount > 0 {
			averageSocialCapital := totalSocialCapital / agentCount
			if averageSocialCapital > maxAverage {
				maxAverage = averageSocialCapital
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

func (e *EnvironmentModule) IsLootboxNearAwdi(lootboxId uuid.UUID) bool {
	lootboxPos, awdiPos := e.GetLootBoxById(lootboxId).GetPosition(), e.GetAwdi().GetPosition()

	return e.GetDistance(lootboxPos, awdiPos) <= AwdiRange
}

func (e *EnvironmentModule) GetDistanceToAwdi() float64 {
	bikePos, awdiPos := e.GetBikeById(e.BikeId).GetPosition(), e.GetAwdi().GetPosition()

	// fmt.Printf("[GetDistanceToAwdi] Pos of bike: %f\n", bikePos)
	// fmt.Printf("[GetDistanceToAwdi] Pos of Awdi: %f\n", awdiPos)

	return e.GetDistance(bikePos, awdiPos)
}

func (e *EnvironmentModule) IsAwdiNear() bool {
	// fmt.Printf("[IsAwdiNear] Distance to awdi: %f\n", e.GetDistanceToAwdi())
	return e.GetDistanceToAwdi() <= AwdiRange
}

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

///
/// Utils
///

func (e *EnvironmentModule) GetDistance(pos1, pos2 utils.Coordinates) float64 {

	return math.Sqrt(math.Pow(pos1.X-pos2.X, 2) + math.Pow(pos1.Y-pos2.Y, 2))
}

func GetEnvironmentModule(agentId uuid.UUID, gameState objects.IGameState, bikeId uuid.UUID) *EnvironmentModule {
	return &EnvironmentModule{
		AgentId:   agentId,
		GameState: gameState,
		BikeId:    bikeId,
	}
}
