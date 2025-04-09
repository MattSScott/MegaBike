package objects

import (
	utils "SOMAS2023/internal/common/utils"
	voting "SOMAS2023/internal/common/voting"
	"math"
	"math/rand"

	baseAgent "github.com/MattSScott/basePlatformSOMAS/BaseAgent"
	"github.com/MattSScott/basePlatformSOMAS/messaging"
	"github.com/google/uuid"
)

type IBaseBiker interface {
	baseAgent.IAgent[IBaseBiker] // Embedding base agent

	// ----- Strategy Functions (to be overridden by a specific agent implementation) -----

	// Decision Making (all agents)

	DecideAction() BikerAction                                   // returns: int reflecting what the agent has decided to do this iteration (pedal the bike (0), or try to change bikes (1))
	DecideJoining(pendinAgents []uuid.UUID) map[uuid.UUID]bool   // returns: map of pending agents uuid -> {true, false} where true means accept and false means dont accept
	DecideChangeBike() uuid.UUID                                       // returns: uuid of the target bike if they want to change, otherwise just return current bike id (?)
	ProposeDirection() uuid.UUID                                 // returns: lootbox uuid to aim towards (the direction)
	ProposeDirectionFromSubset(map[uuid.UUID]ILootBox) uuid.UUID // returns: lootbox uuid to aim towards (the direction) from a subset of lootboxes
	ProposeNewRadius(float64) float64                            // returns: TODO
	DecideAllocation() voting.IdVoteMap                          // returns: map containing bikerID -> distribution (i.e. share of resources)
	VoteForKickout() map[uuid.UUID]int                           // returns: map of UUID -> {0,1} for an agent where 0 means 'don't kick' and 1 means 'do kick'
	DecideForce(direction uuid.UUID)                             // decides: the force the biker is going to pedal with
	HandleAgentUnalive(id uuid.UUID)                             // decides: how to handle a dead agent

	// Decision Making (extra representative functions)

	DecideDirectionMalevolently() uuid.UUID                                      // returns: uuid of lootbox to aim towards (i.e. the direction). decided in a selfless way. only called when the agent is perfect monarch / aristocrat
	DecideDirectionBenevolently() uuid.UUID                                      // returns: uuid of lootbox to aim towards (i.e. the direction). decided in a selfish way, only called when agent is degen monarch / aristocrat
	DecideKickOut() []uuid.UUID                                                  // returns: slice of agents to kick out
	DecideRepresentativeAllocation(governance utils.Governance) voting.IdVoteMap // returns: map containing bikerID -> distribution (i.e. share of resources) (only called by reps)

	// Message Handlers (self-explanatory)

	HandleKickoutMessage(msg KickoutAgentMessage)
	HandleReputationMessage(msg ReputationOfAgentMessage)
	HandleJoiningMessage(msg JoiningAgentMessage)
	HandleLootboxMessage(msg LootboxMessage)
	HandleGovernanceMessage(msg GovernanceMessage)
	HandleForcesMessage(msg ForcesMessage)
	HandleVoteGovernanceMessage(msg VoteGoveranceMessage)
	HandleVoteLootboxDirectionMessage(msg VoteLootboxDirectionMessage)
	HandleVoteRulerMessage(msg VoteRulerMessage)
	HandleVoteKickoutMessage(msg VoteKickoutMessage)
	HandleVoteAllocationMessage(msg VoteAllocationMessage)
	GetAllMessages([]IBaseBiker) []messaging.IMessage[IBaseBiker]


	// ----- Core Functions (do not need to override these, can inherit basebiker's) -----

	// Getters

	GetForces() utils.Forces        // returns: forces for current round
	GetColour() utils.Colour        // returns: the colour of the lootbox that the agent is currently seeking
	GetLocation() utils.Coordinates // returns: the agent's location
	GetBike() uuid.UUID             // returns: the uuid of the bike the agent is currently on
	GetEnergyLevel() float64        // returns: energy level of the agent
	GetPoints() int                 // returns: the points the agent currently has
	GetBikeStatus() bool            // returns: whether the biker is on a bike or not
	GetFellowBikers() []IBaseBiker  // returns: slice containing the bikers on our bike.

	// Setters

	SetBike(uuid.UUID)                       // sets: the megaBikeID. (this is either the id of the bike that the agent is on or the one that it's trying to join)
	SetForces(forces utils.Forces)           // sets: the force (called within DecideForce())
	UpdatePoints(pointGained int)            // increases: the points of an agent by pointsGained
	UpdateEnergyLevel(energyLevel float64)   // increases: the energy level of the agent by the allocated lootbox share or decreases by expended energy
	ToggleOnBike()                           // toggles: whether a user is on bike or not. called when removing or adding a biker on a bike
	ResetPoints()							 // resets: agents points to 0


	// ----- Part 3: Misc -----

	// base biker trust system (i think)

	GetReputation() map[uuid.UUID]float64 // get reputation value of all other agents
	QueryReputation(uuid.UUID) float64    // query for reputation value of specific agent with UUID
	SetReputation(uuid.UUID, float64)     // set reputation value of specific agent with UUID



	// ----- Part 4: Deprecated -----
	// DecideGovernance() utils.Governance
	// VoteDictator() voting.IdVoteMap
	// VoteLeader() voting.IdVoteMap
	// DecideWeights(action utils.Action) map[uuid.UUID]float64 // decide on weights for various actions (for the leader)
	//SetDeterministicColour(col utils.Colour) // allows for deterministic setting of colour
	//FinalDirectionVote(proposals map[uuid.UUID]uuid.UUID) voting.LootboxVoteMap // ** stage 3 of direction voting
}


