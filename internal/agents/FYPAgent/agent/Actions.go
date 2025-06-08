package agent

import (
	"MegabikeFYPVersion/internal/agents/FYPAgent/modules"
	"MegabikeFYPVersion/internal/common/objects"
	"MegabikeFYPVersion/internal/common/utils"
	"MegabikeFYPVersion/internal/common/voting"
	"math"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// ----- Decisions (Iteration level) -----

// returns: a yes/no decision as to whether we want to accept the given agent to our bike
func (a *FYPAgent) DecideJoiningOneAgent(agentId uuid.UUID) bool {
	if _, ok := a.Modules.AgentParameters.TrustNetwork[agentId]; ok {
		// if we know them, check their id against a threshold to accept, else reject.
		if a.Modules.AgentParameters.TrustNetwork[agentId] > modules.AcceptThreshold {
			return true
		} else {
			return false
		}
	} else {
		// if we don't know them, accept.
		return true
	}
}

// returns: a slice of bikes we want to join, sorted in descending order in terms of score.
func (a *FYPAgent) DecideBikePreferenceOrder() []utils.BikePreferenceData {

	megabikes := a.GetGameState().GetMegaBikes()
	var bikePreferenceOrder []utils.BikePreferenceData

	for bikeId, bike := range megabikes {

		// create a new bikedata object and add the id
		bikeData := utils.BikePreferenceData{BikeId: bikeId}

		// Calculate the average trust for this bike
		sum := 0.0
		for _, agent := range bike.GetAgents() {
			sum += a.Modules.AgentParameters.TrustNetwork[agent.GetID()]
		}

		bikeAverageTrust := sum / float64(len(bike.GetAgents()))

		// Regime trust
		regimeTrust := a.Modules.AgentParameters.RegimeTrust[bike.GetGovernance()]

		// Score the bike (weighted average) and add it to bike data
		score := 0.5*bikeAverageTrust + 0.5*regimeTrust
		bikeData.Score = score

		// Save the trust comparison flag to the bike data
		bikeData.BikeTrustHigherThanRegimeTrust = bikeAverageTrust > regimeTrust

		// append it to the array
		bikePreferenceOrder = append(bikePreferenceOrder, bikeData)
	}

	// Sort by descending preference score
	sort.Slice(bikePreferenceOrder, func(i, j int) bool {
		return bikePreferenceOrder[i].Score > bikePreferenceOrder[j].Score
	})

	return bikePreferenceOrder
}

// ----- Decisions (Round level) -----

// returns: uuid of lootbox we want to aim towards (i.e. the direction).
func (a *FYPAgent) DecideDirection() uuid.UUID {

	// first check if awdi is near. if so, then move away regardless of personality or trust. dont want to get obliterated.
	if a.Modules.Environment.IsAwdiNear() {
		return a.Modules.Environment.GetNearestLootboxAwayFromAwdi()
	}

	if a.Modules.AgentParameters.PlatonicTendency > 0.5 || a.GetAverageTrustOnBike() > 0.5 {
		// if selfless or high trust, get highest gain lootbox
		return a.Modules.Environment.GetHighestGainLootbox()
	} else {
		// Otherwise, move towards the nearest lootbox of your own colour
		return a.Modules.Environment.GetNearestLootboxByColour(a.GetColour())
	}
}

// decides the force the biker is going to pedal with
func (a *FYPAgent) DecideForce(direction uuid.UUID) {
	if direction == uuid.Nil {
		return
	}

	if a.Modules.Environment.IsAwdiNear() {
		// Move in opposite direction to Awdi in full force
		bikePos, awdiPos := a.Modules.Environment.GetBikeById(a.GetBike()).GetPosition(), a.Modules.Environment.GetAwdi().GetPosition()
		force := a.Modules.Environment.GetForcesToTargetWithDirectionOffset(utils.BikerMaxForce, 1.0-a.Modules.Environment.GetBikeOrientation(), bikePos, awdiPos)
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

	// get the final forces and set them
	lootboxPosition := a.Modules.Environment.GetLootboxPos(lootboxID)
	force := a.Modules.Environment.GetForcesToTargetWithDirectionOffset(utils.BikerMaxForce, -a.Modules.Environment.GetBikeOrientation(), agentPosition, lootboxPosition)
	a.SetForces(force)
}

// returns: map containing bikerID -> distribution (i.e. share of resources)
func (a *FYPAgent) DecideAllocation() voting.Allocation {

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

// returns: the biker on your bike with the minimum trust
func (a *FYPAgent) GetTeammateWithMinTrust() modules.IDTrustPair {
	fellowBikers := a.GetFellowBikersSlice()
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
func (a *FYPAgent) GetSumOfTrustOnBike() float64 {
	var sum = 0.0
	for _, teammate := range a.GetFellowBikersSlice() {
		sum += a.Modules.AgentParameters.TrustNetwork[teammate.GetID()]
	}
	return sum
}

// returns: float representing average trust of your bike
func (a *FYPAgent) GetAverageTrustOnBike() float64 {

	// Prevent divide
	if len(a.GetFellowBikersSlice()) == 0 {
		return 0.5
	}

	sum := a.GetSumOfTrustOnBike()

	return sum / float64(len(a.GetFellowBikers()))
}

// sets the bike the agent is on
func (a *FYPAgent) SetBike(bikeId uuid.UUID) {
	a.Modules.Environment.BikeId = bikeId
	a.BaseBiker.SetBike(bikeId)
}

// delete the agent with the specified id from this agents trust map
func (a *FYPAgent) HandleAgentUnalive(id uuid.UUID) {
	delete(a.Modules.AgentParameters.TrustNetwork, id)
}

// ----- Not currently used but kept for possible implementation later -----

func (a *FYPAgent) GetTrustOfAgent(agent objects.IBaseBiker) float64 {
	return a.Modules.AgentParameters.TrustNetwork[agent.GetID()]
}

// returns: map of UUID -> {0,1} for an agent where 0 means 'don't kick' and 1 means 'do kick'
func (a *FYPAgent) VoteForKickout() map[uuid.UUID]int {
	VoteMap := make(map[uuid.UUID]int)

	// check all bikers on the bike but ignore ourselves
	for _, agent := range a.GetFellowBikersSlice() {
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
func (a *FYPAgent) DecideKickOut() []uuid.UUID {
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
