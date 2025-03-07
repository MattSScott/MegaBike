package server

import (
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/physics"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

func (s *Server) RunRoundLoop(iterationDump *SimplifiedIterationDump, round int) {

	// ----- 1. Bikes decide on their movement direction and force -----

	s.runActionDeliberation(objects.MoveBike)
	s.runActionDeliberation(objects.KickAgent)
	s.RunDirectionDecisionProcess()

	// ----- 2. Move objects in the world ----- 

	// Move the mega bikes
	for _, bike := range s.megaBikes {
		// update mass dependent on number of agents on bike
		bike.UpdateMass()
		s.runActionDeliberation(objects.Lootbox)
		s.MovePhysicsObject(bike)
	}

	// Move the awdi
	s.MovePhysicsObject(s.awdi)


	// ----- 3. Distributing energy from any collected lootboxes -----

	s.runActionDeliberation(objects.Allocation)
	s.LootboxCheckAndDistributions()

	// ----- 4. Punish and Kill -----

	// Punish bikeless agents
	s.punishBikelessAgents()

	// Check Awdi collision and kill agents if they collided
	s.AwdiCollisionCheck()

	// kill agents that run out of energy
	s.unaliveAgents()



	// ----- 5. Recording events, cleanup and replenishing -----

	// round dump code
	roundDump := s.GenerateRoundDump()
	iterationDump.AddRoundToIteration(roundDump)



	// if the representative(s) die then re-select them
	for _, bike := range s.GetMegaBikes() {
		agents := bike.GetAgents()

		if len(agents) != 0 {
			reps := bike.GetRepresentatives()

			// iterate over reps, and if theyre dead then replace them.
			for deadRepIdx, repID := range reps {
				if _, ok := s.deadAgents[repID]; ok {
					s.HandleDepartingRepresentative(bike, deadRepIdx)
				}
			}
			
		}
	}

	// Replenish and reset
	if utils.ReplenishLootBoxes {
		s.replenishLootBoxes()
	}
	if utils.ReplenishMegaBikes {
		s.replenishMegaBikes()
	}
	for _, bike := range s.GetMegaBikes() {
		bike.ResetCurrentPool()
	}

	// Allow agents to gossip
	s.RunMessagingSession()

}

func (s *Server) runActionDeliberation(action objects.Action) {
	for _, bike := range s.megaBikes {
		if *globals.StratifyRules {
			bike.ActionIsValidForRuleset(action)
		} else {
			bike.ActionCompliesWithLinearRuleset()
		}
	}
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
		case utils.PerfectDemocracy, utils.DegenerateDemocracy:
			direction = s.RunDemocraticAction(bike)

			// agents incur an energetic penalty for participating in a vote
			for _, agent := range agents {
				agent.UpdateEnergyLevel(-utils.DeliberativeDemocracyPenalty)
			}

		case utils.PerfectAristocracy, utils.DegenerateAristocracy, utils.PerfectMonarchy, utils.DegenerateMonarchy:
			direction = s.RunRepresentativeAction(bike)
			// not voting here so no negative energy for now
		}

		for _, agent := range agents {
			agent.DecideForce(direction)
			// deplete energy
			energyLost := agent.GetForces().Pedal * utils.MovingDepletion
			agent.UpdateEnergyLevel(-energyLost)
		}
	}
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

	// checks how many bikes have looted one lootbox to split it between them
	looted := make(map[uuid.UUID]int)
	for _, megabike := range s.GetMegaBikes() {
		for lootid, lootbox := range s.GetLootBoxes() {
			if megabike.CheckForCollision(lootbox) { // && len(megabike.GetAgents()) != 0
				megabike.UpdateCurrentPool(lootbox.GetTotalResources())
				if value, ok := looted[lootid]; ok {
					looted[lootid] = value + 1
				} else {
					looted[lootid] = 1
				}
			}
		}
	}

	for bikeid, megabike := range s.GetMegaBikes() {
		for lootid, lootbox := range s.GetLootBoxes() {
			if megabike.CheckForCollision(lootbox) {
				agents := megabike.GetAgents()
				totAgents := len(agents)

				if totAgents > 0 {
					gov := s.GetMegaBikes()[bikeid].GetGovernance()
					var winningAllocation voting.IdVoteMap


					switch gov {
					case utils.PerfectDemocracy, utils.DegenerateDemocracy:
						allAllocations := make(map[uuid.UUID]voting.IdVoteMap)
						for _, agent := range agents {
							// the agents return their ideal lootbox split by assigning a number between 0 and 1 to
							// each biker on their bike (including themselves) ensuring they sum to 1
							allAllocations[agent.GetID()] = agent.DecideAllocation()
						}

						Iallocations := make(map[uuid.UUID]voting.IVoter)
						for i, v := range allAllocations {
							Iallocations[i] = v
						}
						// make weights of 1 for all agents
						weights := make(map[uuid.UUID]float64)
						for _, agent := range agents {
							weights[agent.GetID()] = 1.0
						}
						winningAllocation = voting.CumulativeDist(Iallocations, weights)

					case utils.DegenerateMonarchy, utils.PerfectMonarchy:
						reps := megabike.GetRepresentatives()
						monarch := s.GetAgentMap()[reps[0]]
						winningAllocation = monarch.DecideRepresentativeAllocation(gov)
						
					case utils.DegenerateAristocracy, utils.PerfectAristocracy:
						// for now just choose an aristocrat randomly to decide
						// TODO: find a way to aggregate the aristocrat distributions
						reps := megabike.GetRepresentatives()
						randIndex := rand.Intn(len(reps))
						chosenAristocrat := s.GetAgentMap()[reps[randIndex]]
						winningAllocation = chosenAristocrat.DecideRepresentativeAllocation(gov)
						
					}
					

					bikeShare := float64(looted[lootid]) // how many other bikes have looted this box

					for agentID, allocation := range winningAllocation {
						lootShare := allocation * (lootbox.GetTotalResources() / bikeShare)
						agent, ok := s.GetAgentMap()[agentID]
						if !ok {
							continue
						}
						// Update agent energy level based on their share of the loot
						agent.UpdateEnergyLevel(lootShare)
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
	// Check collision for awdi with any megaBike
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
func (s *Server) unaliveAgents() {
	for _, agent := range s.GetAgentMap() {
		if agent.GetEnergyLevel() <= 0 {
			fmt.Printf("Agent %s ran out of energy \n", utils.TranslateToName(agent.GetID()))
			s.RemoveAgent(agent)
		}
	}
}