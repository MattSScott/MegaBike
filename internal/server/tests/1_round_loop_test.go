package server_test

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/server"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Direction Decision 

func TestRunOneDirectionDecision(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	// perform role assignment, then get them to decide on a direction. if nil, something wrong. otherwise pass!
	for _, bike := range s.GetMegaBikes(){
		governance := bike.GetGovernance()
		if governance == utils.Some || governance == utils.One {
			s.PerformRoleAssignment(bike)
			lootbox := s.RunOneDirectionDecision(bike)
			fmt.Printf("\n bike targetting lootbox with id %v", lootbox)
			if lootbox == uuid.Nil {
				t.Error("targetting nil lootbox")
			}
		}
	}

	fmt.Println("Representative direction decision process runs successfully")
}

func TestRunDemocraticDirectionDecision(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)


	for _, bike := range s.GetMegaBikes(){
		governance := bike.GetGovernance()
		if governance == utils.Many {
			lootbox := s.RunDemocraticDirectionDecision(bike)
			fmt.Printf("\n bike targetting lootbox with id %v", lootbox) 
			if lootbox == uuid.Nil {
				t.Error("targetting nil lootbox")
			}
		}
	}

	fmt.Println("Democratic direction decision process runs successfully")
}

// Lootbox Stuff
func TestLootboxCheckAndDistributionOne(t *testing.T) {
	iterations := 1
	s := server.GenerateServer()
	s.Initialize(iterations)

	for _, bike := range s.GetMegaBikes() {
		s.PerformRoleAssignment(bike)
	}

	// Get the One bike
	foundOne := false
	var bike objects.IMegaBike
	for !foundOne {
		bike = s.GetMegaBikes()[s.GetRandomBikeId()]
		if bike.GetGovernance() == utils.One {
			foundOne = true
		}
	}

	// Perform role assignment
	s.PerformRoleAssignment(bike)

	fmt.Println("Bike has id", bike.GetID(), "with governance", bike.GetGovernance(), "and ", len(bike.GetAgents()), "agents")
	// Set all agent energys to 0.5 to ensure they are able to get lootbox energy, as they start at 1 and can't go higher
	// without doing this they cant actually get any energy
	for _, agent := range bike.GetAgents() {
		agent.UpdateEnergyLevel(-0.5)
	}

	// find the sum of agent energys before collecting the box
	sumOfAgentEnergyBeforeCollision := 0.0
	for _, agent := range bike.GetAgents() {
		sumOfAgentEnergyBeforeCollision += agent.GetEnergyLevel()
	}

	fmt.Println("sum of agent energy pre collision:", sumOfAgentEnergyBeforeCollision)

	// getting a random lootbox's position 
	var lootbox objects.ILootBox
	for _, lootbox = range s.GetLootBoxes() {
		break
	}
	pos := lootbox.GetPosition()

	// change the bikes position to this lootboxes position 
	ps := bike.GetPhysicalState()
	newPhysicalState := utils.PhysicalState{
		Position:     pos,
		Velocity:     ps.Velocity,
		Mass:         ps.Mass,
		Acceleration: ps.Acceleration,
	}
	bike.SetPhysicalState(newPhysicalState)

	// run lootbox check and distribution
	s.LootboxCheckAndDistributions()

	// find the sum of agent energys AFTER collecting the box
	sumOfAgentEnergyAfterCollision := 0.0
	for _, agent := range bike.GetAgents() {
		sumOfAgentEnergyAfterCollision += agent.GetEnergyLevel()
	}

	fmt.Println("sum of agent energy post collision:", sumOfAgentEnergyAfterCollision)

	// Check the following:
	assert.True(t, sumOfAgentEnergyAfterCollision > sumOfAgentEnergyBeforeCollision, "lootbox energy has not been distributed ")
	
	if _, exists := s.GetLootBoxes()[lootbox.GetID()]; exists {
		t.Error("The collected lootbox hasn't been deleted from the lootbox map")
	}

}

