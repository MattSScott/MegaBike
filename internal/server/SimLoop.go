package server

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"cmp"
	"fmt"
	"math"
	"slices"

	"github.com/google/uuid"
)

// the simulation loop represents a round
func (s *Server) RunSimLoop(iterations int, gameState *SimplifiedGameStateDump) {

	s.ResetGameState()
	s.FoundingInstitutions()

	iterationDump := s.GenerateIterationDump()

	// run this for n iterations
	// gameStates := []GameStateDump{s.NewGameStateDump(-1)}
	for i := 0; i < iterations; i++ {
		s.RunRoundLoop(iterationDump)
		// gameStates = append(gameStates, s.NewGameStateDump(i))
	}

	avgKicks := 0.0

	for bikeID, bike := range s.GetMegaBikes() {
		nKicks := bike.GetKickedOutCount()
		iterationDump.KickOffs[bikeID] = nKicks
		avgKicks += float64(nKicks)
	}

	nBikes := float64(len(s.GetMegaBikes()))

	iterationDump.AverageKickOffs = avgKicks / nBikes

	gameState.AddIterationToGameState(iterationDump)

	for _, bike := range s.GetMegaBikes() {
		bike.ResetKickedOutCount()
	}
}

// remove all agents from bikes, respawn dead agents (if required), replenish energy (if required), reset points (if required)
// replenish environment objects
func (s *Server) ResetGameState() {
	// kick everyone off bikes
	for _, agent := range s.GetAgentMap() {
		if agent.GetBike() != uuid.Nil {
			s.RemoveAgentFromBike(agent)
		} else if agent.GetBikeStatus() {
			agent.ToggleOnBike()
		}
	}

	// respawn people who died in previous round (conditional)
	if utils.RespawnEveryRound && utils.ReplenishEnergyEveryRound {
		for _, agent := range s.deadAgents {
			s.AddAgent(agent)
		}
	}

	// replenish energy (conditional)
	if utils.ReplenishEnergyEveryRound {
		for _, agent := range s.GetAgentMap() {
			agent.UpdateEnergyLevel(1.0)
		}
	}

	// empty the dead agent map
	clear(s.deadAgents)

	// zero the points (conditional)
	if utils.ResetPointsEveryRound {
		for _, agent := range s.GetAgentMap() {
			agent.ResetPoints()
		}
	}

	for _, bike := range s.GetMegaBikes() {
		bike.SetRuler(uuid.Nil)
	}

	for _, agent := range s.GetAgentMap() {
		agent.SetBike(uuid.Nil)
	}

	s.replenishLootBoxes()
	s.replenishMegaBikes()
}

// run the founding stage in which agents organise themselves on bikes
func (s *Server) FoundingInstitutions() {

	// run founding messaging session
	s.RunMessagingSession()

	// check which governance method is chosen for each biker
	s.foundingChoices = make(map[uuid.UUID]utils.Governance)
	for id, agent := range s.GetAgentMap() {
		// collect choice from each agent
		choice := agent.DecideGovernance()
		s.foundingChoices[id] = choice
	}

	// tally the choices
	// FoundingAllocations is a map of governance method to number of agents that want that governance method
	foundingTotals, _ := voting.TallyFoundingVotes(s.foundingChoices)

	// for each governance method, populate megabikes with the bikers who chose that governance method
	govBikes := make(map[utils.Governance][]uuid.UUID)
	bikesUsed := make([]uuid.UUID, 0)

	for governanceMethod, numBikers := range foundingTotals {
		megaBikesNeeded := int(math.Ceil(float64(numBikers) / float64(utils.BikersOnBike)))
		govBikes[governanceMethod] = make([]uuid.UUID, 0, megaBikesNeeded)
		// get bikes for this governance (enough to accommodate all bikers who chose this governance method)
		for i := 0; i < megaBikesNeeded; i++ {
			foundBike := false
			if len(bikesUsed) == len(s.megaBikes) {
				break
			}
			for !foundBike {
				bike := s.GetRandomBikeId()
				if !slices.Contains(bikesUsed, bike) {
					foundBike = true
					bikesUsed = append(bikesUsed, bike)
					govBikes[governanceMethod] = append(govBikes[governanceMethod], bike)

					// set the governance
					bikeObj := s.GetMegaBikes()[bike]
					bikeObj.SetGovernance(governanceMethod)
				}
			}
		}
	}

	for agent, governance := range s.foundingChoices {
		// randomly select a biker from the bikers who chose this governance method
		// add that biker to a megabike
		// if there are more bikers for a governance method than there are seats, then evenly distribute them across megabikes
		// select a bike with this governance method which has been assigned the lowest amount of bikers. If none available, stay in limbo
		bikesAvailable := govBikes[governance]
		if len(bikesAvailable) == 0 {
			continue
			// panic("not enough bikes to accommodate governance choices")
		}

		// Sort bikes from least to most full
		slices.SortFunc(bikesAvailable, func(a, b uuid.UUID) int {
			return cmp.Compare(len(s.megaBikes[a].GetAgents()), len(s.megaBikes[b].GetAgents()))
		})

		// get the first one of the sorted bikes
		chosenBike := bikesAvailable[0]
		// add agent to bike
		agentInt := s.GetAgentMap()[agent]
		agentInt.SetBike(chosenBike) // BUGGY!!! (setting bike here massively slows down iterations)
		agentInt.ToggleOnBike()
		s.AddAgentToBike(agentInt)
	}
	// run election process for Leadership and Dictatorship bikes
	for _, bike := range s.GetMegaBikes() {
		gov := bike.GetGovernance()
		agents := bike.GetAgents()
		if (gov == utils.Leadership || gov == utils.Dictatorship) && len(agents) != 0 {
			ruler := s.RulerElection(agents, gov)
			bike.SetRuler(ruler)
		}
	}

}

func (s *Server) Start() {
	fmt.Printf("Server initialised with %d agents \n\n", len(s.GetAgentMap()))
	// gameStates := make([][]GameStateDump, 0, s.GetIterations())

	gameState := NewSimplifiedGameStateDump()

	s.deadAgents = make(map[uuid.UUID]objects.IBaseBiker)
	for i := 0; i < s.GetIterations(); i++ {
		fmt.Printf("Game Loop %d running... \n \n", i)
		// gameStates = append(gameStates, s.RunSimLoop(utils.RoundIterations))
		s.RunSimLoop(utils.RoundIterations, gameState)
		s.RunMessagingSession()
		fmt.Printf("Game Loop %d completed.\n", i)
	}
	s.outputSimulationResult(*gameState)
}
