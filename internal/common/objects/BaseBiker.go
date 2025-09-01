package objects

import (
	utils "MegabikeFYPVersion/internal/common/utils"
	voting "MegabikeFYPVersion/internal/common/voting"
	"math"
	"math/rand"

	baseAgent "github.com/MattSScott/basePlatformSOMAS/BaseAgent"
	"github.com/MattSScott/basePlatformSOMAS/messaging"
	"github.com/google/uuid"
)

type IBaseBiker interface {
	baseAgent.IAgent[IBaseBiker]

	// ----- 1. Strategy Functions (to be overridden by a specific agent implementation) -----

	// Decision Making

	DecideJoiningOneAgent(agentId uuid.UUID) bool          // returns: bool if you think this single agent should be accepted
	DecideBikePreferenceOrder() []utils.BikePreferenceData // decides: the order of the agents preferred bikes
	DecideDirection() uuid.UUID                            // returns: lootbox uuid we want to aim towards (the direction)
	DecideForce(direction uuid.UUID)                       // decides: the force the biker is going to pedal with
	DecideAllocation() voting.Allocation                   // returns: map containing bikerID -> distribution (i.e. share of resources)
	HandleAgentUnalive(id uuid.UUID)                       // decides: how to handle a dead agent

	// (not currently used)
	VoteForKickout() map[uuid.UUID]int // returns: map of UUID -> {0,1} for an agent where 0 means 'don't kick' and 1 means 'do kick'
	DecideKickOut() []uuid.UUID        // returns: slice of agents to kick out

	// Messaging

	HandleProposedLootboxMessage(msg ProposedLootboxMessage)
	HandleForcesMessage(msg ForcesMessage)
	HandleConformMessage(msg ConformMessage)
	GetAllRoundMessages([]IBaseBiker) []messaging.IMessage[IBaseBiker]

	// HandleKickoutMessage(msg KickoutAgentMessage)
	// HandleChangeBikeMessage(msg ChangeBikeMessage)
	// GetAllIterationMessages([]IBaseBiker) []messaging.IMessage[IBaseBiker]

	// ----- 2. Core Functions (do not need to override these, can inherit basebiker's) -----

	// Getters

	GetForces() utils.Forces                   // returns: forces for current round
	GetColour() utils.Colour                   // returns: the colour of the lootbox that the agent is currently seeking
	GetLocation() utils.Coordinates            // returns: the agent's location
	GetBike() uuid.UUID                        // returns: the uuid of the bike the agent is currently on
	GetEnergyLevel() float64                   // returns: energy level of the agent
	GetPoints() int                            // returns: the points the agent currently has
	GetBikeStatus() bool                       // returns: whether the biker is on a bike or not
	GetFellowBikers() map[uuid.UUID]IBaseBiker // returns: map containing the bikers on our bike.
	GetFellowBikersSlice() []IBaseBiker        // returns: slice containing the bikers on our bike
	GetRoundDirection() uuid.UUID              // returns: the lootbox the agent wanted to travel toward this round
	GetRoundForces() utils.Forces              // returns: the forces the agent applied this round
	GetRoundDidConform() bool                  // returns: whether the agent conformed to the group decision this round or not
	GetIterationIncome() float64               // returns: the amount of resources this agent recieved this round

	// Setters and updaters

	SetBike(uuid.UUID)                                                           // sets: the megaBikeID. (this is either the id of the bike that the agent is on or the one that it's trying to join)
	SetForces(forces utils.Forces)                                               // sets: the force (called within DecideForce())
	UpdatePoints(pointGained int)                                                // increases: the points of an agent by pointsGained
	UpdateEnergyLevel(energyLevel float64)                                       // increases: the energy level of the agent by the allocated lootbox share or decreases by expended energy
	ToggleOnBike()                                                               // toggles: whether a user is on bike or not. called when removing or adding a biker on a bike
	ResetPoints()                                                                // resets: agents points to 0
	SetRoundDirection(direction uuid.UUID)                                       // sets: the round direction attribute of the agent, i.e. which lootbox it wanted to travel toward
	SetRoundForces(forces utils.Forces)                                          // sets: the round forces attribute of the agent, i.e. what forces it applied
	SetRoundDidConform(didConform bool)                                          // sets: the round did conform attribute of the agent, i.e. whether it conformed to the group direction decision
	UpdateRegimeTrustValues(oneGini float64, someGini float64, manyGini float64) // updates: the agents regime trust attribute
	UpdateIterationIncome(resources float64)                                     // updates: the agent's resource income count
	ResetIterationIncome()                                                       // resets: the agent's resource income
}