func TestLootboxCheckAndDistributionSome(t *testing.T) {
	iterations := 1
	s := server.GenerateServer()
	s.Initialize(iterations)

	for _, bike := range s.GetMegaBikes() {
		s.PerformRoleAssignment(bike)
	}

	// Get the Some bike
	foundSome := false
	var bike objects.IMegaBike
	for !foundSome {
		bike = s.GetMegaBikes()[s.GetRandomBikeId()]
		if bike.GetGovernance() == utils.Some {
			foundSome = true
		}
	}

	fmt.Println("Bike has id", bike.GetID(), "with governance", bike.GetGovernance(), "and ", len(bike.GetAgents()), "agents")
	// Set all agent energys to 0.5 to ensure they are able to get lootbox energy, as they start at 1 and can't go higher
	// without doing this they cant actually get any energy
	for _, agent := range bike.GetAgents() {
		agent.UpdateEnergyLevel(-0.5)
	}

	// find the sum of agent energys before collecting the box
	sumOfAgentEnergyBeforeCollision := 0.0
	for _, agent := range bike.GetAgents() {
		sumOfAgentEnergyBeforeCollision += agent.GetEnergyLevel()
	}

	fmt.Println("sum of agent energy pre collision:", sumOfAgentEnergyBeforeCollision)

	// getting a random lootbox's position 
	var lootbox objects.ILootBox
	for _, lootbox = range s.GetLootBoxes() {
		break
	}
	pos := lootbox.GetPosition()

	// change the bikes position to this lootboxes position 
	ps := bike.GetPhysicalState()
	newPhysicalState := utils.PhysicalState{
		Position:     pos,
		Velocity:     ps.Velocity,
		Mass:         ps.Mass,
		Acceleration: ps.Acceleration,
	}
	bike.SetPhysicalState(newPhysicalState)

	// run lootbox check and distribution
	s.LootboxCheckAndDistributions()

	// find the sum of agent energys AFTER collecting the box
	sumOfAgentEnergyAfterCollision := 0.0
	for _, agent := range bike.GetAgents() {
		sumOfAgentEnergyAfterCollision += agent.GetEnergyLevel()
	}

	fmt.Println("sum of agent energy post collision:", sumOfAgentEnergyAfterCollision)

	// Check the following:
	assert.True(t, sumOfAgentEnergyAfterCollision > sumOfAgentEnergyBeforeCollision, "lootbox energy has not been distributed ")
	
	if _, exists := s.GetLootBoxes()[lootbox.GetID()]; exists {
		t.Error("The collected lootbox hasn't been deleted from the lootbox map")
	}

}

func TestLootboxCheckAndDistributionMany(t *testing.T) {
	iterations := 1
	s := server.GenerateServer()
	s.Initialize(iterations)

	for _, bike := range s.GetMegaBikes() {
		s.PerformRoleAssignment(bike)
	}


	// Get the Many bike
	foundMany := false
	var bike objects.IMegaBike
	for !foundMany {
		bike = s.GetMegaBikes()[s.GetRandomBikeId()]
		if bike.GetGovernance() == utils.Many {
			foundMany = true
		}
	}

	fmt.Println("Bike has id", bike.GetID(), "with governance", bike.GetGovernance(), "and ", len(bike.GetAgents()), "agents")
	// Set all agent energys to 0.5 to ensure they are able to get lootbox energy, as they start at 1 and can't go higher
	// without doing this they cant actually get any energy
	for _, agent := range bike.GetAgents() {
		agent.UpdateEnergyLevel(-0.5)
	}

	// find the sum of agent energys before collecting the box
	sumOfAgentEnergyBeforeCollision := 0.0
	for _, agent := range bike.GetAgents() {
		sumOfAgentEnergyBeforeCollision += agent.GetEnergyLevel()
	}

	fmt.Println("sum of agent energy pre collision:", sumOfAgentEnergyBeforeCollision)

	// getting a random lootbox's position 
	var lootbox objects.ILootBox
	for _, lootbox = range s.GetLootBoxes() {
		break
	}
	pos := lootbox.GetPosition()

	// change the bikes position to this lootboxes position 
	ps := bike.GetPhysicalState()
	newPhysicalState := utils.PhysicalState{
		Position:     pos,
		Velocity:     ps.Velocity,
		Mass:         ps.Mass,
		Acceleration: ps.Acceleration,
	}
	bike.SetPhysicalState(newPhysicalState)

	// run lootbox check and distribution
	s.LootboxCheckAndDistributions()

	// find the sum of agent energys AFTER collecting the box
	sumOfAgentEnergyAfterCollision := 0.0
	for _, agent := range bike.GetAgents() {
		sumOfAgentEnergyAfterCollision += agent.GetEnergyLevel()
	}

	fmt.Println("sum of agent energy post collision:", sumOfAgentEnergyAfterCollision)

	// Check the following:
	assert.True(t, sumOfAgentEnergyAfterCollision > sumOfAgentEnergyBeforeCollision, "lootbox energy has not been distributed ")
	
	if _, exists := s.GetLootBoxes()[lootbox.GetID()]; exists {
		t.Error("The collected lootbox hasn't been deleted from the lootbox map")
	}

}

