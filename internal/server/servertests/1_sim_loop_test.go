package ServerTests

import (
	// "MegabikeFYPVersion/internal/common/objects"
	"MegabikeFYPVersion/internal/common/objects"
	"MegabikeFYPVersion/internal/common/utils"
	"MegabikeFYPVersion/internal/server"
	"fmt"
	"testing"

	"github.com/google/uuid"
	// "github.com/stretchr/testify/assert"
)

func TestRemoveAllAgentsFromBikes(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	// get a random bike id for testing
	var testBike objects.IMegaBike

	i := 0
	for _, bike := range s.GetMegaBikes() {
		if i == 1 {
			break
		}
		testBike = bike
		i += 1
	}

	// add a couple of random agents to this bike
	i = 0
	for _, agent := range s.GetAgentMap() {
		if i < 2 {
			s.AddAgentToBike(agent, testBike)
		} else {
			break
		}
		i += 1
	}

	// Call the function and test that it works as intended
	s.RemoveAllAgentsFromBikes()

	for _, agent := range s.GetAgentMap() {
		if agent.GetBikeStatus() {
			t.Error("No agent should be on a bike")
		}

		if agent.GetBike() != uuid.Nil {
			t.Error("No agent should have a bike id")
		}
	}

	for _, bike := range s.GetMegaBikes() {
		if len(bike.GetAgents()) > 0 {
			t.Error("agents are left on bikes still")
		}
	}

	fmt.Printf("\nRemove all agents from bikes passed \n")
}

func TestRandomlyAssignRepresentativesFullAgentMap(t *testing.T) {

	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	// remove agents first just in case
	s.RemoveAllAgentsFromBikes()

	// now run the function
	s.RandomlyAssignRepresentatives()

	for _, bike := range s.GetMegaBikes() {
		switch bike.GetGovernance() {
		case utils.Many:
			if len(bike.GetRepresentatives()) > 0 {
				t.Error("Many bike has incorrect number of reps (should have 0)")
			}
		case utils.Some:
			if len(bike.GetRepresentatives()) != 3 {
				t.Error("Some bike has incorrect number of representatives (should have 3)")
			}
		case utils.One:
			if len(bike.GetRepresentatives()) != 1 {
				t.Error("One bike has incorrect number of representatives (should have 1)")
			}

		}
	}

	fmt.Printf("\n Random assignment of representatives passed \n")
}

func TestRandomlyAssignRepresentativesReducedAgentMap(t *testing.T) {

	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	// remove agents first just in case
	s.RemoveAllAgentsFromBikes()

	// kill 22 agents, leaving 2
	i := 0
	for _, agent := range s.GetAgentMap() {
		if i < 22 {
			s.RemoveAgent(agent)
		} else {
			continue
		}
		i++
	}

	s.RandomlyAssignRepresentatives()

	for _, bike := range s.GetMegaBikes() {
		switch bike.GetGovernance() {
		case utils.Many:
			if len(bike.GetRepresentatives()) > 0 {
				t.Error("Many bike has incorrect number of reps (should have 0)")
			}
		case utils.Some:
			if len(bike.GetRepresentatives()) != 1 {
				t.Error("Some bike has incorrect number of representatives (should have 1)")
			}
		case utils.One:
			if len(bike.GetRepresentatives()) != 1 {
				t.Error("One bike has incorrect number of representatives (should have 1)")
			}

		}
	}

}

// to write
func TestProcessAgentQueue(t *testing.T) {

	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	// setup
	s.RemoveAllAgentsFromBikes()
	queuedAgents := s.RandomlyAssignRepresentatives()
	iterationDump := s.GenerateIterationDump()

	s.ProcessAgentQueue(queuedAgents, iterationDump)

}

func TestGiniCoefficientCalculations(t *testing.T) {

	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	testIncomes := []float64{0.64, 0.32, 0.23, 0.89, 1.43, 0.43, 2.11, 0.82}
	gini := s.CalculateGiniCoefficient(testIncomes)
	fmt.Println(gini)
	if !(0.36880 < gini && gini < 0.36882) {
		t.Error("not calculating gini coefficient correctly")
	}

	fmt.Println("Gini coefficient calculation working correctly")

}

