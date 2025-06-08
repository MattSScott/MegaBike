package server

import (
	"MegabikeFYPVersion/internal/common/objects"
	"MegabikeFYPVersion/internal/common/utils"
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// the iteration loop represents 100 rounds
func (s *Server) RunIterationLoop(rounds int, gameStateDump *SimplifiedGameStateDump, iteration int, reassocationMap map[uuid.UUID]map[uuid.UUID]int) {

	// ----- Setup -----

	// generate a zero-valued iteration dump ready to be filled in
	iterationDump := s.GenerateIterationDump()
	// reset certain game variables ready for a new iteration
	s.ResetGameState()

	// ----- 0. Gossip Phase -----

	if iteration != 0 {
		// s.RunAgentMessagingSession(true)
	}

	// ----- 1. Self-Selection Phase -----

	s.RunVoluntaryReassociation(iterationDump)
	// s.RecordAssociations(reassocationMap)
	s.DisplayAssociations()

	// ----- 2. Operation Phase: Run the n rounds within this iteration. Megabikes move around the world. -----

	for i := 0; i < rounds; i++ {
		s.RunRoundLoop(iterationDump, i)
	}

	// Publish Gini Coefficients to update agents' regime trust values ready for next iteration
	s.PublishGiniCoefficients()

	// Complete the rest of the iteration dump and add it to the game state dump
	s.FillOutIterationDump(gameStateDump, iterationDump)
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

// runs the voluntary association process
func (s *Server) RunVoluntaryReassociation(iterationDump *SimplifiedIterationDump) {

	// Step 1: Take all agents off their bikes if they are on one
	s.RemoveAllAgentsFromBikes()

	// Step 2: Randomly assign reps
	queuedAgents := s.RandomlyAssignRepresentatives()

	// Step 3: Go through the queued agents, asking them their top bike, then prompting that bike to make an acceptance decision.

	rand := false
	if rand {
		s.ProcessAgentQueueRandomly(queuedAgents, iterationDump)
	} else {
		s.ProcessAgentQueue(queuedAgents, iterationDump)
	}
}

// removes all agents from their bikes
func (s *Server) RemoveAllAgentsFromBikes() {
	for id, agent := range s.GetAgentMap() {
		if bikeId, ok := s.megaBikeRiders[id]; ok {
			// remove agent from their bikes agent list
			s.megaBikes[bikeId].RemoveAgent(id)
			// delete them from the megabike riders map
			delete(s.megaBikeRiders, id)
			// toggle their onbike status
			agent.ToggleOnBike()
			// set their bike id to nil
			agent.SetBike(uuid.Nil)
		}
	}
}

// returns: the agents left in the queue after filling the one/some bikes with as many representatives as possible.
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
someBikeLoop:
	for _, bike := range someBikes {
		var repIDs []uuid.UUID

		for i := 0; i < 3; i++ {
			// if we have ran out of reps, stop trying to assign
			if repPointer >= len(ids) {
				break someBikeLoop
			}

			newRepID := ids[repPointer]
			repIDs = append(repIDs, newRepID)
			repPointer++

			s.AddAgentToBike(agentMap[newRepID], bike)
			bike.SetRepresentatives(repIDs)
		}
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
		if _, ok := assignedReps[agentID]; !ok {
			queuedAgents[agentID] = agent
		}
	}

	return queuedAgents
}

// takes the queued agents and lets them apply to bikes to be accepted / rejected
func (s *Server) ProcessAgentQueue(queuedAgents map[uuid.UUID]objects.IBaseBiker, iterationDump *SimplifiedIterationDump) {

	// start off with i = 0, i.e. they choose the bike at the top of their preference order.
	// each time we cycle through the queue, try the next bike down in their preference order (e.g. i=1, i=2, ...) etc

	// CHANGED: now they try the top bike in their preference order every time
	for i := 0; i < s.config.MegaBikeCount; i++ {
		// go through the agents in the queue
		for agentNextInLineId, agentNextInLine := range queuedAgents {
			var accepted bool
			bikePreferenceOrder := agentNextInLine.DecideBikePreferenceOrder()
			nextHighestBike := s.GetMegaBikes()[bikePreferenceOrder[0].BikeId]
			gov := nextHighestBike.GetGovernance()
			switch gov {
			case utils.Many:
				accepted = s.GetManyJoiningDecision(nextHighestBike, agentNextInLine)
			case utils.Some:
				accepted = s.GetSomeJoiningDecision(nextHighestBike, agentNextInLine)
			case utils.One:
				accepted = s.GetOneJoiningDecision(nextHighestBike, agentNextInLine)
			}

			totalSeatsFilled := len(nextHighestBike.GetAgents())
			emptySpaces := utils.BikersOnBike - totalSeatsFilled

			// if we can assign them (i.e. they are accepted and there is space left), add them to the bike and remove them from the queue
			if accepted && emptySpaces > 0 {
				s.AddAgentToBike(agentNextInLine, nextHighestBike)
				delete(queuedAgents, agentNextInLineId)

				// increment the number of joining decisions
				s.totalJoiningDecisions += 1

				// if they had bikeavgtrust > regime trust for the bike they are joining, incremement the joiningbasedontrust count
				if bikePreferenceOrder[0].BikeTrustHigherThanRegimeTrust {
					s.joiningBasedOnTrust += 1
					iterationDump.JoiningDecisions[agentNextInLineId] = true // b > r
				} else {
					iterationDump.JoiningDecisions[agentNextInLineId] = false // b < r
				}
			}
		}
	}

	// --- New code: randomly assign any agents that never got a bike ---
    for agentID, agent := range queuedAgents {
        attempts := 0
        assigned := false
        // Try up to 100 times to find a bike with available capacity.
        for !assigned {
            randomBikeID := s.GetRandomBikeId()
            randomBike := s.GetMegaBikes()[randomBikeID]
            if len(randomBike.GetAgents()) < utils.BikersOnBike {
                s.AddAgentToBike(agent, randomBike)
                assigned = true
            }
            attempts++
        }
        // if !assigned {
            // fmt.Printf("Warning: Agent %v could not be assigned to a bike randomly after %d attempts\n", agentID, attempts)
        // }
        // Remove the agent from the queue regardless.
        delete(queuedAgents, agentID)
    }
}

// test function to assign agents randomly to bikes
func (s *Server) ProcessAgentQueueRandomly(queuedAgents map[uuid.UUID]objects.IBaseBiker, iterationDump *SimplifiedIterationDump) {
	for agentID, agent := range queuedAgents {
		randomBike := s.GetMegaBikes()[s.GetRandomBikeId()]
		for len(randomBike.GetAgents()) == 8 {
			randomBike = s.GetMegaBikes()[s.GetRandomBikeId()]
		}
		s.AddAgentToBike(agent, randomBike)
		delete(queuedAgents, agentID)
	}
}

// returns: the acceptance decision of a many bike for a given agent wanting to join
func (s *Server) GetManyJoiningDecision(nextHighestBike objects.IMegaBike, agentNextInLine objects.IBaseBiker) bool {
	if len(nextHighestBike.GetAgents()) == 0 {
		s.AddAgentToBike(agentNextInLine, nextHighestBike)
		return true
	} else {
		var decisions []bool
		for _, agent := range nextHighestBike.GetAgents() {
			decisions = append(decisions, agent.DecideJoiningOneAgent(agentNextInLine.GetID()))
		}
		acceptedCount := 0
		for _, decision := range decisions {
			if decision {
				acceptedCount++
			}
		}
		if acceptedCount > len(decisions)/2 {
			return true
		} else {
			return false
		}
	}
}

// returns: the acceptance decision of a some bike for a given agent wanting to join
func (s *Server) GetSomeJoiningDecision(nextHighestBike objects.IMegaBike, agentNextInLine objects.IBaseBiker) bool {
	var decisions []bool
	for _, repId := range nextHighestBike.GetRepresentatives() {
		decisions = append(decisions, s.GetAgentMap()[repId].DecideJoiningOneAgent(agentNextInLine.GetID()))
	}
	acceptedCount := 0
	for _, decision := range decisions {
		if decision {
			acceptedCount++
		}
	}
	if acceptedCount > len(decisions)/2 {
		return true
	} else {
		return false
	}
}

// returns: the acceptance decision of a one bike for a given agent wanting to join
func (s *Server) GetOneJoiningDecision(nextHighestBike objects.IMegaBike, agentNextInLine objects.IBaseBiker) bool {
	one := s.GetAgentMap()[nextHighestBike.GetRepresentatives()[0]]
	return one.DecideJoiningOneAgent(agentNextInLine.GetID())
}

// prints the results of the self-selection phase
func (s *Server) DisplayAssociations() {

	for bikeId, bike := range s.megaBikes {
		agentString := ""
		repString := ""
		for ag := range bike.GetAgents() {
			agentString += utils.TranslateToName(ag) + " "
		}
		for _, repID := range bike.GetRepresentatives() {
			repString += utils.TranslateToName(repID) + " "
		}

		fmt.Println("Bike", bikeId, "has governance", bike.GetGovernance(), ". It has the following agents and reps:")
		fmt.Println("Agents: ", agentString)
		fmt.Println("Reps: ", repString)
		fmt.Println(" ")
	}
}

// publishes the gini coefficients of each regime to the agents
func (s *Server) PublishGiniCoefficients() {

	var oneIncomes, someIncomes, manyIncomes []float64

	// make the income arrays for each regime type
	for _, agent := range s.GetAgentMap() {
		if !agent.GetBikeStatus() {
			continue
		}
		bike := agent.GetBike()
		governance := s.GetMegaBikes()[bike].GetGovernance()
		switch governance {
		case utils.One:
			oneIncomes = append(oneIncomes, agent.GetIterationIncome())
		case utils.Some:
			someIncomes = append(someIncomes, agent.GetIterationIncome())
		case utils.Many:
			manyIncomes = append(manyIncomes, agent.GetIterationIncome())
		}
	}

	// calculate the gini coefficients for each regime type
	oneGini := s.CalculateGiniCoefficient(oneIncomes)
	someGini := s.CalculateGiniCoefficient(someIncomes)
	manyGini := s.CalculateGiniCoefficient(manyIncomes)

	fmt.Println("Gini Coefficients of one, some, many: ", oneGini, someGini, manyGini)

	// publish these to the agents so they can update their regime trust. also reset their iteration income.
	for _, agent := range s.GetAgentMap() {
		agent.UpdateRegimeTrustValues(oneGini, someGini, manyGini)
		agent.ResetIterationIncome()
	}

}

// returns: the gini coefficient given a slice of incomes
func (s *Server) CalculateGiniCoefficient(incomes []float64) float64 {

	// this uses the discrete formula given on the wikipedia page for gini coefficient

	n := float64(len(incomes))
	if n == 0 {
		return 0
	}

	// Sort incomes in ascending order
	sort.Float64s(incomes)

	// Calculate the mean income
	var sum float64
	for _, income := range incomes {
		sum += income
	}
	mean := sum / n

	if mean == 0 {
		return 0
	}

	var totalDiff float64
	for i := 0; i < len(incomes); i++ {
		for j := 0; j < len(incomes); j++ {
			totalDiff += math.Abs(incomes[i] - incomes[j])
		}
	}

	gini := totalDiff / (2 * n * n * mean)

	return gini
}

// completes the iteration dump and add it to the gamestatedump
func (s *Server) FillOutIterationDump(gameStateDump *SimplifiedGameStateDump, iterationDump *SimplifiedIterationDump) {

	// Record average kickoffs
	avgKicks := 0.0

	for bikeID, bike := range s.GetMegaBikes() {
		nKicks := bike.GetKickedOutCount()
		iterationDump.KickOffs[bikeID] = nKicks
		avgKicks += float64(nKicks)
	}

	nBikes := float64(len(s.GetMegaBikes()))
	iterationDump.AverageKickOffs = avgKicks / nBikes

	for _, bike := range s.GetMegaBikes() {
		bike.ResetKickedOutCount()
	}

	// Record each agents' fellow bikers and their bikes for info theory calcs
	for agentID, agent := range s.GetAgentMap() {
		if !agent.GetBikeStatus() {
			continue
		}
		fellowBikerIds := []uuid.UUID{}
		for id := range agent.GetFellowBikers() {
			fellowBikerIds = append(fellowBikerIds, id)
		}
		iterationDump.AgentColleagues[agentID] = fellowBikerIds
		iterationDump.AgentBikes[agentID] = agent.GetBike()
	}

	gameStateDump.AddIterationToGameState(iterationDump)
}

// ----- Experimental -----

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