// Awdi Collision Stuff
func TestAwdiCollisionProcess(t *testing.T) {
	iterations := 2
	s := server.GenerateServer()
	s.Initialize(iterations)
	nAgentToDelete := 0
	for _, anyBike := range s.GetMegaBikes() {
		agentsOnBike := anyBike.GetAgents()
		nAgentToDelete = len(agentsOnBike)
		if agentsOnBike == nil {
			continue
		}
		// find any non-empty bike
		if len(agentsOnBike) > 0 {
			// send awdi to it
			s.GetAwdi().SetPhysicalState(utils.PhysicalState{Position: anyBike.GetPosition()})
		}
	}
	nAgentsBefore := len(s.GetAgentMap())
	nMegaBikesBefore := len(s.GetMegaBikes())
	s.AwdiCollisionCheck()
	nAgentsAfter := len(s.GetAgentMap())
	nMegaBikesAfter := len(s.GetMegaBikes())

	// check if remove agents correctly
	if nAgentsBefore-nAgentsAfter != nAgentToDelete {
		fmt.Printf("Before awdi collision, number of agents = %d \n", nAgentsBefore)
		fmt.Printf("After awdi collision, number of agents = %d \n", nAgentsAfter)
		fmt.Printf("On bike collide with awdi, number of agents = %d \n", nAgentToDelete)
		t.Error("Awdi didnt remove agents correctly")
	}

	if utils.AwdiRemovesMegaBike {
		// check if remove megaBike correctly
		if nMegaBikesBefore-nMegaBikesAfter != 1 {
			fmt.Printf("Before awdi collision, number of megaBikes = %d \n", nMegaBikesBefore)
			fmt.Printf("After awdi collision, number of megaBikes = %d \n", nMegaBikesAfter)
			t.Error("Awdi didnt remove megaBike correctly")
		}
	}
	fmt.Printf("\nRun action process passed \n")
}

func TestAwdiTargeting(t *testing.T) {
	iterations := 1
	s := server.GenerateServer()
	s.Initialize(iterations)
	i := 0
	emptyBikeId := uuid.UUID{}
	slowBikeId := uuid.UUID{}
	for id, bike := range s.GetMegaBikes() {
		if i == 0 {
			// remove agents on it
			for _, agents := range bike.GetAgents() {
				bike.RemoveAgent(agents.GetID())
			}
			// stop the bike
			bike.SetPhysicalState(utils.PhysicalState{Velocity: 0.0})
			fmt.Printf("Megabike{%s} has {%d} agents with velocity {%.2f}\n", id, len(bike.GetAgents()), bike.GetVelocity())
			emptyBikeId = id
		} else if i == 1 {
			// give the bike a slow Velocity
			bike.SetPhysicalState(utils.PhysicalState{Velocity: 1.0})
			fmt.Printf("Megabike{%s} has {%d} agents with velocity {%.2f}\n", id, len(bike.GetAgents()), bike.GetVelocity())
			slowBikeId = id
		} else if i == 2 {
			// give the bike a fast Velocity
			bike.SetPhysicalState(utils.PhysicalState{Velocity: 100.0})
			agentsOnBike := bike.GetAgents()
			if len(agentsOnBike) == 0 {
				emptyBikeId = id
			}
			fmt.Printf("Megabike{%s} has {%d} agents with velocity {%.2f}\n", id, len(bike.GetAgents()), bike.GetVelocity())
		}
		i += 1
	}

	s.GetAwdi().UpdateForce()
	targetId := s.GetAwdi().GetTargetID()
	fmt.Printf("Awdi is targeting {%s}\n", targetId)
	if utils.AwdiTargetsEmptyMegaBike {
		if targetId != emptyBikeId {
			t.Error("Awdi didnt target empty megaBike!")
		}
	}
	if utils.AwdiOnlyTargetsStationaryMegaBike {
		if targetId == slowBikeId {
			t.Error("Awdi didnt ignore moving megaBike!")
		}
	}
	fmt.Printf("\nRun action process passed \n")
}

