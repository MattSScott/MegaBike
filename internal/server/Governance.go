package server

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"

	"github.com/google/uuid"
	"math/rand"
	"slices"
)

// returns: slice of uuids of initially selected representatives
func (s *Server) RepresentativeSelection(agentsOnBike []objects.IBaseBiker, governance utils.Governance) []uuid.UUID {

	// select representatives for the bike. this is not a democratic process.
	// each bike is fixed with a governance style and a representatives at the beginning of each iteration for the 'some' bike and 'one' bike.

	if len(agentsOnBike) != 0 {
		switch governance {
		case utils.Some:
			reps := make([]uuid.UUID, 0)
			var chosenAgentIndices []int
			if len(agentsOnBike) > 3 {
				chosenAgentIndices = rand.Perm(len(agentsOnBike))[0:3]
			} else {
				chosenAgentIndices = rand.Perm(len(agentsOnBike))
			}
			for _, v := range chosenAgentIndices {
				reps = append(reps, agentsOnBike[v].GetID())
			}
			return reps
		case utils.One:
			reps := make([]uuid.UUID, 0)
			chosenAgentIndex := rand.Intn(len(agentsOnBike))
			reps = append(reps, agentsOnBike[chosenAgentIndex].GetID())
			return reps
		default:
			panic("trying to run a representative election on a bike without incorrect governance style")
		}
	} else {
		return []uuid.UUID{}
	}

}

// returns: uuid of lootbox to aim toward (i.e. direction) for current round from the representatives
func (s *Server) RunRepresentativeDirectionDecision(bike objects.IMegaBike) uuid.UUID {
	agents := s.GetAgentMap()
	governance := bike.GetGovernance()
	reps := bike.GetRepresentatives()
	var direction uuid.UUID

	// decide differently based on each governance...
	switch governance {
	case utils.Some:

		suggestedDirections := make([]uuid.UUID, 0, len(reps))
		countsPerDirection := make(map[uuid.UUID]int)

		// create a slice of their suggested directions
		for _, repID := range reps {
			decidedDirection := agents[repID].DecideDirection()
			suggestedDirections = append(suggestedDirections, decidedDirection)
			agents[repID].SetRoundDirection(decidedDirection)
		}

		maxCounts := 0

		// find the most voted for direction
		for _, lootbox := range suggestedDirections {
			countsPerDirection[lootbox] += 1
			if countsPerDirection[lootbox] > maxCounts {
				maxCounts = countsPerDirection[lootbox]
				direction = lootbox
			}
		}

		return direction

	case utils.One:
		monarch := agents[reps[0]]
		direction = monarch.DecideDirection()
		monarch.SetRoundDirection(direction)
		return direction
	default:
		panic("trying to run representative action in a non-representative governance")
	}
}

// returns: uuid of lootbox to aim toward (i.e. direction) for current round from the agents
func (s *Server) RunDemocraticDirectionDecision(bike objects.IMegaBike) uuid.UUID {

	agents := bike.GetAgents()
	governance := bike.GetGovernance()
	proposedDirections := make(map[uuid.UUID]uuid.UUID) // maps agent id to their proposed lootbox id
	validLootboxes := s.PruneLootboxes(bike)

	// Fill in the proposed directions map by asking each agent to propose a direction.
	for _, agent := range agents {
		if agent.GetBikeStatus() {
			proposedDirection := agent.ProposeDirectionFromSubset(validLootboxes)
			agent.SetRoundDirection(proposedDirection)

			if proposedDirection == uuid.Nil {
				continue
			}
			if _, ok := s.lootBoxes[proposedDirection]; !ok {
				panic("agent proposed a non-existent lootbox")
			}
			proposedDirections[agent.GetID()] = proposedDirection
		}
	}

	if len(proposedDirections) == 0 {
		return uuid.Nil
	}

	if governance == utils.Many {

		directions := []uuid.UUID{}
		for _, direction := range proposedDirections {
			directions = append(directions, direction)
		}

		// Use a map to count occurrences of each UUID
		counts := make(map[uuid.UUID]int)

		// Count each UUID
		for _, lootbox := range directions {
			counts[lootbox]++
		}

		// Check if any lootbox UUID is voted for by a majority. If so, then return this lootbox.
		threshold := len(directions) / 2
		for lootbox, count := range counts {
			if count > threshold {
				return lootbox
			}
		}

		return uuid.Nil
	} else {
		panic("tring to run a democratic action in a non-democracy")
	}
}

