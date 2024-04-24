package objects

import (
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/server"
	"testing"

	"github.com/google/uuid"
)

type MockBiker struct {
	*objects.BaseBiker
	ID      uuid.UUID
	VoteMap map[uuid.UUID]int
	// kickedOutCount int
	// governance     utils.Governance
	// ruler          uuid.UUID
}

func (mb *MockBiker) GetLocation() utils.Coordinates {
	return utils.GenerateRandomCoordinates()
}

type MegaBike struct {
	// agents         []Biker
	// kickedOutCount int
}

type MockRuleCache struct{}

func (mrc *MockRuleCache) ViewGlobalRuleCache() map[uuid.UUID]*objects.Rule {
	return make(map[uuid.UUID]*objects.Rule)
}

func (mrc *MockRuleCache) AddToGlobalRuleCache(*objects.Rule) {}

func NewMockBiker(gameState objects.IGameState) *MockBiker {
	baseBiker := objects.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), gameState)

	return &MockBiker{
		BaseBiker: baseBiker,
		ID:        uuid.New(),
		VoteMap:   make(map[uuid.UUID]int),
	}
}

func (mb *MockBiker) VoteForKickout() map[uuid.UUID]int {
	return mb.VoteMap
}

type Biker interface {
	VoteForKickout() map[uuid.UUID]int
}

// Ensure that BaseBiker implements the Biker interface.
//var _ Biker = &objects.BaseBiker{}

func (mb *MockBiker) GetID() uuid.UUID {
	return mb.ID
}

func TestGetMegaBike(t *testing.T) {
	mb := objects.GetMegaBike(&MockRuleCache{})

	if mb == nil {
		t.Errorf("GetMegaBike returned nil")
	}

	if mb.GetGovernance() != utils.Democracy {
		t.Errorf("Expected governance to be Democracy, got %v", mb.GetGovernance())
	}

	if mb.GetRuler() != uuid.Nil {
		t.Errorf("Expected ruler to be uuid.Nil, got %v", mb.GetRuler())
	}
}

func TestAddAgent(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	mb := objects.GetMegaBike(&MockRuleCache{})
	biker := NewMockBiker(s)

	mb.AddAgent(biker)

	if len(mb.GetAgents()) != 1 {
		t.Errorf("AddAgent failed to add the agent to MegaBike")
	}

	if mb.GetAgents()[0].GetID() != biker.GetID() {
		t.Errorf("The added agent ID does not match the expected MockBiker ID")
	}
}

func TestRemoveAgent(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	mb := objects.GetMegaBike(&MockRuleCache{})
	biker1 := NewMockBiker(s)
	biker2 := NewMockBiker(s)

	mb.AddAgent(biker1)
	mb.AddAgent(biker2)

	mb.RemoveAgent(biker1.GetID())

	agents := mb.GetAgents()

	if len(agents) != 1 {
		t.Errorf("RemoveAgent failed to remove the agent from MegaBike, expected 1 agent, got %d", len(agents))
	}

	if agents[0].GetID() == biker1.GetID() {
		t.Errorf("RemoveAgent did not remove the correct agent")
	}

	if agents[0].GetID() != biker2.GetID() {
		t.Errorf("The remaining agent ID does not match the expected MockBiker ID")
	}
}

func TestUpdateMass(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	mb := objects.GetMegaBike(&MockRuleCache{})
	initialMass := mb.GetPhysicalState().Mass

	mb.AddAgent(NewMockBiker(s))
	mb.AddAgent(NewMockBiker(s))
	mb.UpdateMass()

	updatedMass := mb.GetPhysicalState().Mass

	expectedMass := initialMass + 2

	if updatedMass != expectedMass {
		t.Errorf("UpdateMass did not calculate the correct mass: got %v, want %v", updatedMass, expectedMass)
	}
}