// this test needs writing
func TestHandleDepartingRepresentative(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)


	
}


















// func TestHandleKickout(t *testing.T) {
// 	iterations := 6
// 	s := server.GenerateServer()
// 	s.Initialize(iterations)
	
// 	s.HandleKickoutProcess()

// 	for _, agent := range s.GetAgentMap() {
// 		if !agent.GetBikeStatus() {
// 			for _, bike := range s.GetMegaBikes() {
// 				for _, agentOnBike := range bike.GetAgents() {
// 					if agentOnBike.GetID() == agent.GetID() {
// 						t.Error("leaving agent is on a bike when it shouldn't be")

// 					}
// 				}
// 			}
// 		}
// 	}
// 	fmt.Printf("\nHandle kickout passed \n")
// }



// func TestRunActionProcess(t *testing.T) {

// 	*globals.BikerAgentCount = 16
// 	globals.MegaBikeCount = 2

// 	for i := 0; i < 10; i++ {
// 		iterations := 1
// 		s := server.GenerateServer()
// 		s.Initialize(iterations)
// 		// required otherwise agents are not initialized to bikes
// 		// s.FoundingInstitutions()

// 		for _, bike := range s.GetMegaBikes() {
// 			fmt.Println(len(bike.GetAgents()))
// 		}

// 		// Loop through each bike
// 		for _, bike := range s.GetMegaBikes() {
// 			// Randomly select a governance strategy for this bike
// 			governanceTypes := []int{int(utils.Democracy), int(utils.Leadership), int(utils.Dictatorship)}
// 			governance := utils.Governance(governanceTypes[rand.Intn(len(governanceTypes))])
// 			// bike.SetGovernance(governance)
// 			bike.SetGovernance(governance)

// 			// Randomly select a ruler if necessary
// 			if governance != utils.Democracy {
// 				agents := bike.GetAgents()
// 				if len(agents) > 0 {
// 					randIndex := rand.Intn(len(agents))
// 					randomAgent := agents[randIndex]
// 					bike.SetRuler(randomAgent.GetID())
// 				} else {
// 					t.Error("Bike is empty")
// 				}
// 			}
// 		}

// 		s.RunActionProcess()
// 		// check all agents have lost energy (proportionally to how much they have pedalled)
// 		for _, agent := range s.GetAgentMap() {
// 			lostEnergy := (utils.MovingDepletion * agent.GetForces().Pedal)

// 			allBikes := s.GetMegaBikes()
// 			agentBike := allBikes[agent.GetBike()]

// 			governance := agentBike.GetGovernance()
// 			switch governance {
// 			case utils.Democracy:
// 				lostEnergy += utils.DeliberativeDemocracyPenalty
// 			case utils.Leadership:
// 				lostEnergy += utils.LeadershipDemocracyPenalty
// 			default:
// 			}
// 			if agent.GetEnergyLevel()+lostEnergy != 1 {
// 				// fmt.Println(agent.GetForces().Pedal*utils.MovingDepletion, agentBike.GetGovernance())
// 				// fmt.Println("NEW:", agent.GetEnergyLevel(), lostEnergy, agent.GetEnergyLevel()+lostEnergy, agent.GetBikeStatus())
// 				fmt.Println(agent.GetID(), agent.GetBike())
// 			}
// 			// FP precision error
// 			if (agent.GetEnergyLevel() - (1.0 - lostEnergy)) > utils.Epsilon {
// 				t.Error("agents energy hasn't been successfully depleted! expected lost energy: ", lostEnergy, "actual lost energy: ", 1.0-agent.GetEnergyLevel(), "difference: ", (1.0-agent.GetEnergyLevel())-lostEnergy)
// 			}
// 		}
// 	}
// 	fmt.Printf("\nRun action process passed \n")
// }

// func TestRunActionProcessDictator(t *testing.T) {
// 	iterations := 1
// 	s := server.GenerateServer()
// 	s.Initialize(iterations)
// 	// required otherwise agents are not initialized to bikes
// 	// s.FoundingInstitutions()

