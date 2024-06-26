package agent

import (
	"SOMAS2023/internal/clients/teamSOSA/modules"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"maps"
	"math"
	"math/rand"
	"runtime"

	"github.com/google/uuid"
)

func (a *AgentSOSA) GetFellowBikers() []objects.IBaseBiker {
	bikes := a.Modules.Environment.GameState.GetMegaBikes()
	if _, ok := bikes[a.GetBike()]; !ok {
		return []objects.IBaseBiker{}
	}
	bike := bikes[a.GetBike()]
	fellowBikers := make([]objects.IBaseBiker, 0)
	for _, biker := range bike.GetAgents() {
		if biker.GetBikeStatus() {
			fellowBikers = append(fellowBikers, biker)
		}
	}

	return fellowBikers
}

// We vote for ourselves and the agent with the highest social capital.
func (a *AgentSOSA) VoteDictator() voting.IdVoteMap {
	votes := make(voting.IdVoteMap)
	agentIDStruct := a.Modules.Environment.GetBikerWithMaxSocialCapital(a.Modules.AgentParameters)
	agentId := agentIDStruct.ID

	if len(a.GetFellowBikers()) > 1 && agentId != a.GetID() && agentId != uuid.Nil {
		fellowBikers := a.GetFellowBikers()
		for _, fellowBiker := range fellowBikers {
			if fellowBiker.GetID() == agentId || fellowBiker.GetID() == a.GetID() {
				votes[fellowBiker.GetID()] = 0.5
			} else {
				votes[fellowBiker.GetID()] = 0.0
			}
		}
	} else {
		fellowBikers := a.GetFellowBikers()
		for _, fellowBiker := range fellowBikers {
			if fellowBiker.GetID() == a.GetID() {
				votes[fellowBiker.GetID()] = 1.0
			} else {
				votes[fellowBiker.GetID()] = 0.0
			}
		}
	}
	return votes
}

func (a *AgentSOSA) calculateUntrustworthyWeighting(id uuid.UUID) float64 {
	if a.GetID() == id {
		return 1.0
	}
	return 0.0
}

func (a *AgentSOSA) DecideWeights(action utils.Action) map[uuid.UUID]float64 {
	// All actions have equal weights. Weighting by AgentId based on social capital.
	// We set the weight for an Agent to be equal to its Social Capital.
	weights := make(map[uuid.UUID]float64)
	agents := a.GetFellowBikers()
	compRoll := rand.Float64()
	willComply := compRoll < a.Modules.AgentParameters.Trustworthiness
	for _, agent := range agents {
		// if agent Id is not in the a.Modules.SocialCapital.SocialCapital map, set the weight to 0.5 (neither trust or distrust)
		if _, ok := a.Modules.AgentParameters.TrustNetwork[agent.GetID()]; !ok {
			// add agent to the map
			a.Modules.AgentParameters.TrustNetwork[agent.GetID()] = 0.5
		}
		agentWeighting := a.calculateUntrustworthyWeighting(agent.GetID()) // weight own action by 100%, if non compliant
		if willComply {                                                    // give 'fair' weighting based on trust, if agent is trustworthy
			agentWeighting = a.Modules.AgentParameters.TrustNetwork[agent.GetID()]
		}
		weights[agent.GetID()] = agentWeighting
		// fmt.Printf("[DecideWeights G2] Agent %s has weight %f\n", agent.GetID(), weights[agent.GetID()])
	}
	return weights
}

func (a *AgentSOSA) DecideKickOut() []uuid.UUID {
	// Only called when the agent is the dictator.
	// We kick out the agent with the lowest social capital on the bike.
	// GetBikerWithMinSocialCapital returns only one agent, if more agents with min SC, it randomly chooses one.
	kickOut_agents := make([]uuid.UUID, 0)
	agentIDStruct := a.Modules.Environment.GetBikerWithMinSocialCapital(a.Modules.AgentParameters)
	agentId := agentIDStruct.ID
	if agentId != uuid.Nil {
		kickOut_agents = append(kickOut_agents, agentId)
	}
	return kickOut_agents
}

func (a *AgentSOSA) VoteLeader() voting.IdVoteMap {
	// We vote 0.5 for ourselves if the agent with the highest SC Agent(that we've met so far) on our bike. If we're alone on a bike, we vote 1 for ourselves.
	votes := make(voting.IdVoteMap)
	fellowBikers := a.GetFellowBikers()
	if len(a.GetFellowBikers()) > 0 {
		agentIDStruct := a.Modules.Environment.GetBikerWithMaxSocialCapital(a.Modules.AgentParameters)
		agentId := agentIDStruct.ID
		for _, fellowBiker := range fellowBikers {
			if fellowBiker.GetID() == agentId {
				votes[fellowBiker.GetID()] = 0.5
			} else if fellowBiker.GetID() == a.GetID() {
				votes[a.GetID()] = 0.5
			} else {
				votes[fellowBiker.GetID()] = 0.0
			}
		}
	} else {
		votes[a.GetID()] = 1.0
	}

	return votes
}

