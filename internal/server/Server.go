package server

import (
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	baseserver "github.com/MattSScott/basePlatformSOMAS/BaseServer"
	"github.com/google/uuid"
)

type IBaseBikerServer interface {
	baseserver.IServer[objects.IBaseBiker]
	objects.IGameState

	Initialize(iterations int)                                                                                   // spawns everything in
	GetJoiningRequests([]uuid.UUID) map[uuid.UUID][]uuid.UUID                                                    // returns: a map of megaBikeIDs-> slice of ids of all Bikers that are trying to join it
	GetRandomBikeId() uuid.UUID                                                                                  // gets the id of any random bike in the map
	RepresentativeSelection(agents []objects.IBaseBiker, governance utils.Governance) []uuid.UUID                // returns: slice of uuids of initially selected representatives
	HandleDepartingRepresentative(bike objects.IMegaBike, repIdxToReplace int)                                   // handles: when a representative leaves / dies / exits a bike. replace rep if possible, otherwise we just remove them
	RunOneDirectionDecision(bike objects.IMegaBike) uuid.UUID                                                    // returns: uuid of lootbox to aim toward (i.e. direction) for current round from the "one" rep
	RunSomeDirectionDecision(bike objects.IMegaBike) uuid.UUID                                                   // returns: uuid of lootbox to aim toward (i.e. direction) for current round from the "some" reps
	RunDemocraticDirectionDecision(bike objects.IMegaBike) uuid.UUID                                             // returns: uuid of lootbox to aim toward (i.e. direction) for current round from the agents
	GetLeavingDecisions() []uuid.UUID                                                                            // returns: slice of all agents that want to leave their bike in current iteration
	HandleKickoutProcess() []uuid.UUID                                                                           // returns: a slice of uuids of all agents that are kicked from their bike in the current iteration.
	ProcessJoiningRequests(inLimbo []uuid.UUID)                                                                  // collect join requests and process them, adding agents to bikes if they are accepted.
	RunDirectionDecisionProcess()                                                                                // run the process on deciding this round's direction for each megabike
	AwdiCollisionCheck()                                                                                         // check for deadly collisions with the awdi
	AddAgentToBike(agent objects.IBaseBiker, bike objects.IMegaBike)                                             // ensures that adding agents to a bike is atomic (ie no agent is added to a bike while still resulting as on another bike)                                                                // runs the founding institutions process
	LootboxCheckAndDistributions()                                                                               // if a bike has looted a box, run the distribution process according to the governance type
	ResetGameState()                                                                                             // respawn agents, reset and replenish game objects conditionally (each iteration)
	GetDeadAgents() map[uuid.UUID]objects.IBaseBiker                                                             // returns: map of dead agent uuid -> agent object
	GetWinningDirection(finalVotes map[uuid.UUID]voting.LootboxVoteMap, weights map[uuid.UUID]float64) uuid.UUID // returns: uuid of chosen lootbox from a set of votes and weights
	PerformRoleAssignment(bike objects.IMegaBike)                                                                // assign representatives
}

type Server struct {
	baseserver.BaseServer[objects.IBaseBiker]
	lootBoxes       map[uuid.UUID]objects.ILootBox
	megaBikes       map[uuid.UUID]objects.IMegaBike
	megaBikeRiders  map[uuid.UUID]uuid.UUID // a mapping from Agent ID -> ID of the bike that they are riding
	awdi            objects.IAwdi
	deadAgents      map[uuid.UUID]objects.IBaseBiker // map of dead agents (used for respawning at the end of a round )
	globalRuleCache *objects.GlobalRuleCache
}

func GenerateServer() IBaseBikerServer {
	return &Server{}
}

// spawns everything in
func (s *Server) Initialize(iterations int) {
	s.BaseServer = *baseserver.CreateServer[objects.IBaseBiker](s.GetAgentGenerators(), iterations)
	s.lootBoxes = make(map[uuid.UUID]objects.ILootBox)
	s.megaBikes = make(map[uuid.UUID]objects.IMegaBike)
	s.megaBikeRiders = make(map[uuid.UUID]uuid.UUID)
	s.deadAgents = make(map[uuid.UUID]objects.IBaseBiker)
	s.awdi = objects.GetIAwdi()
	s.globalRuleCache = objects.GenerateGlobalRuleCache()
	s.PopulateGlobalRuleCache()
	s.replenishLootBoxes()
	s.spawnInitialMegaBikesAndRiders()
	s.awdi.InjectGameState(s)
}