type BaseBiker struct {
	*baseAgent.BaseAgent[IBaseBiker]                // embedding BaseAgent
	soughtColour                     utils.Colour   // the lootbox colour this agent seeks
	onBike                           bool           // "is this agent on a bike"
	energyLevel                      float64        // float between [0,1] 
	points                           int            // increases when lootbox matches colour
	forces                           utils.Forces   // the forces the agent is applying
	megaBikeId                       uuid.UUID      // nil if not on bike
	gameState                        IGameState     // for accessing game state info
	roundDecisions                   roundDecisions // to tell other agents at end of round
	iterationIncome                  float64        // the agents' income this iteration
}

// constructor
func GetBaseBiker(totColours utils.Colour, bikeId uuid.UUID, gameState IGameState) *BaseBiker {
	return &BaseBiker{
		BaseAgent:    baseAgent.NewBaseAgent[IBaseBiker](),
		soughtColour: utils.GenerateRandomColour(),
		onBike:       false,
		energyLevel:  1.0,
		points:       0,
		gameState:    gameState,
	}
}

type BikerAction int

const (
	Pedal BikerAction = iota
	ChangeBike
)

type roundDecisions struct {
	Direction  uuid.UUID
	Forces     utils.Forces
	didConform bool
}

// ----- BASE IMPLEMENTATIONS -----

// ----- 1. Strategy Functions (to be overridden by a specific agent implementation) -----

func (bb *BaseBiker) DecideJoiningOneAgent(agentId uuid.UUID) bool {
	return true
}

func (bb *BaseBiker) DecideBikePreferenceOrder() []utils.BikePreferenceData {

	return []utils.BikePreferenceData{}

}

func (bb *BaseBiker) DecideDirection() uuid.UUID {
	nearest := bb.nearestLoot()
	return nearest
}

func (bb *BaseBiker) DecideForce(direction uuid.UUID) {
	// fmt.Println("I'm trying...")
	if direction == uuid.Nil {
		// fmt.Println("But failed.")
		return
	}
	// NEAREST BOX STRATEGY (MVP)
	currLocation := bb.GetLocation()
	currentLootBoxes := bb.gameState.GetLootBoxes()

	// Check if there are lootboxes available and move towards closest one
	if len(currentLootBoxes) > 0 {
		targetPos := currentLootBoxes[direction].GetPosition()

		deltaX := targetPos.X - currLocation.X
		deltaY := targetPos.Y - currLocation.Y
		angle := math.Atan2(deltaY, deltaX)
		normalisedAngle := angle / math.Pi

		// Default BaseBiker will always
		turningDecision := utils.TurningDecision{
			SteerBike:     true,
			SteeringForce: normalisedAngle - bb.gameState.GetMegaBikes()[bb.megaBikeId].GetOrientation(),
		}

		nearestBoxForces := utils.Forces{
			Pedal:   utils.BikerMaxForce,
			Brake:   0.0,
			Turning: turningDecision,
		}
		bb.SetForces(nearestBoxForces)
	} else { // otherwise move away from awdi
		awdiPos := bb.GetGameState().GetAwdi().GetPosition()

		deltaX := awdiPos.X - currLocation.X
		deltaY := awdiPos.Y - currLocation.Y

		// Steer in opposite direction to awdi
		angle := math.Atan2(deltaY, deltaX)
		normalisedAngle := angle / math.Pi

		// Steer in opposite direction to awdi
		var flipAngle float64
		if normalisedAngle < 0.0 {
			flipAngle = normalisedAngle + 1.0
		} else if normalisedAngle > 0.0 {
			flipAngle = normalisedAngle - 1.0
		}

		// Default BaseBiker will always
		turningDecision := utils.TurningDecision{
			SteerBike:     true,
			SteeringForce: flipAngle - bb.gameState.GetMegaBikes()[bb.megaBikeId].GetOrientation(),
		}

		escapeAwdiForces := utils.Forces{
			Pedal:   utils.BikerMaxForce,
			Brake:   0.0,
			Turning: turningDecision,
		}
		bb.SetForces(escapeAwdiForces)
	}
}

func (bb *BaseBiker) DecideAllocation() voting.Allocation {
	bikeID := bb.GetBike()
	fellowBikers := bb.gameState.GetMegaBikes()[bikeID].GetAgents()
	distribution := make(voting.Allocation)
	for _, agent := range fellowBikers {
		if agent.GetID() == bb.GetID() {
			distribution[agent.GetID()] = 1.0
		} else {
			distribution[agent.GetID()] = 0.0
		}
	}
	return distribution
}

