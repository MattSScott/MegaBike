package server

import (
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/physics"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

func (s *Server) RunRoundLoop(iterationDump *SimplifiedIterationDump, round int) {

	// ----- 1. Bikes decide on their movement direction and force -----

	s.runActionDeliberation(objects.MoveBike)
	s.runActionDeliberation(objects.KickAgent)
	s.RunActionProcess()

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

	// if the representative(s) die then re-elect them
	for _, bike := range s.GetMegaBikes() {
		gov := bike.GetGovernance()
		agents := bike.GetAgents()
		if len(agents) != 0 && (gov == utils.PerfectMonarchy || gov == utils.DegenerateMonarchy) {
			reps := bike.GetRepresentatives()
			
			// if bike leader dead, reassign leadership
			if _, ok := s.deadAgents[reps[0]]; ok {
				agents := bike.GetAgents()
				reps := s.RepresentativeElection(agents, gov)
				bike.SetRepresentatives(reps)
			}
		} else if len(agents) != 0 && (gov == utils.PerfectAristocracy || gov == utils.DegenerateAristocracy) {
			reps := bike.GetRepresentatives()

			// for now, just relect the whole aristocracy if someone is dead
			// TODO 1) just replace the dead person 2) make sure it continues if there are fewer than 3 reps
			for _, id := range(reps){
				if _, ok := s.deadAgents[id]; ok {
					agents := bike.GetAgents()
					reps := s.RepresentativeElection(agents, gov)
					bike.SetRepresentatives(reps)
					break
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

// handles the kick out process according to each bike's governance
func (s *Server) HandleKickoutProcess() []uuid.UUID {
	allKicked := make([]uuid.UUID, 0)
	for _, bike := range s.GetMegaBikes() {
		agents := bike.GetAgents()

		// if bike.GetRuler() == uuid.Nil {
		// 	continue
		// }

		if len(agents) != 0 {

			agentsVotes := make([]uuid.UUID, 0)

			
			// TODO: edit this so the kickout process changes based on governance. 
			// currently all are the same except monarchies
			switch bike.GetGovernance() {
			case utils.PerfectDemocracy:
				// make map of weights of 1 for all agents on bike (as they all have the same voting power)
				agents := bike.GetAgents()
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get which agents are getting kicked out
				agentsVotes = bike.KickOutAgent(weights)

			case utils.DegenerateDemocracy:
				// make map of weights of 1 for all agents on bike (as they all have the same voting power)
				agents := bike.GetAgents()
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get which agents are getting kicked out
				agentsVotes = bike.KickOutAgent(weights)

			case utils.PerfectAristocracy:
				// make map of weights of 1 for all agents on bike (as they all have the same voting power)
				agents := bike.GetAgents()
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get which agents are getting kicked out
				agentsVotes = bike.KickOutAgent(weights)
			
			case utils.DegenerateAristocracy:
				// make map of weights of 1 for all agents on bike (as they all have the same voting power)
				agents := bike.GetAgents()
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get which agents are getting kicked out
				agentsVotes = bike.KickOutAgent(weights)
			
			case utils.PerfectMonarchy:
				monarch := s.GetAgentMap()[bike.GetRepresentatives()[0]]
				agentsVotes = monarch.DecideKickOut()

			case utils.DegenerateMonarchy:
				monarch := s.GetAgentMap()[bike.GetRepresentatives()[0]]
				//TODO: add another agent function below like 'decide kickout malevolently'
				agentsVotes = monarch.DecideKickOut()
			}

			// perform kickout
			repKickedOut := false
			allKicked = append(allKicked, agentsVotes...)
			for _, agentID := range agentsVotes {
				s.RemoveAgentFromBike(s.GetAgentMap()[agentID])
				// if the leader was kicked out will need to vote for a new one
				if slices.Contains(bike.GetRepresentatives(), agentID) {
					repKickedOut = true
				}
			}

			// new elections if needed. yet again needs editing to say 'only re-elect the kicked out reps'
			if repKickedOut && len(bike.GetAgents()) != 0 {
				reps := s.RepresentativeElection(bike.GetAgents(), bike.GetGovernance())
				bike.SetRepresentatives(reps)
			}
		}

	}
	return allKicked
}

// get list of agents that want to leave their bike in current round
func (s *Server) GetLeavingDecisions() []uuid.UUID {
	leavingAgents := make([]uuid.UUID, 0)

	for agentId, agent := range s.GetAgentMap() {
		if agent.GetBikeStatus() {

			agent.UpdateAgentInternalState() // is this necessary
			
			switch agent.DecideAction() {
			case objects.Pedal:
				continue
			case objects.ChangeBike:
				// the bike id is set to be the desired bike and onbike is set to false
				// so by looking at the values of onBike and megaBikeID it will be known
				// whether the agent is trying to join a bike (and which one)

				// the request is handled at the beginning of the next round, so the moving
				// will only be finalised then
				leavingAgents = append(leavingAgents, agentId)
				s.RemoveAgentFromBike(agent)
			default:
				panic("agent decided invalid action")
			}
		}
	}

	// if ruler has left the bike will need to run elections
	// again edit for aristocracy case where still some aristocrats
	for _, bike := range s.GetMegaBikes() {
		reps := bike.GetRepresentatives()
		for _, id := range reps {
			if slices.Contains(leavingAgents, id) && len(bike.GetAgents()) != 0 {
				reps := s.RepresentativeElection(bike.GetAgents(), bike.GetGovernance())
				bike.SetRepresentatives(reps)
			}
	}
	}
	return leavingAgents
}

// dispatch joining requests to the bikes of competence and move bikers from limbo to their desired bike subject to the
// acceptance process outcome
func (s *Server) ProcessJoiningRequests(inLimbo []uuid.UUID) {

	// -------------------------- PROCESS JOINING REQUESTS -------------------------
	// 1. group agents that have onBike = false by the bike they are trying to join
	bikeRequests := s.GetJoiningRequests(inLimbo)
	// panic(s.megaBikes)

	// 2. pass to agents on each of the desired bikes a list of all agents trying to join
	for bikeID, pendingAgents := range bikeRequests {
		bike := s.megaBikes[bikeID]
		agents := bike.GetAgents()
		fmt.Println(len(agents))
		// if there are no agents on the target bike accept all of them (until all seats are filled)
		if len(agents) == 0 {
			// as iterating over a map is pseudo-random it's enough to stop whrn the capacity is reached
			// to ensure a fair (= random) selection in the case of an empty target bike
			for i, pendingAgent := range pendingAgents {
				if i <= utils.BikersOnBike {
					acceptedAgent := s.GetAgentMap()[pendingAgent]
					s.AddAgentToBike(acceptedAgent, bike)
				} else {
					break
				}
			}
		} else {
			acceptedRanked := make([]uuid.UUID, 0)

			// the acceptance process is different for each governance type
			// TODO: customise these for each type
			switch bike.GetGovernance() {

			case utils.PerfectDemocracy:
				// make map of weights of 1 for all agents on bike
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get approval votes from each agent
				responses := make(map[uuid.UUID]map[uuid.UUID]bool, len(agents)) // list containing all the agents' ranking
				for _, agent := range agents {
					responses[agent.GetID()] = agent.DecideJoining(pendingAgents)
				}

				// accept agents based on the response outcome (only capacity-n bikers can be accepted)
				acceptedRanked = voting.GetAcceptanceRanking(responses, weights)
			case utils.DegenerateDemocracy:
				// make map of weights of 1 for all agents on bike
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get approval votes from each agent
				responses := make(map[uuid.UUID]map[uuid.UUID]bool, len(agents)) // list containing all the agents' ranking
				for _, agent := range agents {
					responses[agent.GetID()] = agent.DecideJoining(pendingAgents)
				}

				// accept agents based on the response outcome (only capacity-n bikers can be accepted)
				acceptedRanked = voting.GetAcceptanceRanking(responses, weights)

			case utils.PerfectAristocracy:
				// make map of weights of 1 for all agents on bike
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get approval votes from each agent
				responses := make(map[uuid.UUID]map[uuid.UUID]bool, len(agents)) // list containing all the agents' ranking
				for _, agent := range agents {
					responses[agent.GetID()] = agent.DecideJoining(pendingAgents)
				}

				// accept agents based on the response outcome (only capacity-n bikers can be accepted)
				acceptedRanked = voting.GetAcceptanceRanking(responses, weights)
			case utils.DegenerateAristocracy:
				// make map of weights of 1 for all agents on bike
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get approval votes from each agent
				responses := make(map[uuid.UUID]map[uuid.UUID]bool, len(agents)) // list containing all the agents' ranking
				for _, agent := range agents {
					responses[agent.GetID()] = agent.DecideJoining(pendingAgents)
				}

				// accept agents based on the response outcome (only capacity-n bikers can be accepted)
				acceptedRanked = voting.GetAcceptanceRanking(responses, weights)
			case utils.PerfectMonarchy:
				// // get the map of weights from the leader
				// leader := s.GetAgentMap()[bike.GetRuler()]
				// weights := leader.DecideWeights(utils.Joining)

				// // get approval votes from each agent
				// responses := make(map[uuid.UUID](map[uuid.UUID]bool), len(agents)) // list containing all the agents' ranking
				// for _, agent := range agents {
				// 	responses[agent.GetID()] = agent.DecideJoining(pendingAgents)
				// }

				// // accept agents based on the response outcome (only capacity-n bikers can be accepted)
				// // so the ranking is sorted based on how many people voted positively for each agent
				// acceptedRanked = voting.GetAcceptanceRanking(responses, weights)
				monarch := s.GetAgentMap()[bike.GetRepresentatives()[0]]
				acceptedRankedMap := monarch.DecideJoining(pendingAgents)
				for agentID, accepted := range acceptedRankedMap {
					if accepted {
						acceptedRanked = append(acceptedRanked, agentID)
					}
				}
			case utils.DegenerateMonarchy:
				monarch := s.GetAgentMap()[bike.GetRepresentatives()[0]]
				acceptedRankedMap := monarch.DecideJoining(pendingAgents)
				for agentID, accepted := range acceptedRankedMap {
					if accepted {
						acceptedRanked = append(acceptedRanked, agentID)
					}
				}
			}

			// run acceptance process
			totalSeatsFilled := len(agents)
			emptySpaces := utils.BikersOnBike - totalSeatsFilled

			// accept up to capacity
			for i := 0; i < min(emptySpaces, len(acceptedRanked)); i++ {
				accepted := acceptedRanked[i]
				acceptedAgent := s.GetAgentMap()[accepted]
				s.AddAgentToBike(acceptedAgent, bike)
			}
		}
	}
}

// run the process on deciding this round's direction according to each governance's rules and on deciding the forces
func (s *Server) RunActionProcess() {

	for _, bike := range s.GetMegaBikes() {

		agents := bike.GetAgents()
		if len(agents) == 0 {
			continue
		}

		// get the direction for this round (either the voted on or what's decided by the representatives)
		// TODO: edit degenerate and perfect democracy so that they are different
		var direction uuid.UUID
		governance := bike.GetGovernance()
		switch governance {
		case utils.PerfectDemocracy:
			// make map of weights of 1 for all agents on bike
			weights := make(map[uuid.UUID]float64)
			for _, agent := range agents {
				weights[agent.GetID()] = 1.0
			}

			direction = s.RunDemocraticAction(bike, weights)
			// agents incur an energetic penalty for participating in a vote
			for _, agent := range agents {
				agent.UpdateEnergyLevel(-utils.DeliberativeDemocracyPenalty)
			}
		case utils.DegenerateDemocracy:
			// make map of weights of 1 for all agents on bike
			weights := make(map[uuid.UUID]float64)
			for _, agent := range agents {
				weights[agent.GetID()] = 1.0
			}

			direction = s.RunDemocraticAction(bike, weights)
			// agetns incur in an energetic penalty for partecipating in a vote
			for _, agent := range agents {
				agent.UpdateEnergyLevel(-utils.DeliberativeDemocracyPenalty)
			}
		case utils.PerfectAristocracy, utils.DegenerateAristocracy, utils.PerfectMonarchy, utils.DegenerateMonarchy:
			direction = s.RunRepresentativeAction(bike)
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

// check for deadly collisions
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

// if a bike has looted a box run the distribution process according to the governance type
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
				// Collision detected
				// fmt.Println("GOT ONE !!!")
				agents := megabike.GetAgents()
				totAgents := len(agents)

				if totAgents > 0 {
					gov := s.GetMegaBikes()[bikeid].GetGovernance()
					var winningAllocation voting.IdVoteMap
					switch gov {
					case utils.PerfectDemocracy, utils.DegenerateDemocracy, utils.DegenerateAristocracy, utils.DegenerateMonarchy, utils.PerfectAristocracy, utils.PerfectMonarchy:
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

					}
					// case utils.PerfectMonarchy:
					// 	// get the map of weights from the leader
					// 	leader, ok := s.GetAgentMap()[megabike.GetRuler()]
					// 	if !ok {
					// 		break
					// 	}
					// 	weights := leader.DecideWeights(utils.Allocation)
					// outer:
					// 	for id := range weights {
					// 		for _, agent := range agents {
					// 			if agent.GetID() == id {
					// 				continue outer
					// 			}
					// 		}
					// 		panic("leader gave weight to an agent that isn't on the bike")
					// 	}
					// 	// get allocation votes from each agent
					// 	allAllocations := make(map[uuid.UUID]voting.IdVoteMap)
					// 	for _, agent := range agents {
					// 		allAllocations[agent.GetID()] = agent.DecideAllocation()
					// 	}

					// 	Iallocations := make(map[uuid.UUID]voting.IVoter)
					// 	for i, v := range allAllocations {
					// 		Iallocations[i] = v
					// 	}
					// 	winningAllocation = voting.CumulativeDist(Iallocations, weights)

					// case utils.DegenerateMonarchy:
					// 	// dictator decides the allocation
					// 	leader := s.GetAgentMap()[megabike.GetRuler()]
					// 	winningAllocation = leader.DecideDictatorAllocation()
					

					bikeShare := float64(looted[lootid]) // how many other bikes have looted this box

					for agentID, allocation := range winningAllocation {
						lootShare := allocation * (lootbox.GetTotalResources() / bikeShare)
						agent, ok := s.GetAgentMap()[agentID]
						if !ok {
							continue
						}
						// Allocate loot based on the calculated utility share
						// fmt.Println(agent.GetEnergyLevel())
						agent.UpdateEnergyLevel(lootShare)
						// fmt.Println(agent.GetEnergyLevel())
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
	for id, loot := range looted {
		if loot > 0 {
			delete(s.lootBoxes, id)
		}
	}
}

func (s *Server) SetDestinationBikes() {
	for _, agent := range s.GetAgentMap() {
		if !agent.GetBikeStatus() {
			targetBike := agent.ChangeBike()
			if targetBike == uuid.Nil { // agent didn't specify bike
				continue
			}
			if _, ok := s.megaBikes[targetBike]; !ok {
				panic("agent requested a bike that doesn't exist")
			}
			agent.SetBike(targetBike)
		}
	}
}

func (s *Server) unaliveAgents() {
	for _, agent := range s.GetAgentMap() {
		if agent.GetEnergyLevel() <= 0 {
			// fmt.Printf("Agent %s got game ended\n", id)
			s.RemoveAgent(agent)
		}
	}
}

func (s *Server) punishBikelessAgents() {
	for id, agent := range s.GetAgentMap() {
		if _, ok := s.megaBikeRiders[id]; !ok {
			// Agent is not on a bike
			agent.UpdateEnergyLevel(-utils.LimboEnergyPenalty)
		}
	}
}
