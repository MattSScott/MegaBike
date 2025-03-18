package server

import (
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"encoding/json"
	"fmt"
	"os"

	baseserver "github.com/MattSScott/basePlatformSOMAS/BaseServer"
	"github.com/google/uuid"
)

type IBaseBikerServer interface {
	baseserver.IServer[objects.IBaseBiker]
	objects.IGameState

	Initialize(iterations int)                                                                                   // returns the awdi interface
	GetJoiningRequests([]uuid.UUID) map[uuid.UUID][]uuid.UUID                                                    // returns a map from bike id to the id of all agents trying to joing that bike
	GetRandomBikeId() uuid.UUID                                                                                  // gets the id of any random bike in the map
	RepresentativeSelection(agents []objects.IBaseBiker, governance utils.Governance) []uuid.UUID                // runs the representative election
	HandleDepartingRepresentative(bike objects.IMegaBike, repIdxToReplace int)                                   // replaces a representative when they die /
	RunRepresentativeAction(bike objects.IMegaBike) uuid.UUID                                                    // gets the direction from the dictator
	RunDemocraticAction(bike objects.IMegaBike) uuid.UUID                                                        // gets the direction in voting-based governances
	GetLeavingDecisions() []uuid.UUID                                                                            // gets the list of agents that want to leave their bike
	HandleKickoutProcess() []uuid.UUID                                                                           // handles the kickout process
	ProcessJoiningRequests(inLimbo []uuid.UUID)                                                                  // processes the joining requests
	RunDirectionDecisionProcess()                                                                                // runs the action (direction choice + pedalling) process for each bike
	AwdiCollisionCheck()                                                                                         // checks for collisions between awdi and bikes
	AddAgentToBike(agent objects.IBaseBiker, bike objects.IMegaBike)                                             // adds an agent to a bike (which also has some side effects on some server data structures)                                                                     // runs the founding institutions process
	GetWinningDirection(finalVotes map[uuid.UUID]voting.LootboxVoteMap, weights map[uuid.UUID]float64) uuid.UUID // gets the winning direction according to the selected voting process
	LootboxCheckAndDistributions()                                                                               // checks for collision between bike and lootbox and runs the distribution process
	ResetGameState()                                                                                             // resets game state (at the beginning of a new round)
	GetDeadAgents() map[uuid.UUID]objects.IBaseBiker                                                             // returns the map of dead agents
	// NewGameStateDump(iteration int) GameStateDump                                        					 // creates a new game state dump
	// FoundingInstitutions()
}

type Server struct {
	baseserver.BaseServer[objects.IBaseBiker]
	lootBoxes map[uuid.UUID]objects.ILootBox
	megaBikes map[uuid.UUID]objects.IMegaBike
	// megaBikeRiders is a mapping from Agent ID -> ID of the bike that they are riding
	// helps with efficiently managing ridership status
	megaBikeRiders map[uuid.UUID]uuid.UUID // maps riders to their bike
	awdi           objects.IAwdi
	deadAgents     map[uuid.UUID]objects.IBaseBiker // map of dead agents (used for respawning at the end of a round )
	// foundingChoices map[uuid.UUID]utils.Governance
	globalRuleCache *objects.GlobalRuleCache
}

func GenerateServer() IBaseBikerServer {
	return &Server{}
}

// Spawns everything in
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

func (s *Server) Start() {
	fmt.Printf("Server initialised with %d agents \n\n", len(s.GetAgentMap()))

	gameState := NewSimplifiedGameStateDump()

	// maps each agent to another map, containing how many times they have been on a bike with each agent
	reassocationMap := make(map[uuid.UUID]map[uuid.UUID]int)

	for i := 0; i < s.GetIterations(); i++ {
		fmt.Printf("Game Loop %d running... \n \n", i+1)
		s.RunSimLoop(utils.Rounds, gameState, i, reassocationMap)
		fmt.Printf("Game Loop %d completed.\n", i+1)
		fmt.Println(len(s.GetAgentMap()))
		if len(s.GetAgentMap()) == 0 {
			break
		}
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

// when an agent dies it needs to be removed from its bike, the megabikeriders map and the agents map + it's added to the dead agents map
func (s *Server) RemoveAgent(agent objects.IBaseBiker) {

	id := agent.GetID()
	// add agent to dead agent map
	s.deadAgents[id] = agent
	// remove agent from agent map
	s.BaseServer.RemoveAgent(agent)
	if bikeId, ok := s.megaBikeRiders[id]; ok {
		s.megaBikes[bikeId].RemoveAgent(id)
		delete(s.megaBikeRiders, id)
	}

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

	// set agent on desired bike
	if len(bike.GetAgents()) == 8 {
		return
	}

	bike.AddAgent(agent)
	agent.SetBike(bike.GetID())
	s.megaBikeRiders[agent.GetID()] = bike.GetID()
	if !agent.GetBikeStatus() {
		agent.ToggleOnBike()
	}
}

// remove an agent from its bike
func (s *Server) RemoveAgentFromBike(agent objects.IBaseBiker) {
	bike := s.megaBikes[agent.GetBike()]
	bike.RemoveAgent(agent.GetID())
	agent.ToggleOnBike()

	// get new destination for agent
	targetBike := agent.ChangeBike()
	if _, ok := s.megaBikes[targetBike]; !ok {
		panic("agent requested a bike that doesn't exist")
	}
	agent.SetBike(targetBike)

	delete(s.megaBikeRiders, agent.GetID())
}

// returns: map of dead agents, uuid->agent
func (s *Server) GetDeadAgents() map[uuid.UUID]objects.IBaseBiker {
	return s.deadAgents
}

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

// had to override to address the fact that agents only have access to the game dump
// version of agents, so if the recipients are set to be those it will panic as they
// can't call the handler functions
func (s *Server) RunMessagingSession() {
	agentArray := s.GenerateAgentArrayFromMap()

	for _, agent := range s.GetAgentMap() {
		allMessages := agent.GetAllMessages(agentArray)
		for _, msg := range allMessages {
			recipients := msg.GetRecipients()
			// make recipient list with actual agents
			usableRecipients := make([]objects.IBaseBiker, len(recipients))
			for i, recipient := range recipients {
				usableRecipients[i] = s.GetAgentMap()[recipient.GetID()]
			}
			for _, recip := range usableRecipients {
				if agent.GetID() == recip.GetID() {
					continue
				}
				msg.InvokeMessageHandler(recip)
			}
		}
	}
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