// func TestGetJoiningRequests(t *testing.T) {
// 	iterations := 3
// 	s := server.GenerateServer()
// 	s.Initialize(iterations)

// 	// 1: get two bike ids
// 	targetBikes := make([]uuid.UUID, 2)

// 	i := 0
// 	for bikeId := range s.GetMegaBikes() {
// 		if i == 2 {
// 			break
// 		}
// 		targetBikes[i] = bikeId
// 		i += 1
// 	}

// 	// 2: set one agent requesting the first bike and two other requesting the second one
// 	i = 0
// 	requests := make(map[uuid.UUID][]uuid.UUID)
// 	requests[targetBikes[0]] = make([]uuid.UUID, 1)
// 	requests[targetBikes[1]] = make([]uuid.UUID, 2)
// 	for _, agent := range s.GetAgentMap() {
// 		if i == 0 {
// 			agent.ToggleOnBike()
// 			agent.SetBike(targetBikes[0])
// 			requests[targetBikes[0]][0] = agent.GetID()
// 		} else if i <= 2 {
// 			agent.ToggleOnBike()
// 			agent.SetBike(targetBikes[1])
// 			requests[targetBikes[1]][i-1] = agent.GetID()
// 		} else {
// 			break
// 		}
// 		i += 1
// 	}

// 	// 3. check that joining requests reflect the previous actions
// 	bikeRequests := s.GetJoiningRequests(make([]uuid.UUID, 0))

// 	if len(bikeRequests) != len(requests) {
// 		t.Error("bike requests processed incorrectly: empty")
// 	}

// 	for bikeId, agentIds := range bikeRequests {
// 		if len(agentIds) != len(requests[bikeId]) {
// 			t.Error("bike requests processed incorrectly: wrong number of agents for given bike")
// 		}
// 	}
// 	fmt.Printf("\nJoining request passed \n")
// }

// func TestGetJoiningRequestsWithLimbo(t *testing.T) {
// 	iterations := 3
// 	s := server.GenerateServer()
// 	s.Initialize(iterations)

// 	// 1: get two bike ids
// 	targetBikes := make([]uuid.UUID, 2)

// 	i := 0
// 	for bikeId := range s.GetMegaBikes() {
// 		if i == 2 {
// 			break
// 		}
// 		targetBikes[i] = bikeId
// 		i += 1
// 	}

// 	// 2: set one agent requesting the first bike and two other requesting the second one
// 	i = 0
// 	requests := make(map[uuid.UUID][]uuid.UUID)
// 	requests[targetBikes[0]] = make([]uuid.UUID, 1)
// 	requests[targetBikes[1]] = make([]uuid.UUID, 1)
// 	limbo := make([]uuid.UUID, 1)
// 	for _, agent := range s.GetAgentMap() {
// 		if i == 0 {
// 			agent.ToggleOnBike()
// 			agent.SetBike(targetBikes[0])
// 			requests[targetBikes[0]][0] = agent.GetID()
// 		} else if i == 1 {
// 			// add it to second bike for request
// 			agent.ToggleOnBike()
// 			agent.SetBike(targetBikes[1])
// 			requests[targetBikes[1]][i-1] = agent.GetID()
// 		} else if i == 2 {
// 			//remove it from bike but add it to limbo (to mimic request made in this turn)
// 			agent.ToggleOnBike()
// 			agent.SetBike(targetBikes[1])
// 			limbo[0] = agent.GetID()
// 		} else {
// 			break
// 		}
// 		i += 1
// 	}

// 	// 3. check that joining requests reflect the previous actions
// 	bikeRequests := s.GetJoiningRequests(limbo)
// 	assert.Equal(t, len(bikeRequests), len(requests), "bike requests processed incorrectly: empty")

// 	for bikeId, agentIds := range bikeRequests {
// 		assert.Equal(t, len(agentIds), len(requests[bikeId]), "bike requests processed incorrectly: wrong number of agents for given bike")
// 		assert.False(t, slices.Contains(agentIds, limbo[0]), "bike requests processed incorrectly: agent in limbo is requesting a bike")
// 	}

// 	fmt.Printf("\nJoining request passed \n")
// }
