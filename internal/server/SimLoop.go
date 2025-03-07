package server

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"fmt"
	// "math/rand"
	"slices"

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
		s.UpdateBikeRules(bike)
		s.PerformRoleAssignment(bike)

		repString := ""
		for _, repID := range bike.GetRepresentatives() {
			repString += utils.TranslateToName(repID) + " "
		}

		fmt.Println("Bike", bike.GetGovernance(), "has", len(bike.GetAgents()), "agents with the following reps:", repString)
	}


	// Cleanup and boilerplate
	s.ResetGameState() 
	iterationDump := s.GenerateIterationDump()


	// ----- 3. Operation Phase: Run the n rounds within this iteration. Megabikes move around the world. -----
	for i := 0; i < rounds; i++ {
		s.RunRoundLoop(iterationDump, i)
	}
	
	// ----- Extra: Code for recording game state: -----

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


// handles bikers voluntarily leaving the bike / getting kicked out, followed by the requesting to join and acceptance process 
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

// returns slice of all agents that want to leave their bike in current iteration
func (s *Server) GetLeavingDecisions() []uuid.UUID {

	leavingAgents := make([]uuid.UUID, 0)

	// iterate over all agents to create a slice of all agents in the game who have decided to leave their bikes. remove them from their bike.
	for agentId, agent := range s.GetAgentMap() {
		if agent.GetBikeStatus() {

			agent.UpdateAgentInternalState()
			
			switch agent.DecideAction() {
			case objects.Pedal:
				continue
			case objects.ChangeBike:
				leavingAgents = append(leavingAgents, agentId)
				s.RemoveAgentFromBike(agent)

				// Note: 
				// the bike id is set to be the desired bike and onbike is set to false
				// so by looking at the values of onBike and megaBikeID it will be known
				// whether the agent is trying to join a bike (and which one)

				// the request is handled at the beginning of the next round, so the moving
				// will only be finalised then
			default:
				panic("agent decided invalid action")
			}
		}
	}

	// if representatives have left the bike, need to replace / handle that.
	for _, bike := range s.GetMegaBikes() {
		reps := bike.GetRepresentatives()
		for repIdx, repID := range reps {
			// if this rep is leaving...
			if len(bike.GetAgents()) !=0 && slices.Contains(leavingAgents, repID) {
				s.HandleDepartingRepresentative(bike, repIdx)
			}
		}
	}

	return leavingAgents
}

// returns a slice of all agents that are kicked from their bike in the current iteration.
// process differs based on governance
func (s *Server) HandleKickoutProcess() []uuid.UUID {

	allKicked := make([]uuid.UUID, 0)

	for _, bike := range s.GetMegaBikes() {
		agents := bike.GetAgents()

		if len(agents) > 1 {

			kickedAgents := make([]uuid.UUID, 0)

			switch bike.GetGovernance() {
			case utils.PerfectDemocracy, utils.DegenerateDemocracy:
				// agents vote for they want to kick out
				// behaviour of this function changes based on which type of dem. it is
				kickedAgents = bike.KickOutAgent()

			case utils.PerfectAristocracy, utils.DegenerateAristocracy:

				reps := bike.GetRepresentatives()
				agents := s.GetAgentMap()
				
				// iterate over the reps and find agents which are suggested to be kicked by all reps. (i.e. intersection)
				for i, repID := range reps {
					aristocrat := agents[repID]
					aristocratDecision := aristocrat.DecideKickOut()
					
					if i == 0 {
						// Use first aristocrat's decisions as starting point
						kickedAgents = aristocratDecision
						continue
					}
					
					// Create map from the current kickedAgents slice
					// if an agent is in this map, it is currently one which all reps share in common
					currentCommonMap := make(map[uuid.UUID]bool)
					for _, id := range kickedAgents {
						currentCommonMap[id] = true
					}
					
					// Filter to only keep UUIDs that are in current aristocrat's decision
					// generate a new common list by looking at all ids in the current reps slice, and seeing if that is in the current common map
					// if so, it is shared by all agents up to this point so can be added to the new common list.
					newCommonList := make([]uuid.UUID, 0)
					for _, id := range aristocratDecision {
						if currentCommonMap[id] {
							newCommonList = append(newCommonList, id)
						}
					}
					
					// Update kicked agents for next iteration
					kickedAgents = newCommonList
				}


			case utils.PerfectMonarchy, utils.DegenerateMonarchy:
				reps := bike.GetRepresentatives()
				monarch := s.GetAgentMap()[reps[0]]
				kickedAgents = monarch.DecideKickOut()
			}

			allKicked = append(allKicked, kickedAgents...)

			//iterate over kicked agents. remove them from bike, and attempt to replace them if they were a rep
			for _, kickedAgentID := range kickedAgents {

				s.RemoveAgentFromBike(s.GetAgentMap()[kickedAgentID])
				reps := bike.GetRepresentatives()

				if slices.Contains(reps, kickedAgentID){
						departingRepIdx := slices.Index(reps, kickedAgentID)
						s.HandleDepartingRepresentative(bike, departingRepIdx)
					}
			}
		}
	}

	return allKicked
}