// 	// make bike with dictatorship (by getting one of the existing bikes and making it a dictatorship bike)
// 	foundBike := false
// 	var bikeID uuid.UUID
// 	for !foundBike {
// 		bikeID = s.GetRandomBikeId()
// 		if len(s.GetMegaBikes()[bikeID].GetAgents()) != 0 {
// 			foundBike = true
// 		}
// 	}
// 	bike := s.GetMegaBikes()[bikeID]

// 	bike.SetGovernance(utils.Dictatorship)
// 	agents := bike.GetAgents()
// 	if len(agents) == utils.BikersOnBike {
// 		removable := agents[0]
// 		bike.RemoveAgent(removable.GetID())
// 	}

// 	// place dictator on bike
// 	dictator := NewNegativeAgent(s)
// 	s.AddAgent(dictator)
// 	dictator.SetBike(bikeID)
// 	bike.AddAgent(dictator)
// 	bike.SetRuler(dictator.GetID())

// 	// run the action process and confirm the direction is that of the dictator
// 	s.RunActionProcess()

// 	// check that the direction for the bike with our dictator is the same as the dictator's
// 	for _, bike := range s.GetMegaBikes() {
// 		if bike.GetID() == bikeID {
// 			dictatorDirection := dictator.DictateDirection()
// 			dictator.DecideForce(dictatorDirection)
// 			dictatorForce := dictator.GetForces()
// 			for _, agent := range bike.GetAgents() {
// 				if agent.GetID() == dictator.GetID() {
// 					assert.Equal(t, dictatorForce, agent.GetForces())
// 				}
// 			}
// 		}
// 	}
// }

// func TestRunActionProcessLeader(t *testing.T) {
// 	iterations := 1
// 	s := server.GenerateServer()
// 	s.Initialize(iterations)
// 	// required otherwise agents are not initialized to bikes

// 	// s.FoundingInstitutions()

// 	// make bike with dictatorship (by getting one of the existing bikes and making it a dictatorship bike)
// 	foundBike := false
// 	var bikeID uuid.UUID
// 	for !foundBike {
// 		bikeID = s.GetRandomBikeId()
// 		if len(s.GetMegaBikes()[bikeID].GetAgents()) != 0 {
// 			foundBike = true
// 		}
// 	}
// 	bike := s.GetMegaBikes()[bikeID]
// 	bike.SetGovernance(utils.Leadership)
// 	agents := bike.GetAgents()
// 	if len(agents) == utils.BikersOnBike {
// 		removable := agents[0]
// 		bike.RemoveAgent(removable.GetID())
// 	}

// 	// place dictator on bike
// 	leader := NewNegativeAgent(s)
// 	s.AddAgent(leader)
// 	leader.SetBike(bikeID)
// 	bike.AddAgent(leader)
// 	bike.SetRuler(leader.GetID())

// 	s.RunActionProcess() // DEBUG

// 	// check that the direction of the leader is that of its direction (as the weights emulate a dictatorship)
// 	for _, bike := range s.GetMegaBikes() {
// 		if bike.GetID() == bikeID {
// 			leaderDirection := leader.ProposeDirection()
// 			leader.DecideForce(leaderDirection)
// 			leaderForce := leader.GetForces()
// 			for _, agent := range bike.GetAgents() {
// 				if agent.GetID() == leader.GetID() {
// 					assert.Equal(t, leaderForce, agent.GetForces())
// 				}
// 			}
// 		}
// 	}

// }

// func TestProcessJoiningRequestsWithLimbo(t *testing.T) { // debug
// 	globals.MegaBikeCount += 2 // accomodate two spare bikes
// 	iterations := 3
// 	s := server.GenerateServer()
// 	s.Initialize(iterations)

// 	// 1: get two bike ids
// 	targetBikes := make([]uuid.UUID, 2)

// 	i := 0
// 	for bikeId, bike := range s.GetMegaBikes() {
// 		if len(bike.GetAgents()) == 0 {
// 			targetBikes[i] = bikeId
// 			i++
// 		}
// 	}

