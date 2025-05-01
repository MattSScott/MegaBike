package objects

/*
The IPhysicsObject is an interface class that all moving objects (Biker and Awdi) must implement.
*/

import (
	utils "SOMAS2023/internal/common/utils"

	"math"

	"github.com/google/uuid"
)

type IPhysicsObject interface {
	
	// ----- Core: Don't need to be overriden -----
	GetID() uuid.UUID				
	GetPosition() utils.Coordinates	
	GetVelocity() float64
	GetOrientation() float64
	GetForce() float64
	GetPhysicalState() utils.PhysicalState
	SetPhysicalState(state utils.PhysicalState)
	CheckForCollision(otherObject IPhysicsObject) bool

	// ----- Customisable: Need to be overridden for each physics object -----
	UpdateForce()
	UpdateOrientation()

}

type PhysicsObject struct {
	id           uuid.UUID
	coordinates  utils.Coordinates
	mass         float64
	acceleration float64
	velocity     float64
	orientation  float64
	force        float64
}

// Constructor
func GetPhysicsObject(mass float64) *PhysicsObject {
	return &PhysicsObject{
		id:           uuid.New(),
		coordinates:  utils.GenerateRandomCoordinates(),
		mass:         mass,
		acceleration: 0.0,
		velocity:     0.0,
		orientation:  0.0,
	}
}


// ----- Core -----

// returns: the unique ID of the object
func (po *PhysicsObject) GetID() uuid.UUID {
	return po.id
}

// returns: the current coordinates of the object
func (po *PhysicsObject) GetPosition() utils.Coordinates {
	return po.coordinates
}

// returns: the velocity of the object
func (po *PhysicsObject) GetVelocity() float64 {
	return po.velocity
}

// returns: the orientation of the object
func (po *PhysicsObject) GetOrientation() float64 {
	return po.orientation
}

// returns: the force of the object
func (po *PhysicsObject) GetForce() float64 {
	return po.force
}

// returns: the physical state of the object
func (po *PhysicsObject) GetPhysicalState() utils.PhysicalState {
	return utils.PhysicalState{
		Position:     po.coordinates,
		Acceleration: po.acceleration,
		Velocity:     po.velocity,
		Mass:         po.mass,
	}
}

// Server must set these variables since it updates the gamestate
func (po *PhysicsObject) SetPhysicalState(state utils.PhysicalState) {
	po.mass = state.Mass
	po.coordinates = state.Position
	po.acceleration = state.Acceleration
	po.velocity = state.Velocity
}

// this will be used to check if a MegaBike has looted a Lootbox or if the Awdi has collided with a MegaBike
func (po *PhysicsObject) CheckForCollision(otherObject IPhysicsObject) bool {
	otherPos := otherObject.GetPosition()
	distance := math.Sqrt(math.Pow(otherPos.X-po.coordinates.X, 2) + math.Pow(otherPos.Y-po.coordinates.Y, 2))
	if distance < utils.CollisionThreshold {
		return true
	} else {
		return false
	}
}

// ----- Customisable -----

// This method will update the force of the PhysicsObject based on the current GameState.
func (po *PhysicsObject) UpdateForce() {
	// for MegaBike, force will be calculated from the bikers
	// For the awdi, force will be calculated from the target MegaBike

}

// This method will update the orientation for the PhysicsObject based on the current gamestate.
func (po *PhysicsObject) UpdateOrientation() {
	// for MegaBike, orientation will be calculated from the bikers
	// For the awdi, orientation will be calculated from the target MegaBike
}