func TestUpdateOrientation(t *testing.T) {
	// Scenario 0: No steering bikers
	t.Run("Single Biker Test", func(t *testing.T) {
		iterations := 3
		s := server.GenerateServer()
		s.Initialize(iterations)

		mb := objects.GetMegaBike(&MockRuleCache{})
		biker1 := NewMockBiker(s)
		biker2 := NewMockBiker(s)

		turningDecision := utils.TurningDecision{
			SteerBike:     false,
			SteeringForce: 0.3,
		}

		force := utils.Forces{
			Pedal:   utils.BikerMaxForce,
			Brake:   0.0,
			Turning: turningDecision,
		}

		biker1.SetForces(force)
		biker2.SetForces(force)
		mb.AddAgent(biker1)
		mb.AddAgent(biker2)

		mb.UpdateOrientation()

		// Check if orientation updated correctly
		// Assuming initial orientation is 0.0 and your logic for orientation update
		expectedOrientation := 0.0 // Adjust this value based on your orientation update logic
		if mb.GetOrientation() != expectedOrientation {
			t.Errorf("got %v, want %v", mb.GetOrientation(), expectedOrientation)
		}
	})
	// Scenario 1: Single Biker Test
	t.Run("Single Biker Test", func(t *testing.T) {
		iterations := 3
		s := server.GenerateServer()
		s.Initialize(iterations)

		mb := objects.GetMegaBike(&MockRuleCache{})
		biker := NewMockBiker(s)

		turningDecision := utils.TurningDecision{
			SteerBike:     true,
			SteeringForce: 0.3,
		}

		force := utils.Forces{
			Pedal:   utils.BikerMaxForce,
			Brake:   0.0,
			Turning: turningDecision,
		}

		biker.SetForces(force)
		mb.AddAgent(biker)

		mb.UpdateOrientation()

		// Check if orientation updated correctly
		// Assuming initial orientation is 0.0 and your logic for orientation update
		expectedOrientation := 0.3 // Adjust this value based on your orientation update logic
		if mb.GetOrientation() != expectedOrientation {
			t.Errorf("got %v, want %v", mb.GetOrientation(), expectedOrientation)
		}
	})

	// Scenario 2: Biker doesn't want to steer
	t.Run("Multiple Bikers Test", func(t *testing.T) {
		iterations := 3
		s := server.GenerateServer()
		s.Initialize(iterations)

		mb := objects.GetMegaBike(&MockRuleCache{})
		biker1 := NewMockBiker(s)
		biker2 := NewMockBiker(s)

		turningDecision1 := utils.TurningDecision{
			SteerBike:     true,
			SteeringForce: 0.6,
		}

		force1 := utils.Forces{
			Pedal:   utils.BikerMaxForce,
			Brake:   0.0,
			Turning: turningDecision1,
		}

		turningDecision2 := utils.TurningDecision{
			SteerBike:     false,
			SteeringForce: 0.3,
		}

		force2 := utils.Forces{
			Pedal:   utils.BikerMaxForce,
			Brake:   0.0,
			Turning: turningDecision2,
		}

		biker1.SetForces(force1)
		biker2.SetForces(force2)
		mb.AddAgent(biker1)
		mb.AddAgent(biker2)

		mb.UpdateOrientation()

		// Check if orientation updated correctly
		// Assuming each biker contributes equally and your logic for orientation update
		expectedOrientation := 0.6
		tolerance := 0.001 // Define a small tolerance for floating-point comparison

		actualOrientation := mb.GetOrientation()
		if actualOrientation < expectedOrientation-tolerance || actualOrientation > expectedOrientation+tolerance {
			t.Errorf("got %v, want %v (within a tolerance of %v)", actualOrientation, expectedOrientation, tolerance)
		}
	})

	// Scenario 3: Three Bikers with Different Directions (expected 0.1)
	t.Run("Three Bikers Different Directions", func(t *testing.T) {
		iterations := 3
		s := server.GenerateServer()
		s.Initialize(iterations)

		mb := objects.GetMegaBike(&MockRuleCache{})
		biker1 := NewMockBiker(s)
		biker2 := NewMockBiker(s)
		biker3 := NewMockBiker(s)

		// Set unique forces for each biker
		forces := []utils.Forces{
			{Pedal: utils.BikerMaxForce, Brake: 0.0, Turning: utils.TurningDecision{SteerBike: true, SteeringForce: 0.1}},
			{Pedal: utils.BikerMaxForce, Brake: 0.0, Turning: utils.TurningDecision{SteerBike: true, SteeringForce: -0.7}},
			{Pedal: utils.BikerMaxForce, Brake: 0.0, Turning: utils.TurningDecision{SteerBike: true, SteeringForce: 0.3}},
		}

		bikers := []*MockBiker{biker1, biker2, biker3}
		for i, biker := range bikers {
			biker.SetForces(forces[i])
			mb.AddAgent(biker)
		}

		mb.UpdateOrientation()

		// Hardcoded expected orientation
		expectedOrientation := 0.1
		tolerance := 0.001 // Define a small tolerance for floating-point comparison

		actualOrientation := mb.GetOrientation()
		if actualOrientation < expectedOrientation-tolerance || actualOrientation > expectedOrientation+tolerance {
			t.Errorf("got %v, want %v (within a tolerance of %v)", actualOrientation, expectedOrientation, tolerance)
		}
	})

	// Scenario 4: Two Bikers, one with -1 and one with 1, expected orientation 1 or -1
	t.Run("Two Bikers Opposite Forces", func(t *testing.T) {
		iterations := 3
		s := server.GenerateServer()
		s.Initialize(iterations)

		mb := objects.GetMegaBike(&MockRuleCache{})
		biker1 := NewMockBiker(s)
		biker2 := NewMockBiker(s)

		// Set forces for each biker
		force1 := utils.Forces{
			Pedal: utils.BikerMaxForce, Brake: 0.0, Turning: utils.TurningDecision{SteerBike: true, SteeringForce: -1},
		}
		force2 := utils.Forces{
			Pedal: utils.BikerMaxForce, Brake: 0.0, Turning: utils.TurningDecision{SteerBike: true, SteeringForce: 1},
		}

		biker1.SetForces(force1)
		biker2.SetForces(force2)
		mb.AddAgent(biker1)
		mb.AddAgent(biker2)

		mb.UpdateOrientation()

		// Hardcoded expected orientation
		expectedOrientation1 := 1.0
		expectedOrientation2 := -1.0

		actualOrientation := mb.GetOrientation()
		if actualOrientation != expectedOrientation1 && actualOrientation != expectedOrientation2 {
			t.Errorf("got %v, want %v or %v", actualOrientation, expectedOrientation1, expectedOrientation2)
		}
	})

	// Scenario 5: Two Bikers, one with -0.6 (-108°) and one with 0.7 (126°), expected orientation 0.95 (−171°)
	t.Run("Two Bikers Opposite Forces", func(t *testing.T) {
		iterations := 3
		s := server.GenerateServer()
		s.Initialize(iterations)

		mb := objects.GetMegaBike(&MockRuleCache{})
		biker1 := NewMockBiker(s)
		biker2 := NewMockBiker(s)

		// Set forces for each biker
		force1 := utils.Forces{
			Pedal: utils.BikerMaxForce, Brake: 0.0, Turning: utils.TurningDecision{SteerBike: true, SteeringForce: -0.6},
		}
		force2 := utils.Forces{
			Pedal: utils.BikerMaxForce, Brake: 0.0, Turning: utils.TurningDecision{SteerBike: true, SteeringForce: 0.7},
		}

		biker1.SetForces(force1)
		biker2.SetForces(force2)
		mb.AddAgent(biker1)
		mb.AddAgent(biker2)

		mb.UpdateOrientation()

		// Hardcoded expected orientation
		expectedOrientation := -0.95
		tolerance := 0.001 // Define a small tolerance for floating-point comparison

		actualOrientation := mb.GetOrientation()
		if actualOrientation < expectedOrientation-tolerance || actualOrientation > expectedOrientation+tolerance {
			t.Errorf("got %v, want %v (within a tolerance of %v)", actualOrientation, expectedOrientation, tolerance)
		}
	})

	// Scenario 6: Two Bikers, one with -0.1 (-18°) and one with 0.2 (36°), expected orientation 0.05 (9°)
	t.Run("Two Bikers Opposite Forces", func(t *testing.T) {
		iterations := 3
		s := server.GenerateServer()
		s.Initialize(iterations)

		mb := objects.GetMegaBike(&MockRuleCache{})
		biker1 := NewMockBiker(s)
		biker2 := NewMockBiker(s)

		// Set forces for each biker
		force1 := utils.Forces{
			Pedal: utils.BikerMaxForce, Brake: 0.0, Turning: utils.TurningDecision{SteerBike: true, SteeringForce: -0.1},
		}
		force2 := utils.Forces{
			Pedal: utils.BikerMaxForce, Brake: 0.0, Turning: utils.TurningDecision{SteerBike: true, SteeringForce: 0.2},
		}

		biker1.SetForces(force1)
		biker2.SetForces(force2)
		mb.AddAgent(biker1)
		mb.AddAgent(biker2)

		mb.UpdateOrientation()

		// Hardcoded expected orientation
		expectedOrientation := 0.05
		tolerance := 0.001 // Define a small tolerance for floating-point comparison

		actualOrientation := mb.GetOrientation()
		if actualOrientation < expectedOrientation-tolerance || actualOrientation > expectedOrientation+tolerance {
			t.Errorf("got %v, want %v (within a tolerance of %v)", actualOrientation, expectedOrientation, tolerance)
		}
	})
}

