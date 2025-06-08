package server

import (
	"MegabikeFYPVersion/internal/common/config"
	"MegabikeFYPVersion/internal/common/objects"
	"MegabikeFYPVersion/internal/common/utils"
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

	// Core Server Things
	Initialize(iterations int)

	// Iteration Level
	GenerateIterationDump() *SimplifiedIterationDump
	ResetGameState()
	RunVoluntaryReassociation(iterationDump *SimplifiedIterationDump)
	RemoveAllAgentsFromBikes()
	RandomlyAssignRepresentatives() map[uuid.UUID]objects.IBaseBiker
	ProcessAgentQueue(queuedAgents map[uuid.UUID]objects.IBaseBiker, iterationDump *SimplifiedIterationDump)
	AddAgentToBike(agent objects.IBaseBiker, bike objects.IMegaBike)
	CalculateGiniCoefficient(incomes []float64) float64

	// Round Level
	RunDirectionDecisionProcess()
	RunOneDirectionDecision(bike objects.IMegaBike) uuid.UUID
	RunSomeDirectionDecision(bike objects.IMegaBike) uuid.UUID
	RunManyDirectionDecision(bike objects.IMegaBike) uuid.UUID
	LootboxCheckAndDistributions()
	AwdiCollisionCheck()

	// Getters for testing
	GetDeadAgents() map[uuid.UUID]objects.IBaseBiker
	GetBikerAgentCount() int
	GetMegabikeCount() int
	GetLootboxCount() int
	GetRandomBikeId() uuid.UUID
}

type Server struct {
	baseserver.BaseServer[objects.IBaseBiker]
	config                config.Config
	lootBoxes             map[uuid.UUID]objects.ILootBox
	megaBikes             map[uuid.UUID]objects.IMegaBike
	megaBikeRiders        map[uuid.UUID]uuid.UUID // a mapping from Agent ID -> ID of the bike that they are riding
	awdi                  objects.IAwdi
	deadAgents            map[uuid.UUID]objects.IBaseBiker // map of dead agents (used for respawning at the end of a round )
	joiningBasedOnTrust   float64                          // recording how many joining decisions prioritised agent trust over regime trust
	totalJoiningDecisions float64                          // recording the total joining decisions made
}

// constructor for a zero-valued server object
func GenerateServer() IBaseBikerServer {
	return &Server{}
}

// setups up the game environment and spawns everything in
func (s *Server) Initialize(iterations int) {
	s.config = config.NewConfig()
	s.BaseServer = *baseserver.CreateServer(s.GetAgentGenerators(), iterations)
	s.lootBoxes = make(map[uuid.UUID]objects.ILootBox)
	s.megaBikes = make(map[uuid.UUID]objects.IMegaBike)
	s.megaBikeRiders = make(map[uuid.UUID]uuid.UUID)
	s.deadAgents = make(map[uuid.UUID]objects.IBaseBiker)
	s.awdi = objects.GetIAwdi()
	s.replenishLootBoxes()
	s.spawnInitialMegaBikes()
	s.awdi.InjectGameState(s)
}

// begins the game
func (s *Server) Start() {
	fmt.Printf("Server initialised with %d agents \n\n", len(s.GetAgentMap()))

	gameState := NewSimplifiedGameStateDump()

	// adding all the agent and bike uuids to the gamestatedump for easier processing of data in analysis
	var agentUUIDs, bikeUUIDS []uuid.UUID
	for agentid := range s.GetAgentMap() {
		agentUUIDs = append(agentUUIDs, agentid)
	}
	for bikeid := range s.GetMegaBikes() {
		bikeUUIDS = append(bikeUUIDS, bikeid)
	}
	gameState.AgentUUIDS = agentUUIDs
	gameState.BikeUUIDS = bikeUUIDS

	// maps each agent to another map, containing how many times they have been on a bike with each agent
	reassocationMap := make(map[uuid.UUID]map[uuid.UUID]int)

	for i := 0; i < s.GetIterations(); i++ {
		fmt.Printf("Game Loop %d running... \n \n", i+1)
		s.RunIterationLoop(utils.Rounds, gameState, i, reassocationMap)
		fmt.Printf("Game Loop %d completed.\n", i+1)
		fmt.Println(len(s.GetAgentMap()))
		if len(s.GetAgentMap()) == 0 {
			break
		}
	}

	percentOfDecisionsMadeBasedOnTrust := (s.joiningBasedOnTrust / s.totalJoiningDecisions) * 100
	fmt.Println(percentOfDecisionsMadeBasedOnTrust)

	// this stuff is experimental and possibly to remove but keeping it for now
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

// called when an agent dies
func (s *Server) RemoveAgent(agent objects.IBaseBiker) {

	id := agent.GetID()

	// 1. add agent to dead agent map
	s.deadAgents[id] = agent

	// 2. remove agent from main agent map
	s.BaseServer.RemoveAgent(agent)

	// if they are on a bike,
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

// agents get a chance to send their messages if they desire
func (s *Server) RunAgentMessagingSession(isIteration bool) {

	// create an array of agent objects
	agentArray := s.GenerateAgentArrayFromMap()

	for _, agent := range s.GetAgentMap() {

		// retrieve all the messages this agent wants to send (either round or iteration messages)
		allMessages := agent.GetAllRoundMessages(agentArray)
		if isIteration {
			// allMessages = agent.GetAllIterationMessages(agentArray)
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

// returns: map of dead agent uuid -> agent
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

// returns: awdi
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

// returns: the bikerAgentCount config variable
func (s *Server) GetBikerAgentCount() int {
	return s.config.BikerAgentCount
}

// returns: the megabikecount config variable
func (s *Server) GetMegabikeCount() int {
	return s.config.MegaBikeCount
}

// returns: the lootboxcount config variable.
func (s *Server) GetLootboxCount() int {
	return s.config.LootBoxCount
}

// writes the gamestatedump at the end of the game to a json file
func (s *Server) outputSimulationResult(dump SimplifiedGameStateDump) {

	relativePath, _ := os.Getwd()
	percentageGoodAgents := s.config.ProportionOfGoodAgents * 100
	gameDumpPath := "\\gameDumps\\brexperiments\\" + fmt.Sprint(percentageGoodAgents) + "\\"
	// gameDumpPath := "\\gameDumps\\miexperiments\\"
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

// calculates the lifespan of each agent from the gamestatedump
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
