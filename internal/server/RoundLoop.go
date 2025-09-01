package server

import (
	"MegabikeFYPVersion/internal/common/objects"
	"MegabikeFYPVersion/internal/common/physics"
	"MegabikeFYPVersion/internal/common/utils"
	"MegabikeFYPVersion/internal/common/voting"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

func (s *Server) RunRoundLoop(iterationDump *SimplifiedIterationDump, round int) {

	// ----- 1. Bikes decide on their movement direction and force -----

	s.RunDirectionDecisionProcess()

	// ----- 2. Move objects in the world -----

	// Move the megabikes
	for _, bike := range s.megaBikes {
		bike.UpdateMass()
		s.MovePhysicsObject(bike)
	}

	// Move the awdi
	// s.MovePhysicsObject(s.awdi)

	// ----- 3. Distributing energy from any collected lootboxes -----
	s.LootboxCheckAndDistributions()

	// ----- 4. Punish and Kill -----

	// Punish bikeless agents
	s.punishBikelessAgents()

	// Check Awdi collision and kill agents if they collided
	s.AwdiCollisionCheck()

	// kill agents that run out of energy
	s.terminateAgents()

	// ----- 5. Recording events, cleanup and replenishing -----

	roundDump := s.GenerateRoundDump()
	iterationDump.AddRoundToIteration(roundDump)

	// handle the case where reps die
	s.HandleDeadRepresentatives()

	// Replenish and reset
	if utils.ReplenishLootBoxes {
		s.replenishLootBoxes()
	}
	if utils.ReplenishMegaBikes {
		s.replenishMegaBikes()
	}

	// ----- 6. Allow the agents to gossip about the round that has just passed -----

	s.RunAgentMessagingSession(false)

}

// run the process on deciding this round's direction for each megabike
func (s *Server) RunDirectionDecisionProcess() {

	for _, bike := range s.GetMegaBikes() {
		agents := bike.GetAgents()

		if len(agents) == 0 {
			continue
		}

		// get the direction for this round (either democratically or what's decided by the representatives)
		var direction uuid.UUID
		governance := bike.GetGovernance()

		switch governance {
		case utils.Many:
			direction = s.RunManyDirectionDecision(bike)

			// agents incur an energetic penalty for participating in a vote
			for _, agent := range agents {
				agent.UpdateEnergyLevel(-utils.DecisionPenalty)
			}

		case utils.Some:
			direction = s.RunSomeDirectionDecision(bike)

			// reps incur an energetic penalty
			for _, repID := range bike.GetRepresentatives() {
				s.GetAgentMap()[repID].UpdateEnergyLevel(-utils.DecisionPenalty)
			}

		case utils.One:
			direction = s.RunOneDirectionDecision(bike)

			// reps incur an energetic penalty
			for _, repID := range bike.GetRepresentatives() {
				s.GetAgentMap()[repID].UpdateEnergyLevel(-utils.DecisionPenalty)
			}

		}

		// let agents decide the force they are going to pedal with
		for _, agent := range agents {
			agent.DecideForce(direction)
			agent.SetRoundForces(agent.GetForces())
			energyLost := agent.GetForces().Pedal * utils.MovingDepletion
			agent.UpdateEnergyLevel(-energyLost)
		}
	}
}

// returns: uuid of lootbox to aim toward (i.e. direction) for current round from the "one" rep
func (s *Server) RunOneDirectionDecision(bike objects.IMegaBike) uuid.UUID {
	agents := s.GetAgentMap()
	reps := bike.GetRepresentatives()
	var direction uuid.UUID

	monarch := agents[reps[0]]
	direction = monarch.DecideDirection()
	monarch.SetRoundDirection(direction)
	return direction
}

// returns: uuid of lootbox to aim toward (i.e. direction) for current round from the "some" reps
func (s *Server) RunSomeDirectionDecision(bike objects.IMegaBike) uuid.UUID {
	agents := s.GetAgentMap()
	reps := bike.GetRepresentatives()
	var direction uuid.UUID

	var suggestedDirections []uuid.UUID
	countsPerDirection := make(map[uuid.UUID]int)

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
}

// returns: uuid of lootbox to aim toward (i.e. direction) for current round from the agents
func (s *Server) RunManyDirectionDecision(bike objects.IMegaBike) uuid.UUID {

	agents := bike.GetAgents()
	proposedDirections := make(map[uuid.UUID]uuid.UUID) // maps agent id to their proposed lootbox id

	if len(bike.GetAgents()) == 0 {
		return uuid.Nil
	}

	// Fill in the proposed directions map by asking each agent to propose a direction.
	for _, agent := range agents {
		if agent.GetBikeStatus() {
			proposedDirection := agent.DecideDirection()
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

	// Option one: go for most voted lootbox
	var mostVotedLootbox uuid.UUID
	highestCount := 0

	for lootbox, count := range counts {
		if count > highestCount {
			highestCount = count
			mostVotedLootbox = lootbox
		}
	}

	return mostVotedLootbox
}

// move the physics objects (i.e. mega bikes and awdi) according to the forces and orientations
func (s *Server) MovePhysicsObject(po objects.IPhysicsObject) {

	// Server requests to update their force and orientation based on agents pedaling
	po.UpdateForce()
	force := po.GetForce()
	po.UpdateOrientation()
	orientation := po.GetOrientation()
	// Obtains the current state (i.e. velocity, acceleration, position, mass)
	initialState := po.GetPhysicalState()

	// Generates a new state based on the force and orientation
	finalState := physics.GenerateNewState(initialState, force, orientation)

	// Sets the new physical state (i.e. updates gamestate)
	po.SetPhysicalState(finalState)
}

// if a bike has looted a box, run the distribution process according to the governance type
func (s *Server) LootboxCheckAndDistributions() {

	// 1. create a map of lootboxID -> number of bikes that have looted it
	looted := make(map[uuid.UUID]int)
	for _, megabike := range s.GetMegaBikes() {
		for lootid, lootbox := range s.GetLootBoxes() {
			if megabike.CheckForCollision(lootbox) {
				if value, ok := looted[lootid]; ok {
					looted[lootid] = value + 1
				} else {
					looted[lootid] = 1
				}
			}
		}
	}

	// 2. For each bike: go over all lootboxes, and for those they have looted, get the agents to decide an allocation
	// then update each agents energy by allocating them their share of their bikes share of the total loot.
	for bikeid, megabike := range s.GetMegaBikes() {
		for lootid, lootbox := range s.GetLootBoxes() {
			if megabike.CheckForCollision(lootbox) {
				agents := megabike.GetAgents()

				if len(agents) > 0 {
					gov := s.GetMegaBikes()[bikeid].GetGovernance()
					var winningAllocation voting.Allocation

					switch gov {
					case utils.Many:

						// map of agentID -> map of fellow bikers and their distribution
						allAllocations := make(map[uuid.UUID]voting.Allocation)

						for _, agent := range agents {
							// the agents return their ideal lootbox split by assigning a number between 0 and 1 to
							// each biker on their bike (including themselves) ensuring they sum to 1
							allAllocations[agent.GetID()] = agent.DecideAllocation()
						}

						// make weights of 1 for all agents
						weights := make(map[uuid.UUID]float64)
						for _, agent := range agents {
							weights[agent.GetID()] = 1.0
						}

						winningAllocation = voting.AggregateAllocations(allAllocations, weights)

					case utils.Some:

						reps := megabike.GetRepresentatives()
						agentMap := s.GetAgentMap()

						aristocratAllocations := make(map[uuid.UUID]voting.Allocation)
						for _, repID := range reps {
							// the reps return their ideal lootbox split by assigning a number between 0 and 1 to
							// each biker on their bike (including themselves) ensuring they sum to 1
							aristocratAllocations[repID] = agentMap[repID].DecideAllocation()
						}

						// make map of weights of 1 for all reps (redundant but fine for now)
						weights := make(map[uuid.UUID]float64)
						for _, repID := range reps {
							weights[repID] = 1.0
						}

						winningAllocation = voting.AggregateAllocations(aristocratAllocations, weights)

					case utils.One:
						reps := megabike.GetRepresentatives()
						monarch := s.GetAgentMap()[reps[0]]
						winningAllocation = monarch.DecideAllocation()
					}

					numBikesSharingLootbox := float64(looted[lootid])

					for agentID, allocation := range winningAllocation {
						lootShare := allocation * (lootbox.GetTotalResources() / numBikesSharingLootbox)
						agent, ok := s.GetAgentMap()[agentID]
						if !ok {
							continue
						}

						// Update agent energy level based on their share of the loot
						agent.UpdateEnergyLevel(lootShare)
						agent.UpdateIterationIncome(lootShare)
						// Allocate points if the box is of the right colour
						if agent.GetColour() == lootbox.GetColour() {
							agent.UpdatePoints(utils.PointsFromSameColouredLootBox)
						}
					}
				}
			}
		}
	}

	// despawn lootboxes that have been looted
	for lootboxID, numBikesLooted := range looted {
		if numBikesLooted > 0 {
			delete(s.lootBoxes, lootboxID)
		}
	}
}

// give energy penalty to bikeless agents
func (s *Server) punishBikelessAgents() {
	for id, agent := range s.GetAgentMap() {
		if _, ok := s.megaBikeRiders[id]; !ok {
			// Agent is not on a bike
			agent.UpdateEnergyLevel(-utils.LimboEnergyPenalty)
		}
	}
}

// check for deadly collisions with the awdi
func (s *Server) AwdiCollisionCheck() {

	for _, megabike := range s.GetMegaBikes() {
		if s.awdi.CheckForCollision(megabike) {
			// Collision detected
			for _, agentToDelete := range megabike.GetAgents() {
				s.RemoveAgent(agentToDelete)
			}
			if utils.AwdiRemovesMegaBike {
				delete(s.megaBikes, megabike.GetID())
			}
		}
	}
}

// kill agents that are out of energy
func (s *Server) terminateAgents() {
	for _, agent := range s.GetAgentMap() {
		if agent.GetEnergyLevel() <= 0 {
			fmt.Printf("Agent %s ran out of energy \n", utils.TranslateToName(agent.GetID()))
			s.RemoveAgent(agent)
		}
	}
}

// remove and replace any dead representatives
func (s *Server) HandleDeadRepresentatives() {

	for _, bike := range s.GetMegaBikes() {
		agents := bike.GetAgents()
		if len(agents) == 0 {
			continue
		}

		if bike.GetGovernance() == utils.Many {
			continue
		}

		reps := bike.GetRepresentatives()
		var survivingReps []uuid.UUID

		for _, repID := range reps {
			if _, ok := s.deadAgents[repID]; ok {
				continue
			} else {
				survivingReps = append(survivingReps, repID)
			}
		}

		var expectedNumReps int
		if bike.GetGovernance() == utils.One {
			expectedNumReps = 1
		} else if bike.GetGovernance() == utils.Some {
			expectedNumReps = 3
		}
		
		agentsOnBike := bike.GetAgents()
		for _, agent := range agentsOnBike {
			if !slices.Contains(survivingReps, agent.GetID()) && len(survivingReps) < expectedNumReps {
				survivingReps = append(survivingReps, agent.GetID())
			}
		}
		bike.SetRepresentatives(survivingReps)
	}
}
