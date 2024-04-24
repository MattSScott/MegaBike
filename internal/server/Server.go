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

// const LootBoxCount = BikerAgentCount * 2.5 // 2.5 lootboxes available per Agent
// const MegaBikeCount = 11                   // Megabikes should have 8 riders
// const BikerAgentCount = 56                 // 56 agents in total

type IBaseBikerServer interface {
	baseserver.IServer[objects.IBaseBiker]
	objects.IGameState

	Initialize(iterations int)                                                           // returns the awdi interface
	GetJoiningRequests([]uuid.UUID) map[uuid.UUID][]uuid.UUID                            // returns a map from bike id to the id of all agents trying to joing that bike
	GetRandomBikeId() uuid.UUID                                                          // gets the id of any random bike in the map
	RulerElection(agents []objects.IBaseBiker, governance utils.Governance) uuid.UUID    // runs the ruler election
	RunRulerAction(bike objects.IMegaBike) uuid.UUID                                     // gets the direction from the dictator
	RunDemocraticAction(bike objects.IMegaBike, weights map[uuid.UUID]float64) uuid.UUID // gets the direction in voting-based governances
	NewGameStateDump(iteration int) GameStateDump                                        // creates a new game state dump
	GetLeavingDecisions() []uuid.UUID                                                    // gets the list of agents that want to leave their bike
	HandleKickoutProcess() []uuid.UUID                                                   // handles the kickout process
	ProcessJoiningRequests(inLimbo []uuid.UUID)                                          // processes the joining requests
	RunActionProcess()                                                                   // runs the action (direction choice + pedalling) process for each bike
	AwdiCollisionCheck()                                                                 // checks for collisions between awdi and bikes
	AddAgentToBike(agent objects.IBaseBiker, bike objects.IMegaBike)                     // adds an agent to a bike (which also has some side effects on some server data structures)
	// FoundingInstitutions()                                                                                       // runs the founding institutions process
	GetWinningDirection(finalVotes map[uuid.UUID]voting.LootboxVoteMap, weights map[uuid.UUID]float64) uuid.UUID // gets the winning direction according to the selected voting process
	LootboxCheckAndDistributions()                                                                               // checks for collision between bike and lootbox and runs the distribution process
	ResetGameState()                                                                                             // resets game state (at the beginning of a new round)
	GetDeadAgents() map[uuid.UUID]objects.IBaseBiker                                                             // returns the map of dead agents
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

// func (s *Server) Initialize(iterations int) IBaseBikerServer {
// 	server := &Server{
// 		BaseServer:     *baseserver.CreateServer[objects.IBaseBiker](s.GetAgentGenerators(), iterations),
// 		lootBoxes:      make(map[uuid.UUID]objects.ILootBox),
// 		megaBikes:      make(map[uuid.UUID]objects.IMegaBike),
// 		megaBikeRiders: make(map[uuid.UUID]uuid.UUID),
// 		deadAgents:     make(map[uuid.UUID]objects.IBaseBiker),
// 		awdi:           objects.GetIAwdi(),
// 	}
// 	server.replenishLootBoxes()
// 	server.replenishMegaBikes()

// 	return server
// }

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

// when an agent dies it needs to be removed from its bike, the riders map and the agents map + it's added to the dead agents map
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

func (s *Server) GetDeadAgents() map[uuid.UUID]objects.IBaseBiker {
	return s.deadAgents
}

// func (s *Server) outputResults(gameStates [][]GameStateDump) {
// 	stats := CalculateStatistics(gameStates)

// 	lifeSpans := stats.Average.AgentLifetime

// 	avg := 0.0
// 	size := float64(len(lifeSpans))

// 	for _, val := range lifeSpans {
// 		avg += val
// 	}

// 	avg /= size
// 	fmt.Println(avg)

// 	f, err := os.Create("output.txt") // creating...
// 	if err != nil {
// 		fmt.Printf("error creating file: %v", err)
// 		return
// 	}
// 	defer f.Close()
// 	_, err = f.WriteString(fmt.Sprintf("%f", avg))
// 	if err != nil {
// 		fmt.Printf("error writing string: %v", err)
// 	}

// 	// statisticsJson, _ := json.MarshalIndent(stats.Average.AgentLifetime, "", "    ")
// 	// fmt.Println("Average Statistics:\n" + string(statisticsJson))
// }

// func (s *Server) outputResults(gameStates [][]GameStateDump) {
// 	statistics := CalculateStatistics(gameStates)

// 	statisticsJson, err := json.MarshalIndent(statistics.Average, "", "    ")
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Average Statistics:\n" + string(statisticsJson))

// 	file, err := os.Create("statistics.xlsx")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()
// 	if err := statistics.ToSpreadsheet().Write(file); err != nil {
// 		panic(err)
// 	}

// 	file, err = os.Create("game_dump.json")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()
// 	encoder := json.NewEncoder(file)
// 	encoder.SetIndent("", "    ")
// 	if err := encoder.Encode(gameStates); err != nil {
// 		panic(err)
// 	}
// }

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
	gameDumpPath := "/gameDumps/debug/"
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