// collect join request and process them, adding agents to bikes if they are accepted.
func (s *Server) ProcessJoiningRequests(inLimbo []uuid.UUID) {
	
	// returns a map of bikeID -> slice of agents that want to join it.
	bikeRequests := s.GetJoiningRequests(inLimbo)

	// iterate through each bike and handle the acceptance process for the pending agents wishing to join it.
	for bikeID, pendingAgents := range bikeRequests {

		bike := s.megaBikes[bikeID]
		agents := bike.GetAgents()
		reps := bike.GetRepresentatives()
		
		// if there are no agents on the target bike accept all of them (until all seats are filled)
		if len(agents) == 0 {
			// as iterating over a map is pseudo-random it's enough to stop when the capacity is reached
			// to ensure a fair (= random) selection in the case of an empty target bike
			for i, pendingAgentID := range pendingAgents {
				if i <= utils.BikersOnBike {
					acceptedAgent := s.GetAgentMap()[pendingAgentID]
					s.AddAgentToBike(acceptedAgent, bike)
				} else {
					break
				}
			}
		} else {

			acceptedRanked := make([]uuid.UUID, 0)

			// the acceptance process is different for each governance type
			gov := bike.GetGovernance()

			switch gov {

			case utils.PerfectDemocracy, utils.DegenerateDemocracy:

				// make map of weights of 1 for all agents on bike
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// maps each agents ID to a map of their decision for each pending agent
				responses := make(map[uuid.UUID]map[uuid.UUID]bool, len(agents)) 

				// fill out the map
				for _, agent := range agents {
					responses[agent.GetID()] = agent.DecideJoining(pendingAgents)
				}

				// get a slice of uuids of the pending agents who passed the threshold to be accepted, ranked by num of yes's
				acceptedRanked = voting.GetAcceptanceRanking(responses, weights, gov)


			case utils.PerfectAristocracy, utils.DegenerateAristocracy:
				// EXPERIMENTAL METHOD

				// make map of weights of 1 for aristocrats and 0 for non aristocrats.
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					if slices.Contains(reps, agent.GetID()) {
						weights[agent.GetID()] = 1.0
					} else {
						weights[agent.GetID()] = 0.0
					}
				}

				// maps each agents ID to a map of their decision for each pending agent
				responses := make(map[uuid.UUID]map[uuid.UUID]bool, len(agents)) 

				// fill out the map
				for _, agent := range agents {
					responses[agent.GetID()] = agent.DecideJoining(pendingAgents)
				}

				// get a slice of uuids, ranked by num of yes's.
				acceptedRanked = voting.GetAcceptanceRanking(responses, weights, gov)
				
			case utils.PerfectMonarchy, utils.DegenerateMonarchy:

				monarch := s.GetAgentMap()[bike.GetRepresentatives()[0]]
				
				//returns a map of pendingagentID->decision
				acceptedMap := monarch.DecideJoining(pendingAgents)

				for agentID, accepted := range acceptedMap {
					if accepted {
						// not actually ranked, just a slice of accepted agents
						acceptedRanked = append(acceptedRanked, agentID)
					}
				}
			}

			// run acceptance process
			totalSeatsFilled := len(agents)
			emptySpaces := utils.BikersOnBike - totalSeatsFilled

			// accept up to capacity
			for i := 0; i < min(emptySpaces, len(acceptedRanked)); i++ {
				accepted := acceptedRanked[i]
				acceptedAgent := s.GetAgentMap()[accepted]
				s.AddAgentToBike(acceptedAgent, bike)
			}
		}
	}
}

// if agent is not on a bike, set their target bike ready for the next iteration when they can try to join it.
func (s *Server) SetDestinationBikes() {
	for _, agent := range s.GetAgentMap() {
		if !agent.GetBikeStatus() {
			targetBike := agent.ChangeBike()
			if targetBike == uuid.Nil { // agent didn't specify bike
				continue
			}
			if _, ok := s.megaBikes[targetBike]; !ok {
				panic("agent requested a bike that doesn't exist")
			}
			agent.SetBike(targetBike)
		}
	}
}

// assign roles to agents (i.e. assign representatives)
func (s *Server) PerformRoleAssignment(bike objects.IMegaBike) {
	governanceSystem := bike.GetGovernance()
	// if governance system is some form of monarchy or aristocracy, need representatives.
	if governanceSystem == utils.PerfectMonarchy || governanceSystem == utils.DegenerateMonarchy || governanceSystem == utils.PerfectAristocracy || governanceSystem == utils.DegenerateAristocracy {
		// run selection process
		agentsOnBike := bike.GetAgents()
		reps := s.RepresentativeElection(agentsOnBike, governanceSystem)
		bike.SetRepresentatives(reps)
	}
}

// remove all agents from bikes, respawn dead agents (if required), replenish energy (if required), reset points (if required), replenish environmental objects
func (s *Server) ResetGameState() {

	// respawn people who died in previous iteration (conditional)
	if utils.RespawnEveryIteration && utils.ReplenishEnergyEveryIteration {
		for _, agent := range s.deadAgents {
			s.AddAgent(agent)
		}
	}

	// replenish energy (conditional)
	if utils.ReplenishEnergyEveryIteration {
		for _, agent := range s.GetAgentMap() {
			agent.UpdateEnergyLevel(1.0)
		}
	}

	// empty the dead agent map
	clear(s.deadAgents)

	// zero the points (conditional)
	if utils.ResetPointsEveryIteration {
		for _, agent := range s.GetAgentMap() {
			agent.ResetPoints()
		}
	}

	// for _, agent := range s.GetAgentMap() {
	// 	agent.SetBike(uuid.Nil)
	// }

	s.replenishLootBoxes()
	s.replenishMegaBikes()
}