// 	// 2: set one agent requesting the first bike and two other requesting the second one
// 	i = 0
// 	requests := make(map[uuid.UUID][]uuid.UUID)
// 	requests[targetBikes[0]] = make([]uuid.UUID, 1)
// 	requests[targetBikes[1]] = make([]uuid.UUID, 1)
// 	limbo := make([]uuid.UUID, 1)
// 	for _, agent := range s.GetAgentMap() {
// 		if i == 0 {
// 			if agent.GetBikeStatus() {
// 				agent.ToggleOnBike()
// 			}
// 			agent.SetBike(targetBikes[0])
// 			requests[targetBikes[0]][0] = agent.GetID()
// 		} else if i == 1 {
// 			if agent.GetBikeStatus() {
// 				agent.ToggleOnBike()
// 			}
// 			// add it to second bike for request
// 			agent.SetBike(targetBikes[1])
// 			requests[targetBikes[1]][0] = agent.GetID()
// 		} else if i == 2 {
// 			if agent.GetBikeStatus() {
// 				agent.ToggleOnBike()
// 			}
// 			//remove it from bike but add it to limbo (to mimick request made in this turn)
// 			agent.SetBike(targetBikes[1])
// 			limbo[0] = agent.GetID()
// 		} else {
// 			break
// 		}

// 		i += 1
// 	}

// 	// all agents should be accepted as there should be enough room on all bikes (but make it subject to that)
// 	// check that all of them are now on bikes
// 	// check that there are no bikers left with on bike = false
// 	s.ProcessJoiningRequests(limbo)

// 	for bikeID, agents := range requests {
// 		bike := s.GetMegaBikes()[bikeID]
// 		for _, agent := range agents {
// 			onBike := false
// 			for _, agentOnBike := range bike.GetAgents() {
// 				onBikeId := agentOnBike.GetID()
// 				if onBikeId == agent {
// 					onBike = true
// 					assert.True(t, agentOnBike.GetBikeStatus(), "biker's status wasn't successfully toggled back")
// 					break
// 				}
// 			}
// 			assert.True(t, onBike, "biker wasn't successfully accepted on bike")
// 		}
// 	}
// 	// check that the limbo agent is not on any bikes
// 	for _, agentID := range limbo {
// 		for _, bike := range s.GetMegaBikes() {
// 			for _, agentOnBike := range bike.GetAgents() {
// 				assert.NotEqual(t, agentOnBike.GetID() == agentID, "agent in limbo was accepted")
// 			}
// 		}
// 	}
// 	// check that the limbo agent is still in limbo
// 	for _, agentID := range limbo {
// 		agent := s.GetAgentMap()[agentID]
// 		assert.Equal(t, agent.GetBikeStatus(), false, "agent in limbo was accepted")
// 	}

// 	fmt.Printf("\nProcess joining request passed \n")
// }















// func TestGetWinningDirection1(t *testing.T) {
// 	iterations := 1
// 	s := server.GenerateServer()
// 	s.Initialize(iterations)

// 	fullPower := make([]uuid.UUID, 2)
// 	for i := 0; i < 2; i++ {
// 		fullPower[i] = uuid.New()
// 	}

// 	reducedPower := make([]uuid.UUID, 3)
// 	for i := 0; i < 3; i++ {
// 		reducedPower[i] = uuid.New()
// 	}

// 	// make the weights list
// 	weights := make(map[uuid.UUID]float64)
// 	for _, agent := range fullPower {
// 		weights[agent] = 1.0
// 	}
// 	for _, agent := range reducedPower {
// 		weights[agent] = 0.5
// 	}

// 	// make the votes list
// 	proposals := make(map[uuid.UUID]voting.LootboxVoteMap)
// 	fullPowerProposal := uuid.New()
// 	reducedPowerProposal := uuid.New()

// 	fullPowerVote := make(voting.LootboxVoteMap)
// 	fullPowerVote[fullPowerProposal] = 1.0
// 	fullPowerVote[reducedPowerProposal] = 0.0

// 	reducedPowerVote := make(voting.LootboxVoteMap)
// 	reducedPowerVote[fullPowerProposal] = 0.0
// 	reducedPowerVote[reducedPowerProposal] = 1.0

// 	for _, agent := range fullPower {
// 		proposals[agent] = fullPowerVote
// 	}
// 	for _, agent := range reducedPower {
// 		proposals[agent] = reducedPowerVote
// 	}

// 	assert.Equal(t, fullPowerProposal, s.GetWinningDirection(proposals, weights), "full power proposal should win")
// }

// func TestGetWinningDirection2(t *testing.T) {
// 	iterations := 1
// 	s := server.GenerateServer()
// 	s.Initialize(iterations)