func (bb *BaseBiker) HandleAgentUnalive(id uuid.UUID) {

}

func (bb *BaseBiker) VoteForKickout() map[uuid.UUID]int {
	voteResults := make(map[uuid.UUID]int)
	bikeID := bb.GetBike()

	fellowBikers := bb.gameState.GetMegaBikes()[bikeID].GetAgents()
	for _, agent := range fellowBikers {
		agentID := agent.GetID()
		if agentID != bb.GetID() {
			// random votes to other agents
			voteResults[agentID] = rand.Intn(2) // randomly assigns 0 or 1 vote
		}
	}

	return voteResults
}

func (bb *BaseBiker) DecideKickOut() []uuid.UUID {
	return (make([]uuid.UUID, 0))
}

// Messaging - Round

func (bb *BaseBiker) HandleProposedLootboxMessage(msg ProposedLootboxMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

}

func (bb *BaseBiker) HandleForcesMessage(msg ForcesMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.
}

func (bb *BaseBiker) HandleConformMessage(msg ConformMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.
}

func (bb *BaseBiker) GetAllRoundMessages([]IBaseBiker) []messaging.IMessage[IBaseBiker] {
	// For team's agent add your own logic on chosing when your biker should send messages and which ones to send (return)
	proposedLootboxMessage := bb.CreateProposedLootboxMessage()
	forcesMsg := bb.CreateForcesMessage()
	conformMsg := bb.CreateConformMessage()

	return []messaging.IMessage[IBaseBiker]{forcesMsg, proposedLootboxMessage, conformMsg}
}

func (bb *BaseBiker) CreateProposedLootboxMessage() ProposedLootboxMessage {
	// Currently this returns a default/meaningless message
	// For team's agent, add your own logic to communicate with other agents
	return ProposedLootboxMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikersSlice()),
		Lootbox:     uuid.Nil,
	}
}

func (bb *BaseBiker) CreateForcesMessage() ForcesMessage {
	// Currently this returns a default message which sends to all bikers on the biker agent's bike
	// For team's agent, add your own logic to communicate with other agents
	return ForcesMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikersSlice()),
		AgentForces: utils.Forces{
			Pedal: 0.0,
			Brake: 0.0,
			Turning: utils.TurningDecision{
				SteerBike:     false,
				SteeringForce: 0.0,
			},
		},
	}
}

func (bb *BaseBiker) CreateConformMessage() ConformMessage {
	// Currently this returns a default/meaningless message
	// For team's agent, add your own logic to communicate with other agents
	return ConformMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikersSlice()),
		DidConform:  false,
	}
}

// Messaging - Iteration

// func (bb *BaseBiker) HandleKickoutMessage(msg KickoutAgentMessage) {
// 	// Team's agent should implement logic for handling other biker messages that were sent to them.

// 	// agentId := msg.AgentId
// 	// kickout := msg.Kickout
// }

// func (bb *BaseBiker) HandleChangeBikeMessage(msg ChangeBikeMessage) {
// 	// Team's agent should implement logic for handling other biker messages that were sent to them.

// }

// func (bb *BaseBiker) GetAllIterationMessages([]IBaseBiker) []messaging.IMessage[IBaseBiker] {
// 	// For team's agent add your own logic on chosing when your biker should send messages and which ones to send (return)
// 	kickoutMsg := bb.CreatekickoutMessage()
// 	changeBikeMessage := bb.CreateChangeBikeMessage()

// 	return []messaging.IMessage[IBaseBiker]{kickoutMsg, changeBikeMessage}
// }

// func (bb *BaseBiker) CreatekickoutMessage() KickoutAgentMessage {

// 	// Currently this returns a default message which sends to all bikers on the biker agent's bike
// 	// For team's agent, add your own logic to communicate with other agents
// 	return KickoutAgentMessage{
// 		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikersSlice()),
// 		AgentId:     uuid.Nil,
// 	}
// }

// func (bb *BaseBiker) CreateChangeBikeMessage() ChangeBikeMessage {
// 	// Currently this returns a default message which sends to all bikers on the biker agent's bike
// 	// For team's agent, add your own logic to communicate with other agents
// 	return ChangeBikeMessage{
// 		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikersSlice()),
// 		BikeId:      uuid.Nil,
// 	}
// }

// ----- 2. Core Functions (do not need to override these, can inherit basebiker's) -----

// Getters

func (bb *BaseBiker) GetForces() utils.Forces {
	return bb.forces
}

func (bb *BaseBiker) GetColour() utils.Colour {
	return bb.soughtColour
}