func (a *AgentSOSA) DecideGovernance() utils.Governance {
	// All possibilities except dictatorship.
	// Need to decide weights for each type of Governance
	// Can add an invalid weighting so that it is not 50/50

	randomNumber := rand.Float64()
	if randomNumber < democracyWeight {
		return utils.Democracy
	} else if randomNumber < democracyWeight+leadershipWeight {
		return utils.Leadership
	} else {
		return utils.Dictatorship
	}
}

func (a *AgentSOSA) DecideAllocation() voting.IdVoteMap {
	socialCapital := maps.Clone(a.Modules.AgentParameters.TrustNetwork)
	// Iterate through agents in social capital
	for id := range socialCapital {
		// Iterate through fellow bikers
		for _, biker := range a.GetFellowBikers() {
			// If this agent is a fellow biker, move on
			if biker.GetID() == id {
				continue
			}
		}
		if math.IsNaN(socialCapital[id]) {
			runtime.Breakpoint()
			panic("dhsd")
		}
		// This agent is not a fellow biker - remove it from SC
		delete(socialCapital, id)
	}
	// We give ourselves 1.0
	socialCapital[a.GetID()] = 1.0
	return socialCapital
}

func (a *AgentSOSA) DecideDictatorAllocation() voting.IdVoteMap {
	socialCapital := a.DecideAllocation()

	// Calculate the total social capital
	totalSocialCapital := 0.0
	for _, sc := range socialCapital {
		totalSocialCapital += sc
	}

	// Distribute the allocation based on each agent's share of the total social capital
	result := make(voting.IdVoteMap)
	for agentID, sc := range socialCapital {
		result[agentID] = sc / totalSocialCapital
		if math.IsNaN(result[agentID]) {
			runtime.Breakpoint()
			panic("fuck")
		}
	}
	return result
}

func (a *AgentSOSA) VoteForKickout() map[uuid.UUID]int {
	VoteMap := make(map[uuid.UUID]int)
	kickoutThreshold := modules.KickThreshold
	AgentSOSAID := a.GetID()

	// check all bikers on the bike but ignore ourselves
	for _, agent := range a.GetFellowBikers() {
		if agent.GetID() != AgentSOSAID {
			_, exists := a.Modules.AgentParameters.TrustNetwork[agent.GetID()]

			if a.Modules.AgentParameters.TrustNetwork[agent.GetID()] < kickoutThreshold && exists {
				VoteMap[agent.GetID()] = 1
			} else {
				VoteMap[agent.GetID()] = 0
			}

		}
	}

	return VoteMap
}

func (a *AgentSOSA) DecideJoining(pendingAgents []uuid.UUID) map[uuid.UUID]bool {
	// Accept all agents we don't know about or are higher in social capital.
	// If we know about them and they have a lower social capital, reject them.

	decision := make(map[uuid.UUID]bool)
	for _, agent := range pendingAgents {
		// If we know about them and they have a higher social capital than threshold, accept them.
		if _, ok := a.Modules.AgentParameters.TrustNetwork[agent]; ok {
			if a.Modules.AgentParameters.TrustNetwork[agent] > modules.AcceptThreshold {
				decision[agent] = true
			} else {
				decision[agent] = false
			}
		} else {
			decision[agent] = true
		}
	}
	return decision
}

func (a *AgentSOSA) ProposeDirectionFromSubset(subset map[uuid.UUID]objects.ILootBox) uuid.UUID {
	agentID, agentColour, agentEnergy := a.GetID(), a.GetColour(), a.GetEnergyLevel()
	optimalLootbox := a.Modules.Environment.GetNearestLootboxByColorFromSubset(agentID, agentColour, subset)
	nearestLootbox := a.Modules.Environment.GetNearestLootboxFromSubset(agentID, subset)
	if agentEnergy < modules.EnergyToOptimalLootboxThreshold || optimalLootbox == uuid.Nil {
		return nearestLootbox
	}
	return optimalLootbox
}

func (a *AgentSOSA) ProposeNewRadius(pRad float64) float64 {
	energy := a.GetEnergyLevel()
	newRad := pRad * 0.95
	if energy < 0.5 {
		newRad = pRad * 1.05
	}
	return math.Max(newRad, -100000)
}

func (a *AgentSOSA) ProposeDirection() uuid.UUID {
	agentID, agentColour, agentEnergy := a.GetID(), a.GetColour(), a.GetEnergyLevel()
	optimalLootbox := a.Modules.Environment.GetNearestLootboxByColor(agentID, agentColour)
	nearestLootbox := a.Modules.Environment.GetNearestLootbox(agentID)
	if agentEnergy < modules.EnergyToOptimalLootboxThreshold || optimalLootbox == uuid.Nil {
		return nearestLootbox
	}
	return optimalLootbox
}

