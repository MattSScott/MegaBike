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

func (s *Server) RunRoundLoop(iterationDump *SimplifiedIterationDump) {
	// get destination bikes from bikers not on bike
	s.runActionDeliberation(objects.MoveBike)
	s.SetDestinationBikes()
	// take care of agents that want to leave the bike and of the acceptance/ expulsion process
	s.runActionDeliberation(objects.KickAgent)
	// get the direction decisions and pedalling forces
	s.RunActionProcess()
	// The Awdi makes a decision

	// Move the mega bikes
	for _, bike := range s.megaBikes {
		// update mass dependent on number of agents on bike
		bike.UpdateMass()
		s.runActionDeliberation(objects.Lootbox)
		s.MovePhysicsObject(bike)
	}

	// Move the awdi
	s.MovePhysicsObject(s.awdi)

	// Lootbox Distribution
	s.runActionDeliberation(objects.Allocation)
	s.LootboxCheckAndDistributions()

	// Punish bikeless agents
	s.punishBikelessAgents()

	// Check if agents died
	// Check Awdi collision
	s.AwdiCollisionCheck()
	s.unaliveAgents()

	roundDump := s.GenerateRoundDump()
	iterationDump.AddRoundToIteration(roundDump)

	// if the leader dies hold new elections
	for _, bike := range s.GetMegaBikes() {
		gov := bike.GetGovernance()
		agents := bike.GetAgents()
		if len(agents) != 0 && (gov == utils.Leadership || gov == utils.Dictatorship) {
			ruler := bike.GetRuler()
			if _, ok := s.deadAgents[ruler]; ok {
				agents := bike.GetAgents()
				ruler := s.RulerElection(agents, gov)
				bike.SetRuler(ruler)
			}
		}
	}

	// Replenish objects
	if utils.ReplenishLootBoxes {
		s.replenishLootBoxes()
	}

	if utils.ReplenishMegaBikes {
		s.replenishMegaBikes()
	}

	// Run the messaging session
	s.RunMessagingSession()

	for _, bike := range s.GetMegaBikes() {
		bike.ResetCurrentPool()
	}
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

// handles bikers leaving the bike, potential kick outs and the acceptance process (in this order)
func (s *Server) RunBikeSwitch() {
	inLimbo := make([]uuid.UUID, 0)

	// check if agents want ot leave the bike on this round
	changeBike := s.GetLeavingDecisions()
	inLimbo = append(inLimbo, changeBike...)

	//process the kickout request
	kickedOff := s.HandleKickoutProcess()
	inLimbo = append(inLimbo, kickedOff...)

	// process the joining request
	s.ProcessJoiningRequests(inLimbo)

}

// handles the kick out process according to each bike's governance
func (s *Server) HandleKickoutProcess() []uuid.UUID {
	allKicked := make([]uuid.UUID, 0)
	for _, bike := range s.GetMegaBikes() {
		agents := bike.GetAgents()

		if bike.GetRuler() == uuid.Nil {
			continue
		}

		if len(agents) != 0 {

			agentsVotes := make([]uuid.UUID, 0)

			// the kickout process only happens through a (possibly weighted) vote in deliberative democracy and leadership democracy
			switch bike.GetGovernance() {
			case utils.Democracy:
				// make map of weights of 1 for all agents on bike (as they all have the same voting power)
				agents := bike.GetAgents()
				weights := make(map[uuid.UUID]float64)
				for _, agent := range agents {
					weights[agent.GetID()] = 1.0
				}

				// get which agents are getting kicked out
				agentsVotes = bike.KickOutAgent(weights)

			case utils.Leadership:
				// get the map of weights from the leader
				ruler := bike.GetRuler()
				leader := s.GetAgentMap()[ruler]
				weights := leader.DecideWeights(utils.Kickout)
				// get which agents are getting kicked out
				agentsVotes = bike.KickOutAgent(weights)

			case utils.Dictatorship:
				// in a dictatorship only the ruler can kick out people
				dictator := s.GetAgentMap()[bike.GetRuler()]
				agentsVotes = dictator.DecideKickOut()
			}

			// perform kickout
			leaderKickedOut := false
			allKicked = append(allKicked, agentsVotes...)
			for _, agentID := range agentsVotes {
				s.RemoveAgentFromBike(s.GetAgentMap()[agentID])
				// if the leader was kicked out will need to vote for a new one
				if agentID == bike.GetRuler() {
					leaderKickedOut = true
				}
			}

			// new elections if needed
			if leaderKickedOut && len(bike.GetAgents()) != 0 && bike.GetGovernance() == utils.Leadership {
				ruler := s.RulerElection(bike.GetAgents(), utils.Leadership)
				bike.SetRuler(ruler)
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
			agent.UpdateAgentInternalState()
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
	for _, bike := range s.GetMegaBikes() {
		if slices.Contains(leavingAgents, bike.GetRuler()) && len(bike.GetAgents()) != 0 {
			ruler := s.RulerElection(bike.GetAgents(), utils.Leadership)
			bike.SetRuler(ruler)
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
			// if the governance of the bike is ruler led an election needs to be held
			// TODO!!!
		} else {
			acceptedRanked := make([]uuid.UUID, 0)

			// the acceptance process is different for each governance type
			switch bike.GetGovernance() {
			case utils.Democracy:
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
			case utils.Leadership:
				// get the map of weights from the leader
				leader := s.GetAgentMap()[bike.GetRuler()]
				weights := leader.DecideWeights(utils.Joining)

				// get approval votes from each agent
				responses := make(map[uuid.UUID](map[uuid.UUID]bool), len(agents)) // list containing all the agents' ranking
				for _, agent := range agents {
					responses[agent.GetID()] = agent.DecideJoining(pendingAgents)
				}

				// accept agents based on the response outcome (only capacity-n bikers can be accepted)
				// so the ranking is sorted based on how many people voted positively for each agent
				acceptedRanked = voting.GetAcceptanceRanking(responses, weights)
			case utils.Dictatorship:
				dictator := s.GetAgentMap()[bike.GetRuler()]
				acceptedRankedMap := dictator.DecideJoining(pendingAgents)
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

		// get the direction for this round (either the voted on or what's decided by the leader/ dictator)
		var direction uuid.UUID
		electedGovernance := bike.GetGovernance()
		switch electedGovernance {
		case utils.Democracy:
			// make map of weights of 1 for all agents on bike
			weights := make(map[uuid.UUID]float64)
			for _, agent := range agents {
				weights[agent.GetID()] = 1.0
			}

			direction = s.RunDemocraticAction(bike, weights)
			// agetns incur in an energetic penalty for partecipating in a vote
			for _, agent := range agents {
				// fmt.Println("DEMO LOSS:", agent.GetForces().Pedal, agent.GetEnergyLevel())
				agent.UpdateEnergyLevel(-utils.DeliberativeDemocracyPenalty)
				// fmt.Println(agent.GetEnergyLevel())
			}
		case utils.Leadership:
			// get weights from leader
			leader, ok := s.GetAgentMap()[bike.GetRuler()]
			if !ok {
				break
			}
			weights := leader.DecideWeights(utils.Direction)
			direction = s.RunDemocraticAction(bike, weights)
			for _, agent := range agents {
				// fmt.Println("LEADER LOSS")
				agent.UpdateEnergyLevel(-utils.LeadershipDemocracyPenalty)
			}
		case utils.Dictatorship:
			// the dictator is solely responsible for choosing the direction
			direction = s.RunRulerAction(bike)
		}

		for _, agent := range agents {
			agent.DecideForce(direction)
			// fmt.Println("SUBFUNC INIT:", agent.GetEnergyLevel())
			// deplete energy
			energyLost := agent.GetForces().Pedal * utils.MovingDepletion
			// fmt.Println("BASE LOSS")
			agent.UpdateEnergyLevel(-energyLost)
			// fmt.Println("FINAL LOSS:", agent.GetEnergyLevel())
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
				agents := megabike.GetAgents()
				totAgents := len(agents)

				if totAgents > 0 {
					gov := s.GetMegaBikes()[bikeid].GetGovernance()
					var winningAllocation voting.IdVoteMap
					switch gov {
					case utils.Democracy:
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

					case utils.Leadership:
						// get the map of weights from the leader
						leader, ok := s.GetAgentMap()[megabike.GetRuler()]
						if !ok {
							break
						}
						weights := leader.DecideWeights(utils.Allocation)
					outer:
						for id := range weights {
							for _, agent := range agents {
								if agent.GetID() == id {
									continue outer
								}
							}
							panic("leader gave weight to an agent that isn't on the bike")
						}
						// get allocation votes from each agent
						allAllocations := make(map[uuid.UUID]voting.IdVoteMap)
						for _, agent := range agents {
							allAllocations[agent.GetID()] = agent.DecideAllocation()
						}

						Iallocations := make(map[uuid.UUID]voting.IVoter)
						for i, v := range allAllocations {
							Iallocations[i] = v
						}
						winningAllocation = voting.CumulativeDist(Iallocations, weights)

					case utils.Dictatorship:
						// dictator decides the allocation
						leader := s.GetAgentMap()[megabike.GetRuler()]
						winningAllocation = leader.DecideDictatorAllocation()
					}

					bikeShare := float64(looted[lootid]) // how many other bikes have looted this box

					for agentID, allocation := range winningAllocation {
						lootShare := allocation * (lootbox.GetTotalResources() / bikeShare)
						agent := s.GetAgentMap()[agentID]
						// Allocate loot based on the calculated utility share
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
		if agent.GetEnergyLevel() < 0 {
			// fmt.Printf("Agent %s got game ended\n", id)
			s.RemoveAgent(agent)
		}
	}
}

func (s *Server) punishBikelessAgents() {
	for id, agent := range s.GetAgentMap() {
		if _, ok := s.megaBikeRiders[id]; !ok {
			// Agent is not on a bike
			agent.UpdateEnergyLevel(utils.LimboEnergyPenalty)
		}
	}
}