// handles: when a representative leaves / dies / exits a bike. replace rep if possible, otherwise we just remove them
func (s *Server) HandleDepartingRepresentative(bike objects.IMegaBike, repIdxToReplace int) {

	gov := bike.GetGovernance()

	// first attempt to replace them
	if ((gov == utils.Some) && len(bike.GetAgents()) >= 3) || (gov == utils.One && len(bike.GetAgents()) >= 1) {

		reps := bike.GetRepresentatives()
		agentsOnBike := bike.GetAgents()

		// randomly choose an agent to be a rep. if agent is already rep, choose another one.
		replacementRepID := agentsOnBike[rand.Intn(len(agentsOnBike))].GetID()
		for slices.Contains(reps, replacementRepID) {
			replacementRepID = agentsOnBike[rand.Intn(len(bike.GetAgents()))].GetID()
		}

		// replace the old rep with a new rep
		reps[repIdxToReplace] = replacementRepID
		bike.SetRepresentatives(reps)
	} else {
		// otherwise we just remove them from the rep list
		reps := bike.GetRepresentatives()
		reps = slices.Delete(reps, repIdxToReplace, repIdxToReplace+1)
		bike.SetRepresentatives(reps)

	}
}

// returns: reduced map of uuid->lootbox based on the rules of the given megabike
func (s *Server) PruneLootboxes(bike objects.IMegaBike) map[uuid.UUID]objects.ILootBox {
	relevantRules := bike.GetActiveRulesForAction(objects.Lootbox)

	validLootboxes := make(map[uuid.UUID]objects.ILootBox, len(s.lootBoxes))

	for id, lb := range s.lootBoxes {
		validLootboxes[id] = lb
	}

	for _, r := range relevantRules {
		for id, l := range s.lootBoxes {
			if _, ok := validLootboxes[id]; !ok {
				continue
			}
			if !r.EvaluateLootboxRule(bike, l) {
				delete(validLootboxes, id)
			}
		}
	}

	return validLootboxes
}

// updates: the bike rules
func (s *Server) UpdateBikeRules(bike objects.IMegaBike) {
	tRad := 0.0
	nAg := 0.0

	rule := bike.GetActiveRulesForAction(objects.Lootbox)[0]
	pRad := rule.GetRuleMatrix()[0][1]

	// fmt.Println(rule)

	for _, agent := range bike.GetAgents() {
		if agent.GetBikeStatus() {
			tRad += agent.ProposeNewRadius(pRad)
			nAg += 1
		}
	}

	if nAg == 0 {
		return
	}

	tRad /= nAg

	newRuleMatrix := [][]float64{{1, tRad}}
	rule.UpdateRuleMatrix(newRuleMatrix)

	// nRule := bike.GetActiveRulesForAction(objects.Lootbox)[0]
	// fmt.Println(nRule.GetRuleMatrix())
}

// ----- Currently unused -----

// returns: uuid of chosen lootbox from a set of votes and weights
func (s *Server) GetWinningDirection(finalVotes map[uuid.UUID]voting.LootboxVoteMap, weights map[uuid.UUID]float64) uuid.UUID {
	// this allows to get a slice of the interface from that of the specific type
	// this way we can substitute agent.FInalDirectionVote with another function that returns
	// another type of voting type which still implements INormaliseVoteMap
	IfinalVotes := make(map[uuid.UUID]voting.IVoter)
	for i, v := range finalVotes {
		IfinalVotes[i] = v
	}

	return voting.WinnerFromDist(IfinalVotes, weights)
}

// ----- Legacy code to keep -----

// legacy: different behaviours  of rundemodratcidirectiondecision depending on the kind of democracy
// if governance == utils.PerfectDemocracy {
// 	// look for consensus
// 	directions := []uuid.UUID{}
// 	for _, direction := range proposedDirections {
// 		directions = append(directions, direction)
// 	}

// 	consensusReached := true
// 	for i := 1; i < len(directions); i++ {
// 		if directions[i] != directions[0] {
// 			consensusReached = false
// 			break
// 		}
// 	}

// 	if consensusReached {
// 		return directions[0] // could be any element of the slice they are all the same
// 	} else {
// 		return uuid.Nil // assuming this means 'don't move'
// 	}

// old more complex voting system for aggregating all agents ranked preferences
// kept here as may be useful for more complex additions later on

// // Step 2: iterate through the agents, telling them the proposed directions and letting them rank order them based on preference.
// finalVotes := make(map[uuid.UUID]voting.LootboxVoteMap, len(agents))
// for _, agent := range agents {
// 	finalVotes[agent.GetID()] = agent.FinalDirectionVote(proposedDirections)
// }

// // Step 3: get the winning direction from the final votes
// direction := s.GetWinningDirection(finalVotes, weights)
// if _, ok := s.lootBoxes[direction]; !ok {
// 	panic("agents voted on a non-existent lootbox")
// }

// return direction