func TestGetSetGovernanceAndRuler(t *testing.T) {
	mb := objects.GetMegaBike(&MockRuleCache{})
	originalGovernance := mb.GetGovernance()
	originalRuler := mb.GetRuler()

	newGovernance := utils.Dictatorship
	newRuler := uuid.New()

	mb.SetGovernance(newGovernance)
	mb.SetRuler(newRuler)

	if mb.GetGovernance() != newGovernance {
		t.Errorf("SetGovernance failed, expected %v, got %v", newGovernance, mb.GetGovernance())
	}

	if mb.GetRuler() != newRuler {
		t.Errorf("SetRuler failed, expected %v, got %v", newRuler, mb.GetRuler())
	}

	mb.SetGovernance(originalGovernance)
	mb.SetRuler(originalRuler)
}

func TestKickOutAgent(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)
	s.FoundingInstitutions()

	mb := objects.GetMegaBike(&MockRuleCache{})

	//biker1 := NewMockBiker(uuid.New(), map[uuid.UUID]int{ /* votes */ })
	biker1 := NewMockBiker(s)
	biker2 := NewMockBiker(s)
	biker3 := NewMockBiker(s)
	mb.AddAgent(biker1)
	mb.AddAgent(biker2)
	mb.AddAgent(biker3)

	weights := map[uuid.UUID]float64{
		biker1.GetID(): 1.0,
		biker2.GetID(): 1.0,
		biker3.GetID(): 1.0,
	}

	// Voting
	biker1.VoteMap[biker3.GetID()] = 1
	biker2.VoteMap[biker3.GetID()] = 1
	biker3.VoteMap[biker1.GetID()] = 1

	for _, biker := range []Biker{biker1, biker2, biker3} {
		biker.VoteForKickout()
	}

	// Kick out agents based on votes and weights.
	kickedOutAgents := mb.KickOutAgent(weights)

	if len(kickedOutAgents) != 1 {
		t.Fatalf("KickOutAgent kicked out %d agents; want 1", len(kickedOutAgents))
	}

	if kickedOutAgents[0] != biker3.GetID() {
		t.Errorf("KickOutAgent kicked out incorrect agent: got %v, want %v", kickedOutAgents[0], biker3.GetID())
	}

	for _, anyBike := range s.GetMegaBikes() {
		agentsOnBike := anyBike.GetAgents()
		// Skip empty bikes.
		if agentsOnBike == nil {
			continue
		}
		// Check if biker3 is still on a bike.
		for _, agentOnBike := range agentsOnBike {
			if agentOnBike.GetID() == biker3.GetID() {
				t.Errorf("Kicked out agent is still present on a MegaBike")
			}
		}
	}
}

