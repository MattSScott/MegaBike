package server

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"fmt"

	"github.com/google/uuid"
)

// the simulation loop represents 100 rounds
func (s *Server) RunSimLoop(rounds int, gameState *SimplifiedGameStateDump, iteration int) {

	// ----- 0. Gossip Phase -----
	s.RunMessagingSession()

	// ----- 1. Self-Selection Phase -----
	if iteration != 0 {
		s.RunBikeSwitch() 
		s.SetDestinationBikes()
	}


	// ----- 2. Action phase -----
	for _, bike := range s.megaBikes { 
		fmt.Println("agents on bike:", len(bike.GetAgents()))
		fmt.Println("governance system:", bike.GetGovernance())
		s.UpdateBikeRules(bike)
		s.PerformRoleAssignment(bike)
	}


	s.ResetGameState() 
	iterationDump := s.GenerateIterationDump()


	// ----- 3. Operation Phase: Run the n rounds within this iteration. Megabikes move around the world. -----
	for i := 0; i < rounds; i++ {
		s.RunRoundLoop(iterationDump, i)
	}




	
	// Code for recording game state:

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


// handles bikers leaving the bike, potential kick outs and the acceptance process (in this order)
func (s *Server) RunBikeSwitch() {

	inLimbo := make([]uuid.UUID, 0)

	// 1. Process agents who are leaving bikes (voluntarily or kicked off)

	// voluntary exits
	changeBike := s.GetLeavingDecisions()
	inLimbo = append(inLimbo, changeBike...)

	// forced exit (kicked out)
	kickedOff := s.HandleKickoutProcess()
	inLimbo = append(inLimbo, kickedOff...)

	// 2. Collect join requests from bikeless agents and process them
	s.ProcessJoiningRequests(inLimbo)

}

// remove all agents from bikes, respawn dead agents (if required), replenish energy (if required), reset points (if required)
// replenish environment objects
func (s *Server) ResetGameState() {

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

	// for _, bike := range s.GetMegaBikes() {
	// 	bike.SetRuler(uuid.Nil)
	// }

	// for _, agent := range s.GetAgentMap() {
	// 	agent.SetBike(uuid.Nil)
	// }

	s.replenishLootBoxes()
	s.replenishMegaBikes()
}


// assign roles to agents (i.e. assign representatives)
func (s *Server) PerformRoleAssignment(bike objects.IMegaBike) {
	governanceSystem := bike.GetGovernance()
	// if governance system is some form of monarchy or aristocracy, need representatives.
	if governanceSystem == utils.PerfectMonarchy || governanceSystem == utils.DegenerateMonarchy || governanceSystem == utils.PerfectAristocracy || governanceSystem == utils.DegenerateAristocracy {
		// run election process
		agentsOnBike := bike.GetAgents()
		reps := s.RepresentativeElection(agentsOnBike, governanceSystem)
		bike.SetRepresentatives(reps)
	}
}
