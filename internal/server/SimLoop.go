package server

import (
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"fmt"
	"math/rand"
	"slices"
	"sort"

	"github.com/google/uuid"
)

// the simulation loop (i.e. an iteration) represents 100 rounds
func (s *Server) RunSimLoop(rounds int, gameState *SimplifiedGameStateDump, iteration int, reassocationMap map[uuid.UUID]map[uuid.UUID]int) {

	// record each regimes total collected resources up until this point
	currentPoolPerRegimeStart := make(map[utils.Governance]float64)
	for _, bike := range s.GetMegaBikes() {
		currentPoolPerRegimeStart[bike.GetGovernance()] += bike.GetCurrentPool()
	}

	// ----- 0. Gossip Phase -----
	s.RunAgentMessagingSession(true)

	// ----- 1. Self-Selection Phase -----

	// old version

	// if iteration != 0 {
	// 	// s.RunBikeSwitch()
	// 	// s.SetDestinationBikes()
	// 	s.RunVoluntaryReassociation()

	// 	// EXPERIMENTAL: record associations in the map
	// 	s.RecordAssociations(reassocationMap)
	// }

	// new version
	s.RunVoluntaryReassociation()
	s.RecordAssociations(reassocationMap)


	// ----- 2. Role-Assigning phase (old too) -----
	for _, bike := range s.megaBikes {
		s.UpdateBikeRules(bike)
		// s.PerformRoleAssignment(bike)

		repString := ""
		for _, repID := range bike.GetRepresentatives() {
			repString += utils.TranslateToName(repID) + " "
		}

		fmt.Println("Bike", bike.GetGovernance(), "has", len(bike.GetAgents()), "agents with the following reps:", repString)
	}

	// Admin
	s.ResetGameState()
	iterationDump := s.GenerateIterationDump()

	// ----- 3. Operation Phase: Run the n rounds within this iteration. Megabikes move around the world. -----
	for i := 0; i < rounds; i++ {
		s.RunRoundLoop(iterationDump, i)
	}

	// ----- Recording game state -----

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

	// ----- Let agents adjust their regime trust values -----

	// record each regimes total collected resources at the end of the regime
	currentPoolPerRegimeEnd := make(map[utils.Governance]float64)
	for _, bike := range s.GetMegaBikes() {
		currentPoolPerRegimeEnd[bike.GetGovernance()] += bike.GetCurrentPool()
	}

	// calculated the amount of resources each regime has collected this iteration
	// by doing total resources at end of iteration  - total resources at start of iteration
	resourcesGainedPerRegime := make(map[utils.Governance]float64)

	for regime, end := range currentPoolPerRegimeEnd {
		start := currentPoolPerRegimeStart[regime]
		resourcesGainedPerRegime[regime] = end - start
	}

	var regimeRankThisIteration []utils.Governance
	for gov := range resourcesGainedPerRegime {
		regimeRankThisIteration = append(regimeRankThisIteration, gov)
	}
	
	// Sort them by descending resource gain
	sort.Slice(regimeRankThisIteration, func(i, j int) bool {
		return resourcesGainedPerRegime[regimeRankThisIteration[i]] > resourcesGainedPerRegime[regimeRankThisIteration[j]]
	})

	// let the agents update their regime trust values based on how the regimes performed this iteration
	for _, agent := range s.GetAgentMap() {
		agent.UpdateRegimeTrustValues(regimeRankThisIteration)
	}


}

// ----- OLD SELF SELECTION PHASE AND ROLE ASSIGNMENT -----

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

// returns: slice of all agents that want to leave their bike in current iteration
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

				// the request is handled at the beginning of the next iteration, so the moving
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
			if len(bike.GetAgents()) != 0 && slices.Contains(leavingAgents, repID) {
				s.HandleDepartingRepresentative(bike, repIdx)
			}
		}
	}

	return leavingAgents
}

// returns: a slice of uuids of all agents that are kicked from their bike in the current iteration.
func (s *Server) HandleKickoutProcess() []uuid.UUID {

	allKicked := make([]uuid.UUID, 0)

	for _, bike := range s.GetMegaBikes() {
		agents := bike.GetAgents()

		if len(agents) > 1 {

			kickedAgents := make([]uuid.UUID, 0)

			switch bike.GetGovernance() {
			case utils.Many:
				// agents vote for they want to kick out
				kickedAgents = bike.KickOutAgent()

			case utils.Some:

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

			case utils.One:
				reps := bike.GetRepresentatives()
				monarch := s.GetAgentMap()[reps[0]]
				kickedAgents = monarch.DecideKickOut()
			}

			allKicked = append(allKicked, kickedAgents...)

			//iterate over kicked agents. remove them from bike, and attempt to replace them if they were a rep
			for _, kickedAgentID := range kickedAgents {

				s.RemoveAgentFromBike(s.GetAgentMap()[kickedAgentID])
				reps := bike.GetRepresentatives()

				if slices.Contains(reps, kickedAgentID) {
					departingRepIdx := slices.Index(reps, kickedAgentID)
					s.HandleDepartingRepresentative(bike, departingRepIdx)
				}
			}
		}
	}

	return allKicked
}