func (a *AgentSOSA) FinalDirectionVote(proposals map[uuid.UUID]uuid.UUID) voting.LootboxVoteMap {
	// fmt.Printf("[FFinalDirectionVote] Agent %s got proposals %v\n", a.GetID(), proposals)
	// fmt.Printf("[FFinalDirectionVote] Agent %s has Social Capitals %v\n", a.GetID(), a.Modules.SocialCapital.SocialCapital)

	votes := make(voting.LootboxVoteMap)

	// Assume we set our own social capital to 1.0, thus need to account for it
	weight := 1.0 / (a.Modules.AgentParameters.GetSumOfTrust() + 1)

	for proposerID, proposal := range proposals {
		scWeight := 0.0
		if proposerID == a.GetID() {
			// If the proposal is our own, we vote for it with full weight
			scWeight = weight
		} else {
			scWeight = weight * a.Modules.AgentParameters.TrustNetwork[proposerID]
		}

		// Check if the proposal already exists in votes, if not add it with the calculated weight
		if _, ok := votes[proposal]; !ok {
			votes[proposal] = scWeight
		} else {
			// If the proposal is already there, update it
			votes[proposal] += scWeight
		}
	}
	// fmt.Printf("[FFinalDirectionVote] Agent %s voted %v\n", a.GetID(), votes)
	return votes
}

func (a *AgentSOSA) ChangeBike() uuid.UUID {
	decisionInputs := modules.DecisionInputs{AgentParameters: a.Modules.AgentParameters, Environment: a.Modules.Environment, AgentID: a.GetID()}
	isChangeBike, bikeId := a.Modules.Decision.MakeBikeChangeDecision(decisionInputs)
	// fmt.Printf("[ChangeBike] Agent %s decided to change bike: %v\n", a.GetID(), isChangeBike)
	if isChangeBike {
		// fmt.Printf("[ChangeBike] Agent %s decided to change bike to: %v\n", a.GetID(), bikeId)
		return bikeId
	} else {
		return a.Modules.Environment.BikeId
	}
}

func (a *AgentSOSA) DecideAction() objects.BikerAction {
	// fmt.Printf("[DecideAction] Agent %s has Social Capitals %v\n", a.GetID(), a.Modules.SocialCapital.SocialCapital)
	// a.Modules.SocialCapital.UpdateSocialCapital()

	avgSocialCapital := a.Modules.AgentParameters.GetAverageTrust()

	if avgSocialCapital > 0 {
		// Pedal if members of the bike have high social capital.
		return objects.Pedal
	} else {
		// Otherwise, change bikes.
		return objects.ChangeBike
	}
}

func (a *AgentSOSA) DecideForce(direction uuid.UUID) {
	if direction == uuid.Nil {
		return
		// lootboxId := a.Modules.Environment.GetHighestGainLootbox()
		// lootboxPos := a.Modules.Environment.GetLootboxPos(lootboxId)
		// a.SetForces(a.Modules.Utils.GetForcesToTarget(a.GetLocation(), lootboxPos))
		// return
	}

	a.Modules.VotedDirection = direction

	if a.Modules.Environment.IsAwdiNear() {
		// fmt.Printf("[DecideForce] Agent %s is near Awdi\n", a.GetID())
		// Move in opposite direction to Awdi in full force
		bikePos, awdiPos := a.Modules.Environment.GetBike().GetPosition(), a.Modules.Environment.GetAwdi().GetPosition()
		force := a.Modules.Utils.GetForcesToTargetWithDirectionOffset(utils.BikerMaxForce, 1.0-a.Modules.Environment.GetBikeOrientation(), bikePos, awdiPos)
		a.SetForces(force)
		return
	}
	// Use the average social capital to decide whether to pedal in the voted direciton or not
	probabilityOfConformity := a.Modules.AgentParameters.GetAverageTrust()
	randomNumber := rand.Float64()
	agentPosition := a.GetLocation()
	lootboxID := direction
	if randomNumber > probabilityOfConformity {
		lootboxID = a.Modules.Environment.GetHighestGainLootbox()
	}
	lootboxPosition := a.Modules.Environment.GetLootboxPos(lootboxID)
	force := a.Modules.Utils.GetForcesToTargetWithDirectionOffset(utils.BikerMaxForce, -a.Modules.Environment.GetBikeOrientation(), agentPosition, lootboxPosition)
	a.SetForces(force)
}

func (a *AgentSOSA) DictateDirection() uuid.UUID {
	// Move in opposite direction to Awdi in full force
	if a.Modules.Environment.IsAwdiNear() {
		// fmt.Printf("[DictateDirection] Agent %s is near Awdi\n", a.GetID())
		return a.Modules.Environment.GetNearestLootboxAwayFromAwdi()
	}
	// Otherwise, move towards the lootbox with the highest gain
	return a.Modules.Environment.GetHighestGainLootbox()
}

func (a *AgentSOSA) SetBike(bikeId uuid.UUID) {
	a.Modules.Environment.BikeId = bikeId
	a.BaseBiker.SetBike(bikeId)
}