func (bb *BaseBiker) GetLocation() utils.Coordinates {

	// the biker itself doesn't technically have a location (as it's on the map only when it's on a bike)
	// in fact this function is only called when the biker needs to make a decision about the pedaling forces

	if !bb.GetBikeStatus() {
		return utils.Coordinates{X: -1, Y: -1}
	}
	megaBikes := bb.gameState.GetMegaBikes()
	return megaBikes[bb.megaBikeId].GetPosition()
}

func (bb *BaseBiker) GetBike() uuid.UUID {
	return bb.megaBikeId
}

func (bb *BaseBiker) GetEnergyLevel() float64 {
	return bb.energyLevel
}

func (bb *BaseBiker) GetPoints() int {
	return bb.points
}

func (bb *BaseBiker) GetBikeStatus() bool {
	return bb.onBike
}

func (bb *BaseBiker) GetFellowBikers() map[uuid.UUID]IBaseBiker {
	// for use in adding / removal
	bikes := bb.gameState.GetMegaBikes()
	if _, ok := bikes[bb.GetBike()]; !ok {
		panic("agents bike not found")
	}
	bike := bikes[bb.GetBike()]
	fellowBikers := bike.GetAgents()
	return fellowBikers
}

func (bb *BaseBiker) GetFellowBikersSlice() []IBaseBiker {
	// for use in messaging
	bikes := bb.gameState.GetMegaBikes()
	if _, ok := bikes[bb.GetBike()]; !ok {
		panic("agents bike not found")
	}
	bike := bikes[bb.GetBike()]
	fellowBikers := bike.GetAgents()
	var fellowBikerSlice []IBaseBiker
	for _, agent := range fellowBikers {
		fellowBikerSlice = append(fellowBikerSlice, agent)
	}

	return fellowBikerSlice
}

func (bb *BaseBiker) GetRoundDirection() uuid.UUID {
	return bb.roundDecisions.Direction
}

func (bb *BaseBiker) GetRoundForces() utils.Forces {
	return bb.roundDecisions.Forces
}

func (bb *BaseBiker) GetRoundDidConform() bool {
	return bb.roundDecisions.didConform
}

func (bb *BaseBiker) GetIterationIncome() float64 {
	return bb.iterationIncome
}

// Setters

func (bb *BaseBiker) SetBike(bikeId uuid.UUID) {
	bb.megaBikeId = bikeId
}

func (bb *BaseBiker) SetForces(forces utils.Forces) {
	bb.forces = forces
}

func (bb *BaseBiker) UpdatePoints(pointsGained int) {
	bb.points += pointsGained
}

func (bb *BaseBiker) UpdateEnergyLevel(energyLevel float64) {
	bb.energyLevel += energyLevel
	if bb.energyLevel > 1.0 {
		bb.energyLevel = 1.0
	}
}

func (bb *BaseBiker) ToggleOnBike() {
	bb.onBike = !bb.onBike
}

func (bb *BaseBiker) ResetPoints() {
	bb.points = 0
}

func (bb *BaseBiker) SetRoundDirection(direction uuid.UUID) {
	bb.roundDecisions.Direction = direction
}

func (bb *BaseBiker) SetRoundForces(forces utils.Forces) {
	bb.roundDecisions.Forces = forces
}

func (bb *BaseBiker) SetRoundDidConform(didConform bool) {
	bb.roundDecisions.didConform = didConform
}

func (bb *BaseBiker) UpdateRegimeTrustValues(oneGini float64, someGini float64, manyGini float64) {

}

func (bb *BaseBiker) UpdateIterationIncome(resources float64) {
	bb.iterationIncome += resources
}

func (bb *BaseBiker) ResetIterationIncome() {
	bb.iterationIncome = 0.0
}

// ----- Helper Functions for the above -----

func (bb *BaseBiker) nearestLoot() uuid.UUID {

	// returns the nearest lootbox with respect to the agent's bike current position
	// in the MVP this is used to determine the pedalling forces as all agent will be
	// aiming to get to the closest lootbox by default

	currLocation := bb.GetLocation()
	shortestDist := math.MaxFloat64
	var nearestBox uuid.UUID
	var currDist float64
	for _, loot := range bb.gameState.GetLootBoxes() {
		x, y := loot.GetPosition().X, loot.GetPosition().Y
		currDist = math.Sqrt(math.Pow(currLocation.X-x, 2) + math.Pow(currLocation.Y-y, 2))
		if currDist < shortestDist {
			nearestBox = loot.GetID()
			shortestDist = currDist
		}
	}
	return nearestBox
}

func (bb *BaseBiker) GetGameState() IGameState {
	return bb.gameState
}