// 	fullPower := make([]uuid.UUID, 2)
// 	for i := 0; i < 2; i++ {
// 		fullPower[i] = uuid.New()
// 	}

// 	reducedPower := make([]uuid.UUID, 3)
// 	for i := 0; i < 3; i++ {
// 		reducedPower[i] = uuid.New()
// 	}

// 	// make the weights list
// 	weights := make(map[uuid.UUID]float64)
// 	for _, agent := range fullPower {
// 		weights[agent] = 1.0
// 	}
// 	// in this case they all have the same power, so the reducedPower one should win
// 	for _, agent := range reducedPower {
// 		weights[agent] = 1.0
// 	}

// 	// make the votes list
// 	proposals := make(map[uuid.UUID]voting.LootboxVoteMap)
// 	fullPowerProposal := uuid.New()
// 	reducedPowerProposal := uuid.New()

// 	fullPowerVote := make(voting.LootboxVoteMap)
// 	fullPowerVote[fullPowerProposal] = 1.0
// 	fullPowerVote[reducedPowerProposal] = 0.0

// 	reducedPowerVote := make(voting.LootboxVoteMap)
// 	reducedPowerVote[fullPowerProposal] = 0.0
// 	reducedPowerVote[reducedPowerProposal] = 1.0

// 	for _, agent := range fullPower {
// 		proposals[agent] = fullPowerVote
// 	}
// 	for _, agent := range reducedPower {
// 		proposals[agent] = reducedPowerVote
// 	}

// 	assert.Equal(t, reducedPowerProposal, s.GetWinningDirection(proposals, weights), "reduced power proposal should win")
// }



// func TestProcessJoiningRequests(t *testing.T) {
// 	globals.MegaBikeCount += 2 // accomodate two spare bikes
// 	OnlySpawnBaseBikers(t)
// 	iterations := 3
// 	s := server.GenerateServer()
// 	s.Initialize(iterations)

// 	// 1: get two bike ids (choose the 2 most empty bikes)
// 	bikes := make([]objects.IMegaBike, 0, len(s.GetMegaBikes()))
// 	for _, b := range s.GetMegaBikes() {
// 		bikes = append(bikes, b)
// 	}
// 	slices.SortFunc(bikes, func(a, b objects.IMegaBike) int {
// 		return cmp.Compare(len(a.GetAgents()), len(b.GetAgents()))
// 	})
// 	targetBikes := []uuid.UUID{bikes[0].GetID(), bikes[1].GetID()}

// 	// 2: set one agent requesting the first bike and two other requesting the second one
// 	i := 0
// 	requests := make(map[uuid.UUID][]uuid.UUID)
// 	requests[targetBikes[0]] = make([]uuid.UUID, 1)
// 	requests[targetBikes[1]] = make([]uuid.UUID, 2)
// 	for _, agent := range s.GetAgentMap() {
// 		if i == 0 {
// 			if agent.GetBikeStatus() {
// 				agent.ToggleOnBike()
// 			}
// 			agent.SetBike(targetBikes[0])
// 			requests[targetBikes[0]][0] = agent.GetID()
// 		} else if i <= 2 {
// 			if agent.GetBikeStatus() {
// 				agent.ToggleOnBike()
// 			}
// 			agent.SetBike(targetBikes[1])
// 			requests[targetBikes[1]][i-1] = agent.GetID()
// 		} else {
// 			break
// 		}
// 		i += 1
// 	}

// 	// all agents should be accepted as there should be enough room on all bikes (but make it subject to that)
// 	// check that all of them are now on bikes
// 	// check that there are no bikers left with on bike = false
// 	s.ProcessJoiningRequests(make([]uuid.UUID, 0))

// 	for bikeID, agents := range requests {
// 		bike := s.GetMegaBikes()[bikeID]
// 		for _, agent := range agents {
// 			onBike := false
// 			for _, agentOnBike := range bike.GetAgents() {
// 				onBikeId := agentOnBike.GetID()
// 				if onBikeId == agent {
// 					onBike = true
// 					if !agentOnBike.GetBikeStatus() {
// 						t.Error("biker's status wasn't successfully toggled back")
// 					}
// 					break
// 				}
// 			}
// 			if !onBike {
// 				t.Error("biker wasn't successfully accepted on bike")
// 			}
// 		}
// 	}
// 	fmt.Printf("\nProcess joining request passed \n")
// }