// begins the game
func (s *Server) Start() {
	fmt.Printf("Server initialised with %d agents \n\n", len(s.GetAgentMap()))

	gameState := NewSimplifiedGameStateDump()

	// maps each agent to another map, containing how many times they have been on a bike with each agent
	reassocationMap := make(map[uuid.UUID]map[uuid.UUID]int)

	numAgentsDevelopedCohesion := 0 
	agentMapCopy := make(map[uuid.UUID]objects.IBaseBiker)
	for agentID, agent := range s.GetAgentMap() {
		agentMapCopy[agentID] = agent
	}

	for i := 0; i < s.GetIterations(); i++ {
		fmt.Printf("Game Loop %d running... \n \n", i+1)
		s.RunSimLoop(utils.Rounds, gameState, i, reassocationMap)
		fmt.Printf("Game Loop %d completed.\n", i+1)
		fmt.Println(len(s.GetAgentMap()))
		if len(s.GetAgentMap()) == 0 {
			break
		}

		for agentID, agent := range agentMapCopy {
			if agent.GetDevelopedCohesion() {
				numAgentsDevelopedCohesion += 1
				delete(agentMapCopy, agentID)
			}
		}

		fmt.Println("Number of agents who are prioritising agent trust over regime trust", numAgentsDevelopedCohesion)
	}


	adjacencyMatrix := createAdjacencyMatrix(reassocationMap)

	file, err := os.Create("adjacency_matrix.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	json.NewEncoder(file).Encode(adjacencyMatrix)

	s.outputSimulationResult(*gameState)
}

// called when an agent dies. it needs to be added to the dead agents map, then removed from the main agents map, its bike, and the megabikeriders map
func (s *Server) RemoveAgent(agent objects.IBaseBiker) {

	id := agent.GetID()

	// 1. add agent to dead agent map
	s.deadAgents[id] = agent

	// 2. remove agent from main agent map
	s.BaseServer.RemoveAgent(agent)
	if bikeId, ok := s.megaBikeRiders[id]; ok {
		// 3. remove agent from bike
		s.megaBikes[bikeId].RemoveAgent(id)
		// 4. remove agent from megabikeriders map
		delete(s.megaBikeRiders, id)
	}

	// let all agents in the game run their own process for how to handle the death of this agent
	for _, agent := range s.GetAgentMap() {
		agent.HandleAgentUnalive(agent.GetID())
	}
}

// ensures that adding agents to a bike is atomic (ie no agent is added to a bike while still resulting as on another bike)
func (s *Server) AddAgentToBike(agent objects.IBaseBiker, bike objects.IMegaBike) {
	// Remove the agent from the old bike, if it was on one
	if oldBikeId, ok := s.megaBikeRiders[agent.GetID()]; ok {
		s.megaBikes[oldBikeId].RemoveAgent(agent.GetID())
		agent.ToggleOnBike()
	}

	// if bike is full, panic (shouldn't happen)
	if len(bike.GetAgents()) == utils.BikersOnBike {
		panic("trying to add agent to full bike")
	}

	bike.AddAgent(agent)
	agent.SetBike(bike.GetID())
	s.megaBikeRiders[agent.GetID()] = bike.GetID()
	if !agent.GetBikeStatus() {
		agent.ToggleOnBike()
	}
}

// remove an agent from its bike (e.g. when kicked off or voluntarily leaving)
func (s *Server) RemoveAgentFromBike(agent objects.IBaseBiker) {
	bike := s.megaBikes[agent.GetBike()]
	bike.RemoveAgent(agent.GetID())
	agent.ToggleOnBike()

	// get new destination for agent
	targetBike := agent.DecideChangeBike()
	if _, ok := s.megaBikes[targetBike]; !ok {
		panic("agent requested a bike that doesn't exist")
	}
	agent.SetBike(targetBike)

	delete(s.megaBikeRiders, agent.GetID())
}

// returns: map of dead agent uuid -> agent object
func (s *Server) GetDeadAgents() map[uuid.UUID]objects.IBaseBiker {
	return s.deadAgents
}

// returns: map of uuid->megabike
func (s *Server) GetMegaBikes() map[uuid.UUID]objects.IMegaBike {
	return s.megaBikes
}

// returns: map of uuid->lootbox
func (s *Server) GetLootBoxes() map[uuid.UUID]objects.ILootBox {
	return s.lootBoxes
}

// returns: awdi object
func (s *Server) GetAwdi() objects.IAwdi {
	return s.awdi
}

// returns: id of a random bike
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

// agents get a chance to send their messages if they desire
func (s *Server) RunAgentMessagingSession(isIteration bool) {

	// note - may be able to slim this down by cutting the usable recipients thing. left in however as may be important

	// note:  had to override to address the fact that agents only have access to the game dump
	// version of agents, so if the recipients are set to be those it will panic as they
	// can't call the handler functions

	// create an array of agent objects
	agentArray := s.GenerateAgentArrayFromMap()

	for _, agent := range s.GetAgentMap() {

		// retrieve all the messages this agent wants to send (either round or iteration messages)

		allMessages := agent.GetAllRoundMessages(agentArray)
		if isIteration {
			allMessages = agent.GetAllIterationMessages(agentArray)
		}

		// for each message ...
		for _, msg := range allMessages {
			recipients := msg.GetRecipients()

			usableRecipients := make([]objects.IBaseBiker, len(recipients))
			for i, recipient := range recipients {
				usableRecipients[i] = s.GetAgentMap()[recipient.GetID()]
			}

			// iterate over these recipients of the message and invoke their message handler.
			for _, recip := range usableRecipients {
				if agent.GetID() == recip.GetID() {
					continue
				}
				msg.InvokeMessageHandler(recip)
			}
		}
	}
}

// ----- Game State Dump Stuff -----

func lifespan(dump SimplifiedGameStateDump) map[uuid.UUID]int {
	result := make(map[uuid.UUID]int)
	for idx, gameState := range dump.Iterations {
		for _, round := range gameState.Rounds {
			for _, bike := range round.Bikes {
				for id := range bike.Agents {
					result[id] = idx + 1
				}
			}
		}
	}
	return result
}

func (s *Server) outputSimulationResult(dump SimplifiedGameStateDump) {

	relativePath, _ := os.Getwd()
	gameDumpPath := "\\gameDumps\\heterogenous\\" //change to homo/heterogenous depending on colour composition
	gameDumpHash := uuid.New().String()

	gameDumpFile := relativePath + gameDumpPath + gameDumpHash + ".json"

	file, err := os.Create(gameDumpFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(dump); err != nil {
		panic(err)
	}
	for id, span := range lifespan(dump) {
		fmt.Println(id, span)
	}
	fmt.Println(gameDumpFile)
}

// ----- Rule Stuff (possibly to remove) -----

func (s *Server) PopulateGlobalRuleCache() {
	// generate 100 rules split across N actions
	nActions := int(objects.MAX_ACTIONS)
	rulesPerAction := int(*globals.GlobalRuleCount / nActions)

	for i := 0; i < nActions; i++ {
		for j := 0; j < rulesPerAction; j++ {
			s.AddToGlobalRuleCache(objects.GenerateNullPassingRuleForAction(objects.Action(i)))
		}
	}
}

func (s *Server) ViewGlobalRuleCache() map[uuid.UUID]*objects.Rule {
	return s.globalRuleCache.ViewGlobalRuleSet()
}

func (s *Server) AddToGlobalRuleCache(rule *objects.Rule) {
	s.globalRuleCache.AddRuleToCache(rule)
}

// ----- EXPERIMENTAL -----

func createAdjacencyMatrix(reassocationMap map[uuid.UUID]map[uuid.UUID]int) [][]int {

	// Step 1: Extract unique agent IDs
	agentIDs := make([]uuid.UUID, 0, len(reassocationMap))
	for agentID := range reassocationMap {
		agentIDs = append(agentIDs, agentID)
	}

	// Step 2: Create a map to associate agent UUIDs with indices
	agentIndex := make(map[uuid.UUID]int)
	for i, agentID := range agentIDs {
		agentIndex[agentID] = i
	}

	// Step 3: Initialize the adjacency matrix (with all zeros initially)
	matrixSize := len(agentIDs)
	adjacencyMatrix := make([][]int, matrixSize)
	for i := range adjacencyMatrix {
		adjacencyMatrix[i] = make([]int, matrixSize)
	}

	// Step 4: Populate the adjacency matrix with the re-association data
	for agentID, associations := range reassocationMap {
		for otherAgentID, count := range associations {
			// Get indices for agent and other agent
			i := agentIndex[agentID]
			j := agentIndex[otherAgentID]
			adjacencyMatrix[i][j] = count
		}
	}

	return adjacencyMatrix
}