type BikerAction int

const (
	Pedal BikerAction = iota
	ChangeBike
)

type BaseBiker struct {
	*baseAgent.BaseAgent[IBaseBiker]              // BaseBiker inherits functions from BaseAgent such as GetID(), GetAllMessages() and UpdateAgentInternalState()
	soughtColour                     utils.Colour // the colour of the lootbox that the agent is currently seeking
	onBike                           bool
	energyLevel                      float64 // float between 0 and 1
	points                           int
	forces                           utils.Forces
	megaBikeId                       uuid.UUID             // if they are not on a bike it will be 0
	gameState                        IGameState            // updated by the server at every round
	reputation                       map[uuid.UUID]float64 // record reputation for other agents in float
}


// ----- Strategy Functions (to be overridden by a specific agent implementation) -----

// in the MVP the biker's action defaults to pedaling (as it won't be able to change bikes)
// in future implementations this function will be overridden by the agent's specific strategy
// which will be used to determine whether to pedal or try to change bike
func (bb *BaseBiker) DecideAction() BikerAction {
	return Pedal
}

// an agent will have to rank the agents that are trying to join and that they will try to
func (bb *BaseBiker) DecideJoining(pendingAgents []uuid.UUID) map[uuid.UUID]bool {
	decision := make(map[uuid.UUID]bool)
	for _, agent := range pendingAgents {
		decision[agent] = true
	}
	return decision
}

// decide which bike to go to. the base agent chooses a random bike
func (bb *BaseBiker) DecideChangeBike() uuid.UUID {
	megaBikes := bb.gameState.GetMegaBikes()
	i, targetI := 0, rand.Intn(len(megaBikes))
	// Go doesn't have a sensible way to do this...
	for id := range megaBikes {
		if i == targetI {
			return id
		}
		i++
	}
	panic("no bikes")
}

// default implementation returns the id of the nearest lootbox
func (bb *BaseBiker) ProposeDirection() uuid.UUID {
	return bb.nearestLoot()
}

func (bb *BaseBiker) ProposeDirectionFromSubset(subset map[uuid.UUID]ILootBox) uuid.UUID {
	currLocation := bb.GetLocation()
	shortestDist := math.MaxFloat64
	var nearestBox uuid.UUID
	var currDist float64
	for _, loot := range subset {
		x, y := loot.GetPosition().X, loot.GetPosition().Y
		currDist = math.Sqrt(math.Pow(currLocation.X-x, 2) + math.Pow(currLocation.Y-y, 2))
		if currDist < shortestDist {
			nearestBox = loot.GetID()
			shortestDist = currDist
		}
	}
	return nearestBox
}

func (bb *BaseBiker) ProposeNewRadius(pRad float64) float64 {
	return pRad * 1.1
}

