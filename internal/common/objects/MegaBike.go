package objects

import (
	utils "MegabikeFYPVersion/internal/common/utils"
	"math"

	"github.com/google/uuid"
)

type IMegaBike interface {
	IPhysicsObject
	AddAgent(biker IBaseBiker)
	RemoveAgent(bikerId uuid.UUID)
	GetAgents() map[uuid.UUID]IBaseBiker
	UpdateMass()
	KickOutAgent() []uuid.UUID
	GetGovernance() utils.Governance
	GetRepresentatives() []uuid.UUID
	GetKickedOutCount() int
	ResetKickedOutCount()
	SetGovernance(governance utils.Governance)
	SetRepresentatives(reps []uuid.UUID)
}

type MegaBike struct {
	*PhysicsObject
	agents          map[uuid.UUID]IBaseBiker // agents on the bike
	kickedOutCount  int					
	governance      utils.Governance // type of regime this bike is under
	representatives []uuid.UUID      // slice of the representatives for this bike
}

// constructor
func GetMegaBike(governance utils.Governance) *MegaBike {
	return &MegaBike{
		agents:          make(map[uuid.UUID]IBaseBiker),
		PhysicsObject:   GetPhysicsObject(utils.MassBike),
		governance:      governance,
		representatives: make([]uuid.UUID, 0),
	}
}

// ----- Overriding some PhysicsObject stuff -----

// updates: the total force of the Megabike based on the biker's force
func (mb *MegaBike) UpdateForce() {
	if len(mb.agents) == 0 {
		mb.force = 0.0
	}
	totalPedal := 0.0
	totalBrake := 0.0
	for _, agent := range mb.agents {
		force := agent.GetForces()

		if force.Pedal != 0 {
			totalPedal += float64(force.Pedal)
		} else {
			totalBrake += float64(force.Brake)
		}
	}
	mb.force = (float64(totalPedal) - float64(totalBrake))
}

// updates: the final orientation of the Megabike, between -1 and 1 (-180° to 180°), given the Biker's Turning forces
func (mb *MegaBike) UpdateOrientation() {
	var xSum, ySum float64
	numOfSteeringAgents := 0

	for _, agent := range mb.agents {
		turningDecision := agent.GetForces().Turning
		if turningDecision.SteerBike {
			numOfSteeringAgents += 1

			// Ensure input is between -1 and 1 (wrap around if in excess)
			steeringForce := math.Mod(turningDecision.SteeringForce+1.0, 2.0) - 1.0

			// Convert steering force to cartesian coordinates and sum up
			angle := math.Pi * float64(steeringForce)
			xSum += math.Cos(angle) // X component of the vector
			ySum += math.Sin(angle) // Y component of the vector
		}
	}

	// Average x and y components and return polar form
	if numOfSteeringAgents > 0 {
		avgX := xSum / float64(numOfSteeringAgents)
		avgY := ySum / float64(numOfSteeringAgents)
		mb.orientation += math.Atan2(avgY, avgX) / math.Pi // Converts back to -1 to 1 range
	}
}

// ----- General -----

// adds an agent to the bike
func (mb *MegaBike) AddAgent(biker IBaseBiker) {
	mb.agents[biker.GetID()] = biker
}

// removes: an agent from the bike, given its ID
func (mb *MegaBike) RemoveAgent(bikerId uuid.UUID) {
	delete(mb.agents, bikerId)
}

// returns: a slice containing the agents on the bike
func (mb *MegaBike) GetAgents() map[uuid.UUID]IBaseBiker {
	return mb.agents
}

// updates: the mass of the bike accounting for all its agents
func (mb *MegaBike) UpdateMass() {
	mass := utils.MassBike
	mass += float64(len(mb.agents))
	mb.mass = mass
}

// returns: a slice of agent ids to kick out
func (mb *MegaBike) KickOutAgent() []uuid.UUID {

	// a map of agent id -> number of votes
	voteCount := make(map[uuid.UUID]int)

	// Count votes for each agent
	for _, agent := range mb.agents {
		agentVotes := agent.VoteForKickout() // Assuming this now returns map[uuid.UUID]int, where each agent ID is mapped to 0 (no) or 1 (yes)
		for agentID, vote := range agentVotes {
			if val, ok := voteCount[agentID]; ok {
				voteCount[agentID] = val + vote
			} else {
				voteCount[agentID] = vote
			}
		}
	}

	agentsToKickOut := make([]uuid.UUID, 0)

	if mb.governance == utils.Many {
		// Find all agents where there is consensus, i.e. agents that have everyone voting for them but themselves
		for agentID, votes := range voteCount {
			if votes > (len(mb.agents))-1 {
				agentsToKickOut = append(agentsToKickOut, agentID)
			}
		}
	} else {
		panic("non-democratic bike performing an agent kickout")
	}

	mb.kickedOutCount += len(agentsToKickOut)

	return agentsToKickOut
}

// returns: the governance style of the megabike
func (mb *MegaBike) GetGovernance() utils.Governance {
	return mb.governance
}

// returns: a slice containing the UUIDs of the representatives
func (mb *MegaBike) GetRepresentatives() []uuid.UUID {
	return mb.representatives
}

// returns: the count of kicked out agents
func (mb *MegaBike) GetKickedOutCount() int {
	return mb.kickedOutCount
}

// resets: the kicked out count
func (mb *MegaBike) ResetKickedOutCount() {
	mb.kickedOutCount = 0
}

// sets: the governance of the megabike
func (mb *MegaBike) SetGovernance(governance utils.Governance) {
	mb.governance = governance
}

// sets: the representatives of the megabike
func (mb *MegaBike) SetRepresentatives(reps []uuid.UUID) {
	mb.representatives = reps
}
