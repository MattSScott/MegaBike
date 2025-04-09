package agent

import (
	"SOMAS2023/internal/clients/teamSOSA/modules"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"math"
	"math/rand"
	"runtime"

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

// returns: slice of agents to kick out (currently just does one)
func (a *AgentSOSA) DecideKickOut() []uuid.UUID {
	// Only called when the agent is a representative agent.
	// We kick out the agent with the lowest trust on the bike.
	// GetBikerWithMinTrust returns only one agent, if more agents with min Trust, it randomly chooses one.
	kickOut_agents := make([]uuid.UUID, 0)
	agentIDStruct := a.Modules.Environment.GetBikerWithMinTrust(a.Modules.AgentParameters)
	agentId := agentIDStruct.ID
	if agentId != uuid.Nil {
		kickOut_agents = append(kickOut_agents, agentId)
	}
	return kickOut_agents
}

// returns: map of pending agents uuid -> {true, false} where true means accept and false means dont accept
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

// returns: uuid of the target bike if they want to change, otherwise just return current bike id 
func (a *AgentSOSA) DecideChangeBike() uuid.UUID {
	shouldChangeBike := false
	targetBikeID := uuid.Nil
	if a.Modules.AgentParameters.GetAverageTrust() < modules.StayOnBikeThreshold {
		shouldChangeBike = true
		targetBikeID = a.Modules.Environment.GetBikeWithMaximumTrust(a.Modules.AgentParameters)
	}

	if shouldChangeBike {
		return targetBikeID
	} else {
		return a.Modules.Environment.BikeId
	}
}

// ----- Decisions (Round level) -----

// returns: lootbox uuid to aim towards (the direction) from the choice of a subset of lootboxes
func (a *AgentSOSA) ProposeDirectionFromSubset(subset map[uuid.UUID]objects.ILootBox) uuid.UUID {
	agentColour, agentEnergy := a.GetColour(), a.GetEnergyLevel()
	optimalLootbox := a.Modules.Environment.GetNearestLootboxByColourFromSubset(agentColour, subset)
	nearestLootbox := a.Modules.Environment.GetNearestLootboxFromSubset(subset)
	if agentEnergy < modules.EnergyToOptimalLootboxThreshold || optimalLootbox == uuid.Nil {
		return nearestLootbox
	}
	return optimalLootbox
}

// returns: uuid of lootbox to aim towards (i.e. the direction). decided in a selfless way. only called when the agent is perfect monarch / aristocrat
func (a *AgentSOSA) DecideDirectionBenevolently() uuid.UUID {
	// Move in opposite direction to Awdi in full force
	if a.Modules.Environment.IsAwdiNear() {
		// fmt.Printf("[DictateDirection] Agent %s is near Awdi\n", a.GetID())
		return a.Modules.Environment.GetNearestLootboxAwayFromAwdi()
	}
	// Otherwise, move towards the lootbox with the highest gain
	return a.Modules.Environment.GetHighestGainLootbox()
}

// returns: uuid of lootbox to aim towards (i.e. the direction). decided in a selfish way, only called when agent is degen monarch / aristocrat
func (a *AgentSOSA) DecideDirectionMalevolently() uuid.UUID {

	// Move in opposite direction to Awdi in full force - a representative still doesn't want to get obliterated
	if a.Modules.Environment.IsAwdiNear() {
		// fmt.Printf("[DictateDirection] Agent %s is near Awdi\n", a.GetID())
		return a.Modules.Environment.GetNearestLootboxAwayFromAwdi()
	}

	// Otherwise, move towards the nearest lootbox of your own colour
	return a.Modules.Environment.GetNearestLootboxByColour(a.GetColour())
}

// decides the force the biker is going to pedal with (untouched)
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

	return allocation
}

// returns: map containing bikerID -> distribution (i.e. share of resources) (only called by reps)
func (a *AgentSOSA) DecideRepresentativeAllocation(governance utils.Governance) voting.IdVoteMap {

	socialCapital := a.DecideAllocation()
	// Calculate the total social capital
	totalSocialCapital := 0.0
	for _, sc := range socialCapital {
		totalSocialCapital += sc
	}

	if governance == utils.PerfectMonarchy || governance == utils.PerfectAristocracy {
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
	} else if governance == utils.DegenerateMonarchy  || governance == utils.DegenerateAristocracy {
		// same as perfect monarchy, but cut every elses share by 50% and give yourself the rest.
		ownID := a.GetID()
		shareDistributed := float64(0)
		
		result := make(voting.IdVoteMap)
		for agentID, sc := range socialCapital {
			if agentID != ownID {
				agentShare:= sc*0.5 / totalSocialCapital
				result[agentID] = agentShare
				shareDistributed += agentShare
				if math.IsNaN(result[agentID]) {
					runtime.Breakpoint()
					panic("fuck")
				}
			}
		}

		monarchShare := 1 - shareDistributed
		result[ownID] = monarchShare

		return result
	} else {
		panic("deciding rep allocation gone wrong")
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