// through this function the agent submits their desired allocation of resources
// in the MVP each agent returns 1 whcih will cause the distribution to be equal across all of them
func (bb *BaseBiker) DecideAllocation() voting.IdVoteMap {
	bikeID := bb.GetBike()
	fellowBikers := bb.gameState.GetMegaBikes()[bikeID].GetAgents()
	distribution := make(voting.IdVoteMap)
	for _, agent := range fellowBikers {
		if agent.GetID() == bb.GetID() {
			distribution[agent.GetID()] = 1.0
		} else {
			distribution[agent.GetID()] = 0.0
		}
	}
	return distribution
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

// determine the forces (pedalling, breaking and turning)
// in the MVP the pedalling force will be 1, the breaking 0 and the tunring is determined by the
// location of the nearest lootbox
// the function is passed in the id of the voted lootbox and the default base bikers steer to that lootbox.
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

func (bb *BaseBiker) HandleAgentUnalive(id uuid.UUID) {}

// doesnt matter what this is as overwritten anyway
func (bb *BaseBiker) DecideDirectionMalevolently() uuid.UUID {
	nearest := bb.nearestLoot()
	return nearest
}

func (bb *BaseBiker) DecideDirectionBenevolently() uuid.UUID {
	nearest := bb.nearestLoot()
	return nearest
}

// only called when the agent is the dictator
func (bb *BaseBiker) DecideKickOut() []uuid.UUID {
	return (make([]uuid.UUID, 0))
}

// only called when the agent is the dictator
func (bb *BaseBiker) DecideRepresentativeAllocation(governance utils.Governance) voting.IdVoteMap {
	bikeID := bb.GetBike()
	fellowBikers := bb.gameState.GetMegaBikes()[bikeID].GetAgents()
	distribution := make(voting.IdVoteMap)
	equalDist := 1.0 / float64(len(fellowBikers))
	for _, agent := range fellowBikers {
		distribution[agent.GetID()] = equalDist
	}
	return distribution
}


func (bb *BaseBiker) HandleKickoutMessage(msg KickoutAgentMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// agentId := msg.AgentId
	// kickout := msg.Kickout
}

func (bb *BaseBiker) HandleReputationMessage(msg ReputationOfAgentMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// agentId := msg.AgentId
	// reputation := msg.Reputation
}

func (bb *BaseBiker) HandleJoiningMessage(msg JoiningAgentMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// agentId := msg.AgentId
	// bikeId := msg.BikeId
}

func (bb *BaseBiker) HandleLootboxMessage(msg LootboxMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// lootboxId := msg.LootboxId
}

func (bb *BaseBiker) HandleGovernanceMessage(msg GovernanceMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// bikeId := msg.BikeId
	// governanceId := msg.GovernanceId
}

func (bb *BaseBiker) HandleForcesMessage(msg ForcesMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// agentId := msg.AgentId
	// agentForces := msg.AgentForces
}

func (bb *BaseBiker) HandleVoteGovernanceMessage(msg VoteGoveranceMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// voteMap := msg.VoteMap
}

func (bb *BaseBiker) HandleVoteLootboxDirectionMessage(msg VoteLootboxDirectionMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// voteMap := msg.VoteMap
}

func (bb *BaseBiker) HandleVoteRulerMessage(msg VoteRulerMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// voteMap := msg.VoteMap
}

func (bb *BaseBiker) HandleVoteKickoutMessage(msg VoteKickoutMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// voteMap := msg.VoteMap
}

func (bb *BaseBiker) HandleVoteAllocationMessage(msg VoteAllocationMessage) {
	// Team's agent should implement logic for handling other biker messages that were sent to them.

	// sender := msg.BaseMessage.GetSender()
	// voteMap := msg.VoteMap
}

// This function updates all the messages for that agent i.e. both sending and receiving.
// And returns the new messages from other agents to your agent
func (bb *BaseBiker) GetAllMessages([]IBaseBiker) []messaging.IMessage[IBaseBiker] {
	// For team's agent add your own logic on chosing when your biker should send messages and which ones to send (return)
	wantToSendMsg := false
	if wantToSendMsg {
		reputationMsg := bb.CreateReputationMessage()
		kickoutMsg := bb.CreatekickoutMessage()
		lootboxMsg := bb.CreateLootboxMessage()
		joiningMsg := bb.CreateJoiningMessage()
		governceMsg := bb.CreateGoverenceMessage()
		forcesMsg := bb.CreateForcesMessage()
		voteGoveranceMessage := bb.CreateVoteGovernanceMessage()
		voteLootboxDirectionMessage := bb.CreateVoteLootboxDirectionMessage()
		voteRulerMessage := bb.CreateVoteRulerMessage()
		voteKickoutMessage := bb.CreateVotekickoutMessage()
		voteAllocationMessage := bb.CreateVoteAllocationMessage()
		return []messaging.IMessage[IBaseBiker]{reputationMsg, kickoutMsg, lootboxMsg, joiningMsg, governceMsg, forcesMsg, voteGoveranceMessage, voteLootboxDirectionMessage, voteRulerMessage, voteKickoutMessage, voteAllocationMessage}
	}
	return []messaging.IMessage[IBaseBiker]{}
}


// ----- Core Functions (do not need to override these, can inherit basebiker's) -----

func (bb *BaseBiker) GetForces() utils.Forces {
	return bb.forces
}

func (bb *BaseBiker) GetColour() utils.Colour {
	return bb.soughtColour
}

// the biker itself doesn't technically have a location (as it's on the map only when it's on a bike)
// in fact this function is only called when the biker needs to make a decision about the pedaling forces
func (bb *BaseBiker) GetLocation() utils.Coordinates {
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

// Returns the other agents on your bike :)
func (bb *BaseBiker) GetFellowBikers() []IBaseBiker {
	bikes := bb.gameState.GetMegaBikes()
	if _, ok := bikes[bb.GetBike()]; !ok {
		return []IBaseBiker{}
	}
	bike := bikes[bb.GetBike()]
	fellowBikers := bike.GetAgents()
	return fellowBikers
}

func (bb *BaseBiker) SetBike(bikeId uuid.UUID) {
	bb.megaBikeId = bikeId
}

func (bb *BaseBiker) SetForces(forces utils.Forces) {
	bb.forces = forces
}

// update the points at the end of a round
func (bb *BaseBiker) UpdatePoints(pointsGained int) {
	bb.points += pointsGained
}

// the function will be called by the server to:
// - reduce the energy level based on the force spent pedalling (energyLevel will be neg.ve)
// - increase the energy level after a lootbox has been looted (energyLevel will be pos.ve)
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


// ----- Base Biker's Trust System -----

// GetReputation map from agent, need to check if nil when call this function
func (bb *BaseBiker) GetReputation() map[uuid.UUID]float64 {
	return bb.reputation
}

// QueryReputation of specific agent with given ID, if there is no record for given agentID then return 0
func (bb *BaseBiker) QueryReputation(agentId uuid.UUID) float64 {
	if bb.reputation == nil {
		return 0
	}
	return bb.reputation[agentId]
}

func (bb *BaseBiker) SetReputation(agentId uuid.UUID, reputation float64) {
	if bb.reputation == nil {
		bb.reputation = make(map[uuid.UUID]float64)
	}
	bb.reputation[agentId] = reputation
}


// ----- Helper Functions and Utils -----


// returns the nearest lootbox with respect to the agent's bike current position
// in the MVP this is used to determine the pedalling forces as all agent will be
// aiming to get to the closest lootbox by default
func (bb *BaseBiker) nearestLoot() uuid.UUID {
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

// this function is going to be called by the server to instantiate bikers in the MVP
func GetIBaseBiker(totColours utils.Colour, bikeId uuid.UUID, gameState IGameState) IBaseBiker {
	return &BaseBiker{
		BaseAgent:    baseAgent.NewBaseAgent[IBaseBiker](),
		soughtColour: utils.GenerateRandomColour(),
		onBike:       true,
		energyLevel:  1.0,
		points:       0,
		gameState:    gameState,
	}
}

// this function will be used by GetTeamAgent to get the ref to the BaseBiker
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


// ----- Create message functions - not in the interface not sure why -----

func (bb *BaseBiker) CreatekickoutMessage() KickoutAgentMessage {
	// Currently this returns a default message which sends to all bikers on the biker agent's bike
	// For team's agent, add your own logic to communicate with other agents
	return KickoutAgentMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		AgentId:     uuid.Nil,
		Kickout:     false,
	}
}

func (bb *BaseBiker) CreateReputationMessage() ReputationOfAgentMessage {
	// Currently this returns a default message which sends to all bikers on the biker agent's bike
	// For team's agent, add your own logic to communicate with other agents
	return ReputationOfAgentMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		AgentId:     uuid.Nil,
		Reputation:  1.0,
	}
}

func (bb *BaseBiker) CreateJoiningMessage() JoiningAgentMessage {
	// Currently this returns a default message which sends to all bikers on the biker agent's bike
	// For team's agent, add your own logic to communicate with other agents
	return JoiningAgentMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		AgentId:     uuid.Nil,
		BikeId:      uuid.Nil,
	}
}
func (bb *BaseBiker) CreateLootboxMessage() LootboxMessage {
	// Currently this returns a default message which sends to all bikers on the biker agent's bike
	// For team's agent, add your own logic to communicate with other agents
	return LootboxMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		LootboxId:   uuid.Nil,
	}
}

