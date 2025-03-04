package server

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"fmt"
	"math/rand"

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


	// cleanup and admin
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

// get list of agents that want to leave their bike in current iteration
func (s *Server) GetLeavingDecisions() []uuid.UUID {

	leavingAgents := make([]uuid.UUID, 0)

	for agentId, agent := range s.GetAgentMap() {
		if agent.GetBikeStatus() {

			agent.UpdateAgentInternalState() // is this necessary
			
			switch agent.DecideAction() {
			case objects.Pedal:
				continue
			case objects.ChangeBike:
				// the bike id is set to be the desired bike and onbike is set to false
				// so by looking at the values of onBike and megaBikeID it will be known
				// whether the agent is trying to join a bike (and which one)

				// the request is handled at the beginning of the next round, so the moving
				// will only be finalised then
				leavingAgents = append(leavingAgents, agentId)
				s.RemoveAgentFromBike(agent)
			default:
				panic("agent decided invalid action")
			}
		}
	}

	// if representatives have left the bike, need to replace them 
	for _, bike := range s.GetMegaBikes() {
		reps := bike.GetRepresentatives()
		for departingRepIdx, departingRepID := range reps {
			// if this rep is leaving, replace them
			if len(bike.GetAgents()) !=0 && slices.Contains(leavingAgents, departingRepID) {
				fmt.Println("rep voluntarily left bike with ", len(bike.GetAgents()), "agents")
				fmt.Println("reps before replacement", len(reps))
				s.HandleDepartingRepresentative(bike, departingRepIdx)
				fmt.Println("reps after replacement", len(reps))
			}
		}
	}

	return leavingAgents
}

// handles the kick out process according to each bike's governance
func (s *Server) HandleKickoutProcess() []uuid.UUID {

	allKicked := make([]uuid.UUID, 0)
	for _, bike := range s.GetMegaBikes() {
		agents := bike.GetAgents()

		if len(agents) > 1 {

			kickedAgents := make([]uuid.UUID, 0)

			switch bike.GetGovernance() {
			case utils.PerfectDemocracy, utils.DegenerateDemocracy:
				// make map of weights of 1 for all agents on bike (as they all have the same voting power)
				agents := bike.GetAgents()
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get which agents are getting kicked out
				kickedAgents = bike.KickOutAgent(weights)

			case utils.PerfectAristocracy, utils.DegenerateAristocracy:
				// for now a random aristocrat decides, todo: aggregrate them.
				reps := bike.GetRepresentatives()

				randIndex := rand.Intn(len(reps))
				chosenAristocrat := s.GetAgentMap()[reps[randIndex]]
				kickedAgents = chosenAristocrat.DecideKickOut()

			case utils.PerfectMonarchy, utils.DegenerateMonarchy:
				reps := bike.GetRepresentatives()

				// both perfect and degen kick out min social capital. no need to differentiate there.
				monarch := s.GetAgentMap()[reps[0]]
				kickedAgents = monarch.DecideKickOut()

			}

			// perform kickout and replace rep if one is kicked out
			allKicked = append(allKicked, kickedAgents...)
			for _, kickedAgentID := range kickedAgents {

				// remove them
				s.RemoveAgentFromBike(s.GetAgentMap()[kickedAgentID])
				reps := bike.GetRepresentatives()

				// if a representative was kicked out will need to select a new one
				if slices.Contains(reps, kickedAgentID){
						departingRepIdx := slices.Index(reps, kickedAgentID)
						s.HandleDepartingRepresentative(bike, departingRepIdx)
					}
			}
		}
	}

	return allKicked
}

// dispatch joining requests to the bikes of competence and move bikers from limbo to their desired bike subject to the acceptance process outcome
func (s *Server) ProcessJoiningRequests(inLimbo []uuid.UUID) {

	// -------------------------- PROCESS JOINING REQUESTS -------------------------
	// 1. group agents that have onBike = false by the bike they are trying to join
	bikeRequests := s.GetJoiningRequests(inLimbo)
	// panic(s.megaBikes)

	// 2. pass to agents on each of the desired bikes a list of all agents trying to join
	for bikeID, pendingAgents := range bikeRequests {
		bike := s.megaBikes[bikeID]
		agents := bike.GetAgents()
		// fmt.Println(len(agents))
		// if there are no agents on the target bike accept all of them (until all seats are filled)
		if len(agents) == 0 {
			// as iterating over a map is pseudo-random it's enough to stop whrn the capacity is reached
			// to ensure a fair (= random) selection in the case of an empty target bike
			for i, pendingAgent := range pendingAgents {
				if i <= utils.BikersOnBike {
					acceptedAgent := s.GetAgentMap()[pendingAgent]
					s.AddAgentToBike(acceptedAgent, bike)
				} else {
					break
				}
			}
		} else {
			acceptedRanked := make([]uuid.UUID, 0)

			// the acceptance process is different for each governance type

			// TODO: customise these for each type of governance
			switch bike.GetGovernance() {

			case utils.PerfectDemocracy, utils.DegenerateDemocracy, utils.PerfectAristocracy, utils.DegenerateAristocracy:
				// make map of weights of 1 for all agents on bike
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get approval votes from each agent
				responses := make(map[uuid.UUID]map[uuid.UUID]bool, len(agents)) // list containing all the agents' ranking
				for _, agent := range agents {
					responses[agent.GetID()] = agent.DecideJoining(pendingAgents)
				}

				// accept agents based on the response outcome (only capacity-n bikers can be accepted)
				acceptedRanked = voting.GetAcceptanceRanking(responses, weights)
				
			case utils.PerfectMonarchy, utils.DegenerateMonarchy:
				monarch := s.GetAgentMap()[bike.GetRepresentatives()[0]]
				acceptedRankedMap := monarch.DecideJoining(pendingAgents)
				for agentID, accepted := range acceptedRankedMap {
					if accepted {
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

// still don't understand the point of this function when we have runbikeswitch
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