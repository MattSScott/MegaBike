package server

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"

	"github.com/google/uuid"
	"math/rand"
)

// obtain direction for current round from the representatives
func (s *Server) RunRepresentativeAction(bike objects.IMegaBike) uuid.UUID {
	agents := s.GetAgentMap()
	governance := bike.GetGovernance()
	reps := bike.GetRepresentatives()
	var direction uuid.UUID

	// decide differently based on each governance...
	switch governance {
	case utils.PerfectAristocracy:
		selectedAgents := make([]objects.IBaseBiker, 0, len(reps))
		suggestedDirections := make([]uuid.UUID, 0, len(reps))
		countsPerDirection := make(map[uuid.UUID]int)
		maxCounts := 0
		
		// create a slice of representative agents
		for _, id := range reps {
			if agent, exists := agents[id]; exists {
				selectedAgents = append(selectedAgents, agent)
			}
		}

		// create a slice of their suggested directions
		for _, ag := range selectedAgents {
			suggestedDirections = append(suggestedDirections, ag.DecideDirectionBenevolently())
		}

		// aggregate these to find the majority voted direction
		for _, lootbox := range suggestedDirections {
			countsPerDirection[lootbox] += 1
			if countsPerDirection[lootbox] > maxCounts {
				maxCounts = countsPerDirection[lootbox]
				direction = lootbox
			}
		}

		return direction

	case utils.DegenerateAristocracy:
		selectedAgents := make([]objects.IBaseBiker, 0, len(reps))
		suggestedDirections := make([]uuid.UUID, 0, len(reps))
		countsPerDirection := make(map[uuid.UUID]int)
		maxCounts := 0
		
		// create a slice of representative agents
		for _, id := range reps {
			if agent, exists := agents[id]; exists {
				selectedAgents = append(selectedAgents, agent)
			}
		}

		// create a slice of their suggested directions
		for _, ag := range selectedAgents {
			suggestedDirections = append(suggestedDirections, ag.DecideDirectionMalevolently())
		}

		// aggregate these to find the majority voted direction
		for _, lootbox := range suggestedDirections {
			countsPerDirection[lootbox] += 1
			if countsPerDirection[lootbox] > maxCounts {
				maxCounts = countsPerDirection[lootbox]
				direction = lootbox
			}
		}

		return direction
		
	case utils.PerfectMonarchy:
		monarch := agents[reps[0]]
		direction = monarch.DecideDirectionBenevolently()
		return direction

	case utils.DegenerateMonarchy:
		monarchID := reps[0]
		monarch := agents[monarchID]
		direction = monarch.DecideDirectionMalevolently()
		return direction

	default:
		panic("trying to run representative action in a non-representative governance")
	}
}

// select representatives for the bike. is not a democratic process:
// each bike is fixed with a governance style and agents on the bike are randomly selected to fit that style.
// happens at the beginning of each iteration, or when a bike with certain governance styles is left without representatives for any of various reasons
func (s *Server) RepresentativeElection(agentsOnBike []objects.IBaseBiker, governance utils.Governance) []uuid.UUID {
	// votes := make(map[uuid.UUID]voting.IdVoteMap, len(agents))
	// voteWeight := make(map[uuid.UUID]float64)
	// for _, agent := range agents {
	// 	voteWeight[agent.GetID()] = 1
	// 	switch governance {
	// 	case utils.PerfectMonarchy:
	// 		votes[agent.GetID()] = agent.VoteDictator()
	// 	case utils.DegenerateMonarchy:
	// 		votes[agent.GetID()] = agent.VoteDictator()
	// 	}
	// }

	// // required as a list of interfaces that implement IVoter is not percieved as a list of IVoters due to Go weirdness
	// IVotes := make(map[uuid.UUID]voting.IVoter, len(votes))
	// for i, vote := range votes {
	// 	IVotes[i] = vote
	// }

	// ruler := voting.WinnerFromDist(IVotes, voteWeight)
	// return ruler
	if len(agentsOnBike) != 0 {
		switch governance {
		case utils.PerfectAristocracy, utils.DegenerateAristocracy:
			reps := make([]uuid.UUID, 3)
			chosenAgentIndices := rand.Perm(len(agentsOnBike))[0:3]
			for i, v := range chosenAgentIndices{
				reps[i] = agentsOnBike[v].GetID()
			}
			return reps
		case utils.PerfectMonarchy, utils.DegenerateMonarchy:
			reps := make([]uuid.UUID, 1)
			chosenAgentIndex := rand.Intn(len(agentsOnBike))
			reps[0] = agentsOnBike[chosenAgentIndex].GetID()
			return reps
		default:
			panic("trying to run a representative election on a bike without incorrect governance style")
		}
	} else {
		return []uuid.UUID{}
	}
		
}

// somehow reduces the number of lootboxes available - maybe todo with radius?
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

// select this round's decision following a voting-based approach
func (s *Server) RunDemocraticAction(bike objects.IMegaBike, governance utils.Governance) uuid.UUID {

	agents := bike.GetAgents()
	// maps agent id to their proposed lootbox id
	proposedDirections := make(map[uuid.UUID]uuid.UUID)
	validLootboxes := s.PruneLootboxes(bike)


	for _, agent := range agents {
		// agents that have decided to stay on the bike (and that haven't been kicked off it)
		// will participate in the voting for the directions

		// Step 1: get every agent on the bikes' proposed direction
		if agent.GetBikeStatus() {
			// proposedDirection := agent.ProposeDirection()
			proposedDirection := agent.ProposeDirectionFromSubset(validLootboxes)

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


	// different behaviours depending on the kind of democracy
	if governance == utils.PerfectDemocracy {
		// look for consensus
		directions := []uuid.UUID{}
		for _, direction := range proposedDirections {
			directions = append(directions, direction)
		}

		consensusReached := true
		for i := 1; i < len(directions); i++ {
			if directions[i] != directions[0] {
				consensusReached = false
				break
			}
		}
		
		if consensusReached == true {
			return directions[0] // could be any element of the slice they are all the asme
		} else {
			return uuid.Nil // assuming this means 'don't move'
		}


	} else if governance == utils.DegenerateDemocracy {
		// look for majority

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
}

func (s *Server) GetWinningDirection(finalVotes map[uuid.UUID]voting.LootboxVoteMap, weights map[uuid.UUID]float64) uuid.UUID {
	// get overall winner direction using chosen voting strategy

	// this allows to get a slice of the interface from that of the specific type
	// this way we can substitute agent.FInalDirectionVote with another function that returns
	// another type of voting type which still implements INormaliseVoteMap
	IfinalVotes := make(map[uuid.UUID]voting.IVoter)
	for i, v := range finalVotes {
		IfinalVotes[i] = v
	}

	return voting.WinnerFromDist(IfinalVotes, weights)
}