func TestPopulateBikeWithFullRuleset(t *testing.T) {
	serv := server.GenerateServer()
	serv.Initialize(1)

	mb := objects.GetMegaBike(serv)

	if len(mb.ViewLocalRuleMap()) != 0 {
		t.Error("Rulemap not initialised as empty")
	}

	mb.ActivateAllGlobalRules()

	categories := int(*globals.GlobalRuleCount) / int(objects.MAX_ACTIONS)

	if len(mb.ViewLocalRuleMap()) != categories {
		t.Error("Rules not generated in all categories")
	}
}

func TestRulesExtractedAndPassEvent(t *testing.T) {

	serv := server.GenerateServer()
	serv.Initialize(1)

	mb := objects.GetMegaBike(serv)
	for i := 0; i < 8; i++ {
		mb.AddAgent(NewMockBiker(serv))
	}

	if len(mb.GetAgents()) != 8 {
		t.Error("Agents not correctly added to bike")
	}

	if len(mb.ViewLocalRuleMap()) != 0 {
		t.Error("Rulemap not initialised as empty")
	}

	mb.ActivateAllGlobalRules()

	if !mb.ActionIsValidForRuleset(objects.AppliesAll) {
		t.Error("Rules not properly passing (applies all)")
	}

	if !mb.ActionIsValidForRuleset(objects.Lootbox) {
		t.Error("Rules not properly passing (lootbox)")
	}

}
