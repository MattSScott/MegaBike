package modules

import (
	objects "MegabikeFYPVersion/internal/common/objects"
	"MegabikeFYPVersion/internal/common/utils"
	"math"

	"github.com/google/uuid"
)

type EnvironmentModule struct {
	AgentId   uuid.UUID          // our agent id
	GameState objects.IGameState // the game state
	BikeId    uuid.UUID          // our bike id
}

// constructor
func NewEnvironmentModule(agentId uuid.UUID, gameState objects.IGameState, bikeId uuid.UUID) *EnvironmentModule {
	return &EnvironmentModule{
		AgentId:   agentId,
		GameState: gameState,
		BikeId:    bikeId,
	}
}

// ----- Lootboxes -----

// returns: a lootbox object given an id
func (e *EnvironmentModule) GetLootBoxById(lootboxId uuid.UUID) objects.ILootBox {
	return e.GetLootBoxes()[lootboxId]
}

// returns: the lootbox map for the agent to use
func (e *EnvironmentModule) GetLootBoxes() map[uuid.UUID]objects.ILootBox {
	return e.GameState.GetLootBoxes()
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

// returns: the orientation of the bike
func (e *EnvironmentModule) GetBikeOrientation() float64 {
	return e.GetBikeById(e.BikeId).GetOrientation()
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

// returns: the forces to a target coordinate
func (e *EnvironmentModule) GetForcesToTarget(agentPosition, targetPosition utils.Coordinates) utils.Forces {

	deltaX := targetPosition.X - agentPosition.X
	deltaY := targetPosition.Y - agentPosition.Y
	angle := math.Atan2(deltaY, deltaX)
	normalisedAngle := angle / math.Pi
	turningDecision := utils.TurningDecision{
		SteerBike:     true,
		SteeringForce: normalisedAngle,
	}
	return utils.Forces{
		Pedal:   utils.BikerMaxForce,
		Brake:   0.0,
		Turning: turningDecision,
	}
}

// GetForcesToTargetWithDirectionOffset calculates the forces to be applied on an agent to steer towards a target position,
// taking into account a specified degree of angular offset.
func (e *EnvironmentModule) GetForcesToTargetWithDirectionOffset(force, degree float64, currPos, targetPos utils.Coordinates) utils.Forces {
	deltaX := targetPos.X - currPos.X
	deltaY := targetPos.Y - currPos.Y
	angle := math.Atan2(deltaY, deltaX)
	normalisedAngle := angle/math.Pi + math.Remainder(degree, 2)

	if normalisedAngle < -1 {
		normalisedAngle = normalisedAngle + 2
	} else if normalisedAngle > 1 {
		normalisedAngle = normalisedAngle - 2
	}
	turningDecision := utils.TurningDecision{
		SteerBike:     true,
		SteeringForce: normalisedAngle,
	}
	return utils.Forces{
		Pedal:   force,
		Brake:   0.0,
		Turning: turningDecision,
	}
}
