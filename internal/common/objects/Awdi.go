package objects

import (
	phy "MegabikeFYPVersion/internal/common/physics"
	"MegabikeFYPVersion/internal/common/utils"
	"math"

	"github.com/google/uuid"
)

type IAwdi interface {
	IPhysicsObject
	InjectGameState(gameState IGameState)
	GetTargetID() uuid.UUID
}

type Awdi struct {
	*PhysicsObject
	target    IMegaBike
	gameState IGameState
}

// constructor
func GetAwdi() *Awdi {
	return &Awdi{
		PhysicsObject: GetPhysicsObject(utils.MassAwdi),
	}
}

// constructor
func GetIAwdi() IAwdi {
	return &Awdi{
		PhysicsObject: GetPhysicsObject(utils.MassAwdi),
	}
}

// ----- Awdi Functions -----

// updates: the force of the awdi, which is dependent on whether it has a target bike or not.
func (awdi *Awdi) UpdateForce() {
	// Compute the target Megabike, which will update target field
	awdi.ComputeTarget()

	if awdi.target == nil {
		awdi.force = 0.0 // no target, awdi will not apply a force and eventually come to a stop
	} else {
		awdi.force = utils.AwdiMaxForce // Otherwise apply max force to get to target MegaBike
	}
}

// computes: the target Megabike based on current gameState
func (awdi *Awdi) ComputeTarget() {
	// search for target
	minDistance := math.Inf(1)
	minVelocity := math.Inf(1)
	awdi.target = nil
	for _, bike := range awdi.gameState.GetMegaBikes() {
		if utils.AwdiOnlyTargetsStationaryMegaBike {
			if bike.GetVelocity() != 0.0 {
				continue
			}
		}

		if !utils.AwdiTargetsEmptyMegaBike {
			agentsOnBike := bike.GetAgents()
			if len(agentsOnBike) == 0 {
				continue
			}
		}

		// ignore faster bike
		if bike.GetVelocity() > minVelocity {
			continue
		}

		distance := phy.ComputeDistance(awdi.coordinates, bike.GetPosition())
		// minimize the velocity first
		if bike.GetVelocity() < minVelocity {
			awdi.target = bike
		} else if distance < minDistance { // if same velocity, then minimize distance
			awdi.target = bike
		} else {
			continue
		}
		minVelocity = awdi.target.GetVelocity()
		minDistance = distance
	}
}

// updates: the awdis orientation if it has a target bike
func (awdi *Awdi) UpdateOrientation() {
	// If no target, awdi will not change orientation
	// Otherwise, new orientation is calculated based on positioning of target
	if awdi.target != nil {
		awdi.orientation = phy.ComputeOrientation(awdi.coordinates, awdi.target.GetPosition())
	}
}

// returns: the ID of the target megabike
func (awdi *Awdi) GetTargetID() uuid.UUID {
	if awdi.target != nil {
		return awdi.target.GetID()
	} else {
		return uuid.UUID{}
	}
}

// used to give the awdi access to the game state
func (awdi *Awdi) InjectGameState(gameState IGameState) {
	awdi.gameState = gameState
}
