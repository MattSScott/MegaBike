package agent

import (
	"SOMAS2023/internal/clients/teamSOSA/modules"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"math"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// ----- Decisions (Iteration level) -----

// returns: int reflecting what action the agent has decided to do this iteration, pedal the bike (0) or try to change bikes (1)
func (a *AgentSOSA) DecideAction() objects.BikerAction {

	avgTrust := a.Modules.AgentParameters.GetAverageTrust()

	if avgTrust > modules.StayOnBikeThreshold {
		// Pedal if members of the bike have high trust
		return objects.Pedal
	} else {
		// Otherwise, change bikes.
		return objects.ChangeBike
	}
}

// returns: map of UUID -> {0,1} for an agent where 0 means 'don't kick' and 1 means 'do kick'
func (a *AgentSOSA) VoteForKickout() map[uuid.UUID]int {
	VoteMap := make(map[uuid.UUID]int)

	// check all bikers on the bike but ignore ourselves
	for _, agent := range a.GetFellowBikers() {
		if agent.GetID() != a.GetID() {
			_, exists := a.Modules.AgentParameters.TrustNetwork[agent.GetID()]

			if a.Modules.AgentParameters.TrustNetwork[agent.GetID()] < modules.KickThreshold && exists {
				VoteMap[agent.GetID()] = 1
			} else {
				VoteMap[agent.GetID()] = 0
			}

		}
	}

	return VoteMap
}

// returns: slice of agents to kick out 
func (a *AgentSOSA) DecideKickOut() []uuid.UUID {
	// currently just kick out lowest agent if below a threshold.
	kickOut_agents := make([]uuid.UUID, 0)
	minTrustBiker := a.GetTeammateWithMinTrust()
	// if minTrustBiker.ID != uuid.Nil && minTrustBiker.Trust < modules.KickThreshold {
	// 	kickOut_agents = append(kickOut_agents, minTrustBiker.ID)
	// }
	if minTrustBiker.ID != uuid.Nil {
		kickOut_agents = append(kickOut_agents, minTrustBiker.ID)
	}
	return kickOut_agents
}

// returns: map of pending agents uuid -> {true, false} where true means accept and false means dont accept
func (a *AgentSOSA) DecideJoining(pendingAgents []uuid.UUID) map[uuid.UUID]bool {

	// Accept all agents we don't know about or are higher in trust than a threshold.

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

func (a *AgentSOSA) DecideJoiningOneAgent(agentId uuid.UUID) bool {
	// same logic as above, basically if we know them, check their id against a threshold to accept, else reject. if not known accept
	if _, ok := a.Modules.AgentParameters.TrustNetwork[agentId]; ok {
		if a.Modules.AgentParameters.TrustNetwork[agentId] > modules.AcceptThreshold {
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}

// returns: uuid of the target bike if they want to change, otherwise just return current bike id 
func (a *AgentSOSA) DecideChangeBike() uuid.UUID {
	shouldChangeBike := false
	targetBikeID := uuid.Nil
	// if the average trust on our bike is less than threshold, change bike (note: previusly getaveragetrust calculated average of whole network, even agents not on bike. made no sense.)
	if a.GetAverageTrustOnBike() < modules.StayOnBikeThreshold {
		shouldChangeBike = true
		targetBikeID = a.Modules.Environment.GetBikeWithMaximumTrust(a.Modules.AgentParameters)
	}

	if shouldChangeBike {
		return targetBikeID
	} else {
		return a.Modules.Environment.BikeId // maybe change to a.getbike
	}
}

func (a *AgentSOSA) DecideBikePreferenceOrder() ([]uuid.UUID, map[uuid.UUID]bool) {
	megabikes := a.GetGameState().GetMegaBikes()
	bikePreferenceScores := make(map[uuid.UUID]float64)
	bikeTrustHigherThanRegime := make(map[uuid.UUID]bool)

	for bikeId, bike := range megabikes {
		// Calculate the average trust for this bike
		sum := 0.0
		for _, agent := range bike.GetAgents() {
			sum += a.Modules.AgentParameters.TrustNetwork[agent.GetID()]
		}
		bikeAverageTrust := sum / float64(len(bike.GetAgents()))

		// Regime trust
		regimeTrust := a.Modules.AgentParameters.RegimeTrust[bike.GetGovernance()]

		// Score the bike (weighted average)
		score := 0.5*bikeAverageTrust + 0.5*regimeTrust
		bikePreferenceScores[bikeId] = score

		// Save the trust comparison flag
		bikeTrustHigherThanRegime[bikeId] = bikeAverageTrust > regimeTrust
	}

	// Build slice of bike IDs to sort
	var bikePreferenceOrder []uuid.UUID
	for id := range bikePreferenceScores {
		bikePreferenceOrder = append(bikePreferenceOrder, id)
	}

	// Sort by descending preference score
	sort.Slice(bikePreferenceOrder, func(i, j int) bool {
		return bikePreferenceScores[bikePreferenceOrder[i]] > bikePreferenceScores[bikePreferenceOrder[j]]
	})

	return bikePreferenceOrder, bikeTrustHigherThanRegime
}

// ----- Decisions (Round level) -----

// returns: lootbox uuid to aim towards (the direction) from a subset of lootboxes
func (a *AgentSOSA) ProposeDirectionFromSubset(subset map[uuid.UUID]objects.ILootBox) uuid.UUID {
	agentColour, agentEnergy := a.GetColour(), a.GetEnergyLevel()
	optimalLootbox := a.Modules.Environment.GetNearestLootboxByColourFromSubset(agentColour, subset)
	nearestLootbox := a.Modules.Environment.GetNearestLootboxFromSubset(subset)
	if agentEnergy < modules.EnergyToOptimalLootboxThreshold || optimalLootbox == uuid.Nil {
		return nearestLootbox
	}
	return optimalLootbox
}

// returns: uuid of lootbox to aim towards (i.e. the direction).
func (a *AgentSOSA) DecideDirection() uuid.UUID {

	// first check if awdi is near. if so, then move away regardless of personality or trust. dont want to get obliterated.
	if a.Modules.Environment.IsAwdiNear() {
		return a.Modules.Environment.GetNearestLootboxAwayFromAwdi()
	}

	if a.Modules.AgentParameters.PlatonicTendency > 0.5 || a.Modules.AgentParameters.GetAverageTrust() > 0.5 {
		// if selfless or high trust, get highest gain lootbox
		return a.Modules.Environment.GetHighestGainLootbox()
	} else {
		// Otherwise, move towards the nearest lootbox of your own colour
		return a.Modules.Environment.GetNearestLootboxByColour(a.GetColour())
	}
}

// decides the force the biker is going to pedal with
func (a *AgentSOSA) DecideForce(direction uuid.UUID) {
	if direction == uuid.Nil {
		return
		// lootboxId := a.Modules.Environment.GetHighestGainLootbox()
		// lootboxPos := a.Modules.Environment.GetLootboxPos(lootboxId)
		// a.SetForces(a.Modules.Utils.GetForcesToTarget(a.GetLocation(), lootboxPos))
		// return
	}

	if a.Modules.Environment.IsAwdiNear() {
		// fmt.Printf("[DecideForce] Agent %s is near Awdi\n", a.GetID())
		// Move in opposite direction to Awdi in full force
		bikePos, awdiPos := a.Modules.Environment.GetBike().GetPosition(), a.Modules.Environment.GetAwdi().GetPosition()
		force := a.Modules.Utils.GetForcesToTargetWithDirectionOffset(utils.BikerMaxForce, 1.0-a.Modules.Environment.GetBikeOrientation(), bikePos, awdiPos)
		a.SetForces(force)
		return
	}
	// Use the platonic tendency to decide whether to pedal in the chosen direction or not
	probabilityOfConformity := a.Modules.AgentParameters.PlatonicTendency
	randomNumber := rand.Float64()
	agentPosition := a.GetLocation()
	lootboxID := direction
	a.SetRoundDidConform(true)
	if randomNumber > probabilityOfConformity {
		lootboxID = a.Modules.Environment.GetHighestGainLootbox()
		a.SetRoundDidConform(false)
		
	}
	lootboxPosition := a.Modules.Environment.GetLootboxPos(lootboxID)
	force := a.Modules.Utils.GetForcesToTargetWithDirectionOffset(utils.BikerMaxForce, -a.Modules.Environment.GetBikeOrientation(), agentPosition, lootboxPosition)
	a.SetForces(force)
}

// returns: map containing bikerID -> distribution (i.e. share of resources) 
func (a *AgentSOSA) DecideAllocation() voting.IdVoteMap {

	allocation := make(map[uuid.UUID]float64)
	allocation[a.GetID()] = 1.0

	normConst := 1.0

	trustNet := a.Modules.AgentParameters.TrustNetwork

	for _, biker := range a.GetFellowBikers() {
		if a.GetID() == biker.GetID() {
			continue
		}

		agentAllocation, ok := trustNet[biker.GetID()]
		if !ok {
			agentAllocation = 0.5
		}

		allocation[biker.GetID()] = agentAllocation
		normConst += agentAllocation
	}

	for id, val := range allocation {
		allocation[id] = val / normConst
	}

	if a.Modules.AgentParameters.PlatonicTendency > 0.5 || a.GetAverageTrustOnBike() > 0.5 {
		// if a fair agent or high trust, simply allocate according to trust in each agent as done above
		return allocation
	} else {
		// if these conditions do not hold, then do same as above, but cut every elses share by 50% and give yourself the rest.
		totalShareDistributed := float64(0)
		
		for agentID, share := range allocation {
			if agentID != a.GetID() {
				share = share * 0.5
				totalShareDistributed += share
			}
		}

		myShare := 1 - totalShareDistributed
		allocation[a.GetID()] = myShare

		return allocation
	}
}

// ----- Helper Functions -----

// returns: slice containing the bikers on our bike.
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

// returns: the biker on your bike with the minimum trust 
func (a *AgentSOSA) GetTeammateWithMinTrust() modules.IDTrustPair {
	fellowBikers := a.GetFellowBikers()
	minTrustAgentId := uuid.Nil
	minTrust := math.MaxFloat64
	for _, fellowBiker := range fellowBikers {
		if fellowBiker.GetID() != a.GetID() {
			if trust, ok := a.Modules.AgentParameters.TrustNetwork[fellowBiker.GetID()]; ok {
				if trust < minTrust {
					minTrustAgentId = fellowBiker.GetID()
					minTrust = trust
				}
			}
		}
	}

	return modules.IDTrustPair{ID: minTrustAgentId, Trust: minTrust}

}

// returns: float representing the sum of the trust of agents on your bike.
func (a *AgentSOSA) GetSumOfTrustOnBike() float64 {
	var sum = 0.0
	for _, teammate := range a.GetFellowBikers() {
		sum += a.Modules.AgentParameters.TrustNetwork[teammate.GetID()]
	}
	return sum
}

// returns: float representing average trust of your bike
func (a *AgentSOSA) GetAverageTrustOnBike() float64 {
	// Prevent divide
	if len(a.GetFellowBikers()) == 0 {
		return 0.5
	}

	sum := a.GetSumOfTrustOnBike()

	return sum / float64(len(a.GetFellowBikers()))
}

// sets the bike the agent is on (or wants to be on)
func (a *AgentSOSA) SetBike(bikeId uuid.UUID) {
	a.Modules.Environment.BikeId = bikeId
	a.BaseBiker.SetBike(bikeId)
}

// delete the agent with the specified id from this agents trust map
func (a *AgentSOSA) HandleAgentUnalive(id uuid.UUID) {
	delete(a.Modules.AgentParameters.TrustNetwork, id)
}




// ----- do i keep? -----

func (a *AgentSOSA) ProposeNewRadius(pRad float64) float64 {
	energy := a.GetEnergyLevel()
	newRad := pRad
	if energy < 0.75 {
		newRad = pRad * 2
	}

	return math.Max(newRad, -100000)

}

// returns: lootbox uuid to aim towards (the direction)
func (a *AgentSOSA) ProposeDirection() uuid.UUID {
	agentColour, agentEnergy := a.GetColour(), a.GetEnergyLevel()
	optimalLootbox := a.Modules.Environment.GetNearestLootboxByColour(agentColour)
	nearestLootbox := a.Modules.Environment.GetNearestLootbox()
	if agentEnergy < modules.EnergyToOptimalLootboxThreshold || optimalLootbox == uuid.Nil {
		return nearestLootbox
	}
	return optimalLootbox
}

func (a *AgentSOSA) GetTrustOfAgent(agent objects.IBaseBiker) float64 {
    return a.Modules.AgentParameters.TrustNetwork[agent.GetID()]
}