func (bb *BaseBiker) CreateGoverenceMessage() GovernanceMessage {
	// Currently this returns a default message which sends to all bikers on the biker agent's bike
	// For team's agent, add your own logic to communicate with other agents
	return GovernanceMessage{
		BaseMessage:  messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		BikeId:       uuid.Nil,
		GovernanceId: 0,
	}
}

func (bb *BaseBiker) CreateForcesMessage() ForcesMessage {
	// Currently this returns a default message which sends to all bikers on the biker agent's bike
	// For team's agent, add your own logic to communicate with other agents
	return ForcesMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		AgentId:     uuid.Nil,
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

func (bb *BaseBiker) CreateVoteGovernanceMessage() VoteGoveranceMessage {
	// Currently this returns a default/meaningless message
	// For team's agent, add your own logic to communicate with other agents
	return VoteGoveranceMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		VoteMap:     make(voting.IdVoteMap),
	}
}

func (bb *BaseBiker) CreateVoteLootboxDirectionMessage() VoteLootboxDirectionMessage {
	// Currently this returns a default/meaningless message
	// For team's agent, add your own logic to communicate with other agents
	return VoteLootboxDirectionMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		VoteMap:     make(voting.IdVoteMap),
	}
}

func (bb *BaseBiker) CreateVoteRulerMessage() VoteRulerMessage {
	// Currently this returns a default/meaningless message
	// For team's agent, add your own logic to communicate with other agents
	return VoteRulerMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		VoteMap:     make(voting.IdVoteMap),
	}
}

