package server

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"

	"github.com/google/uuid"
)

// obtain direction for current round from the dictator
func (s *Server) RunRulerAction(bike objects.IMegaBike) uuid.UUID {
	agents := s.GetAgentMap()
	ruler := agents[bike.GetRuler()]
	// get dictators direction choice
	direction := ruler.DictateDirection()
	return direction
}

// elect ruler (happens during the foundation stage, or when a bike with ruler-lead
// governance is left without ruler for any of various reasons)
func (s *Server) RulerElection(agents []objects.IBaseBiker, governance utils.Governance) uuid.UUID {
	votes := make(map[uuid.UUID]voting.IdVoteMap, len(agents))
	voteWeight := make(map[uuid.UUID]float64)
	for _, agent := range agents {
		voteWeight[agent.GetID()] = 1
		switch governance {
		case utils.Dictatorship:
			votes[agent.GetID()] = agent.VoteDictator()
		case utils.Leadership:
			votes[agent.GetID()] = agent.VoteLeader()
		}
	}

	// required as a list of interfaces that implement IVoter is not percieved as a list of IVoters due to Go weirdness
	IVotes := make(map[uuid.UUID]voting.IVoter, len(votes))
	for i, vote := range votes {
		IVotes[i] = vote
	}

	ruler := voting.WinnerFromDist(IVotes, voteWeight)
	return ruler
}

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

// select this round's decision following a voting-based approach (with weights in the case of a leadership-led governance)
func (s *Server) RunDemocraticAction(bike objects.IMegaBike, weights map[uuid.UUID]float64) uuid.UUID {
	// map of the proposed lootboxes by bike (for each bike a list of lootbox proposals is made, with one lootbox proposed by each agent on the bike)
	agents := bike.GetAgents()
	proposedDirections := make(map[uuid.UUID]uuid.UUID)
	validLootboxes := s.PruneLootboxes(bike)

	for _, agent := range agents {
		// agents that have decided to stay on the bike (and that haven't been kicked off it)
		// will participate in the voting for the directions
		// ---------------------------VOTING ROUTINE - STEP 1 ---------------------
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

	s.UpdateBikeRules(bike)

	if len(proposedDirections) == 0 {
		return uuid.Nil
	}

	finalVotes := make(map[uuid.UUID]voting.LootboxVoteMap, len(agents))
	for _, agent := range agents {
		// ---------------------------VOTING ROUTINE - STEP 2 ---------------------
		// pass the pitched directions of a bike to all agents on that bike and get their final vote
		finalVotes[agent.GetID()] = agent.FinalDirectionVote(proposedDirections)
	}

	// ---------------------------VOTING ROUTINE - STEP 3 --------------
	// get the winning direction from the final votes
	direction := s.GetWinningDirection(finalVotes, weights)
	if _, ok := s.lootBoxes[direction]; !ok {
		panic("agents voted on a non-existent lootbox")
	}
	return direction
}