// collect join requests and process them, adding agents to bikes if they are accepted.
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

			case utils.Many:
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

			case utils.Some:
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

			case utils.One:

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

// returns: a map of megaBikeIDs-> slice of ids of all Bikers that are trying to join it
func (s *Server) GetJoiningRequests(inLimbo []uuid.UUID) map[uuid.UUID][]uuid.UUID {

	// iterate over all agents, if their onBike is false, put their agent id in the slice tied to their desired bike.

	bikeRequests := make(map[uuid.UUID][]uuid.UUID)

	for agentID, agent := range s.GetAgentMap() {
		// don't process joining requests of agents in first round of limbo (ie the ones that have just left the bike)
		if !agent.GetBikeStatus() && !slices.Contains(inLimbo, agentID) {
			bike := agent.GetBike()
			if bike == uuid.Nil {
				continue
			}
			if ids, ok := bikeRequests[bike]; ok {
				bikeRequests[bike] = append(ids, agentID)
			} else {
				bikeRequests[bike] = []uuid.UUID{agentID}
			}
		}
	}

	return bikeRequests
}

// if agent is not on a bike, set their target bike ready for the next iteration when they can try to join it.
func (s *Server) SetDestinationBikes() {
	for _, agent := range s.GetAgentMap() {
		if !agent.GetBikeStatus() {
			targetBike := agent.DecideChangeBike()
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

// assign representatives (old)
func (s *Server) PerformRoleAssignment(bike objects.IMegaBike) {
	governanceSystem := bike.GetGovernance()
	// if governance system is one or some, we need representatives.
	if governanceSystem == utils.One || governanceSystem == utils.Some {
		// run selection process
		agentsOnBike := bike.GetAgents()
		reps := s.RepresentativeSelection(agentsOnBike, governanceSystem)
		bike.SetRepresentatives(reps)
	}
}



// ----- NEW SELF SELECTION PHASE QUEUEING SYSTEM -----

// runs the voluntary association process
func (s *Server) RunVoluntaryReassociation() {

	// Step 1: Take all agents off their bikes if they are on one
	s.RemoveAllAgentsFromBikes()

	// Step 2: Randomly assign reps
	queuedAgents := s.RandomlyAssignRepresentatives()

	// Step 3: Go through the queued agents, asking them their top bike, then prompting that bike to make an acceptance decision.
	s.ProcessAgentQueue(queuedAgents)
}

// removes all agents from their bikes
func (s *Server) RemoveAllAgentsFromBikes() {
	for id, agent := range s.GetAgentMap() {
		if bikeId, ok := s.megaBikeRiders[id]; ok {
			// remove agent from bike
			s.megaBikes[bikeId].RemoveAgent(id)
			// delete them from the megabike riders map
			delete(s.megaBikeRiders, id)
			// toggle their onbike status
			agent.ToggleOnBike()
		}
	}
}

// goes through the agents and fills the one/some bikes with as many representatives as possible. returns the agents left in the queue
func (s *Server) RandomlyAssignRepresentatives() map[uuid.UUID]objects.IBaseBiker {

	agentMap := s.GetAgentMap()
	
	// Collect all agent IDs
	ids := make([]uuid.UUID, 0, len(agentMap))
	for id := range agentMap {
		ids = append(ids, id)
	}

	// Shuffle the IDs
	rand.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})


	repPointer := 0
	
	// Stratify bikes by regime
	var oneBikes, someBikes []objects.IMegaBike
	for _, bike := range s.GetMegaBikes() {
		switch bike.GetGovernance() {
		case utils.One:
			oneBikes = append(oneBikes, bike)
		case utils.Some:
			someBikes = append(someBikes, bike)
		}
	}

	// For all "one" bikes, try to assign a representative
	for _, bike := range oneBikes {
		if repPointer >= len(ids) {
			break // No more agents left
		}
		repID := ids[repPointer]
		repPointer++
	
		agent := agentMap[repID]
		s.AddAgentToBike(agent, bike)
		bike.SetRepresentatives([]uuid.UUID{repID})
	}

	// For all "some" bikes, assign up to 3 representatives
	for _, bike := range someBikes {
		remaining := len(ids) - repPointer
		if remaining <= 0 {
			break // No more agents left
		}

		numReps := 3
		if remaining < 3 {
			numReps = remaining // Use whatever is left
		}

		repIDs := ids[repPointer : repPointer+numReps]
		repPointer += numReps

		for _, id := range repIDs {
			s.AddAgentToBike(agentMap[id], bike)
		}
		bike.SetRepresentatives(repIDs)
	}


	// make a reduced version of the agentmap containing the rest of the agents that havent been chosen as reps
	queuedAgents := make(map[uuid.UUID]objects.IBaseBiker)

	// map of agentId->(are they a rep, yes/no)
	assignedReps := make(map[uuid.UUID]bool)
	for _, id := range ids[:repPointer] {
		assignedReps[id] = true
	}

	// add all the non-rep agents to the queue.
	for agentID, agent := range agentMap {
		if assignedReps[agentID] {
			continue
		} else {
			queuedAgents[agentID] = agent
		}
	}

	return queuedAgents
}

// takes the queued agents and lets them apply to bikes to be accepted / rejected
func (s *Server) ProcessAgentQueue(queuedAgents map[uuid.UUID]objects.IBaseBiker) {

	// start off with i = 0, i.e. they choose the bike at the top of their preference order. 
	// each time we cycle through the queue, try the next bike down in their preference order (e.g. i=1, i=2, ...) etc
	for i := 0; i < globals.MegaBikeCount; i++ {
		// go through the agents in the queue
		for agentNextInLineId, agentNextInLine := range queuedAgents {
			var accepted bool
			bikePreferenceOrder, bikeTrustHigherThanRegimeTrustMap := agentNextInLine.DecideBikePreferenceOrder()
			nextHighestBike := s.GetMegaBikes()[bikePreferenceOrder[i]]
			gov := nextHighestBike.GetGovernance()
			switch gov {
			case utils.Many: 
				if len(nextHighestBike.GetAgents()) == 0 {
					s.AddAgentToBike(agentNextInLine, nextHighestBike)
				} else {
					var decisions []bool
					for _, agent := range nextHighestBike.GetAgents() {
						decisions = append(decisions, agent.DecideJoiningOneAgent(agentNextInLineId))
					}
					acceptedCount := 0
					for _, decision := range decisions {
						if decision {
							acceptedCount++
						}
					}
					if acceptedCount > len(decisions)/2 {
						accepted = true
					} else {
						accepted = false
					}
				}
			case utils.Some:

				var decisions []bool
				for _, repId := range nextHighestBike.GetRepresentatives() {
					decisions = append(decisions, s.GetAgentMap()[repId].DecideJoiningOneAgent(agentNextInLineId))
				}
				acceptedCount := 0
				for _, decision := range decisions {
					if decision {
						acceptedCount++
					}
				}
				if acceptedCount > len(decisions)/2 {
					accepted = true
				} else {
					accepted = false
				}
			
			case utils.One:
				one := s.GetAgentMap()[nextHighestBike.GetRepresentatives()[0]]
				accepted = one.DecideJoiningOneAgent(agentNextInLineId)
			}


			totalSeatsFilled := len(nextHighestBike.GetAgents())
			emptySpaces := utils.BikersOnBike - totalSeatsFilled
			
			// if we can assign them (i.e. they are accepted and there is space left), add them to the bike and remove them from the queue
			if accepted && emptySpaces > 0 {
				s.AddAgentToBike(agentNextInLine, nextHighestBike)
				delete(queuedAgents, agentNextInLineId)

				// if they had bikeavgtrust > regime trust for the bike they are joining, incremement the joiningbasedontrust count
				if bikeTrustHigherThanRegimeTrustMap[nextHighestBike.GetID()] {
					s.joiningBasedOnTrust += 1
				}
			}
		}
	}
}




// respawn agents, reset and replenish game objects conditionally
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

	s.replenishLootBoxes()
	s.replenishMegaBikes()
}

// ----- EXPERIMENTAL -----
func (s *Server) RecordAssociations(reassocationMap map[uuid.UUID]map[uuid.UUID]int) {
	for agentID, agent := range s.GetAgentMap() {
		bikes := s.GetMegaBikes()

		// check if bike is in bikes map
		if _, ok := bikes[agent.GetBike()]; !ok {
			continue
		}

		// get a slice of the fellow bikers this agent has for this iteration
		bike := bikes[agent.GetBike()]
		fellowBikers := make([]objects.IBaseBiker, 0)
		for _, biker := range bike.GetAgents() {
			if biker.GetBikeStatus() {
				fellowBikers = append(fellowBikers, biker)
			}
		}

		// increment the count for each of these fellow bikers in the agent's map.
		for _, fellowBiker := range fellowBikers {
			// Make sure the inner map exists
			if _, exists := reassocationMap[agentID]; !exists {
				reassocationMap[agentID] = make(map[uuid.UUID]int)
			}

			// Increment the count
			reassocationMap[agentID][fellowBiker.GetID()]++
		}

	}
}