func (bb *BaseBiker) CreateVotekickoutMessage() VoteKickoutMessage {
	// Currently this returns a default/meaningless message
	// For team's agent, add your own logic to communicate with other agents
	return VoteKickoutMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		VoteMap:     make(map[uuid.UUID]int),
	}
}

func (bb *BaseBiker) CreateVoteAllocationMessage() VoteAllocationMessage {
	// Currently this returns a default/meaningless message
	// For team's agent, add your own logic to communicate with other agents
	return VoteAllocationMessage{
		BaseMessage: messaging.CreateMessage[IBaseBiker](bb, bb.GetFellowBikers()),
		VoteMap:     make(voting.IdVoteMap),
	}
}




// ----- DEPRECATED FUNCTIONS -----

// func (bb *BaseBiker) DecideGovernance() utils.Governance {
// 	// Change behaviour here to return different governance
// 	return utils.PerfectDemocracy
// }

// // defaults to voting for first agent in the list
// func (bb *BaseBiker) VoteDictator() voting.IdVoteMap {
// 	votes := make(voting.IdVoteMap)
// 	fellowBikers := bb.GetFellowBikers()
// 	for i, fellowBiker := range fellowBikers {
// 		if i == 0 {
// 			votes[fellowBiker.GetID()] = 1.0
// 		} else {
// 			votes[fellowBiker.GetID()] = 0.0
// 		}
// 	}
// 	return votes
// }



// // defaults to voting for first agent in the list
// func (bb *BaseBiker) VoteLeader() voting.IdVoteMap {
// 	votes := make(voting.IdVoteMap)
// 	fellowBikers := bb.GetFellowBikers()
// 	for i, fellowBiker := range fellowBikers {
// 		if i == 0 {
// 			votes[fellowBiker.GetID()] = 1.0
// 		} else {
// 			votes[fellowBiker.GetID()] = 0.0
// 		}
// 	}
// 	return votes
// }

// // defaults to an equal distribution over all agents for all actions
// func (bb *BaseBiker) DecideWeights(action utils.Action) map[uuid.UUID]float64 {
// 	weights := make(map[uuid.UUID]float64)
// 	agents := bb.GetFellowBikers()
// 	for _, agent := range agents {
// 		weights[agent.GetID()] = 1.0
// 	}
// 	return weights
// }

// func (bb *BaseBiker) GetGroupID() int {
// 	return bb.GroupID
// }

// // this is called when a lootbox of the desidered colour has been looted in order to update the sought colour (this was literally never called)
// func (bb *BaseBiker) UpdateColour(totColours utils.Colour) {
// 	bb.soughtColour = utils.Colour(rand.Intn(int(totColours)))
// }


// what is the point of this. also never called
// func (bb *BaseBiker) SetDeterministicColour(col utils.Colour) {
// 	bb.soughtColour = col
// }


// this function will contain the agent's strategy on deciding which direction to go to
// // the default implementation returns an equal distribution over all options
// // this will also be tried as returning a rank of options
// func (bb *BaseBiker) FinalDirectionVote(proposals map[uuid.UUID]uuid.UUID) voting.LootboxVoteMap {
// 	votes := make(voting.LootboxVoteMap)
// 	totOptions := len(proposals)
// 	normalDist := 1.0 / float64(totOptions)
// 	for _, proposal := range proposals {
// 		if val, ok := votes[proposal]; ok {
// 			votes[proposal] = val + normalDist
// 		} else {
// 			votes[proposal] = normalDist
// 		}
// 	}
// 	return votes
// }