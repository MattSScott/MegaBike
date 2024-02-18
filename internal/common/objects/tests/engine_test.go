package objects

import (
	"SOMAS2023/internal/clients/teamSOSA/agent"
	"SOMAS2023/internal/common/objects"

	"SOMAS2023/internal/server"
	"testing"

	"github.com/google/uuid"
)

func TestRuleEvaluatesAsTrue(t *testing.T) {
	// generate default agent with colour == 1 and energy == 100
	testServer := server.GenerateServer()
	testAgent := agent.NewAgentSOSA(objects.GetBaseBiker(1, uuid.New(), testServer))
	// make colour deterministic (red) and energy (1-0.25 = 0.75)
	const COL = 1
	testAgent.SetDeterministicColour(COL)
	testAgent.UpdateEnergyLevel(-0.25)

	// generate test rule - energy > 50 and energy < 100, colour == 1
	testRuleMatrix := [][]float64{{1, 0, -COL}, {0, 1, -1}, {0, 1, -0.5}}
	testInputVector := objects.RuleInputs{objects.Colour, objects.Energy}
	testRuleComps := objects.RuleComparators{objects.EQ, objects.LT, objects.GT}

	rule := objects.GenerateRule(objects.Lootbox, "testRule", testInputVector, testRuleMatrix, testRuleComps, false)

	if rule.EvaluateAgentRule(testAgent) != true {
		t.Error("Rule incorrectly evaluated as false")
	}
}

func TestRuleEvaluatesAsFalse(t *testing.T) {
	// generate default agent with colour == 1 and energy == 100
	testServer := server.GenerateServer()
	testAgent := agent.NewAgentSOSA(objects.GetBaseBiker(1, uuid.New(), testServer))
	// make colour deterministic (red) and energy (1-0.25 = 0.75)
	const COL = 1
	testAgent.SetDeterministicColour(COL)
	testAgent.UpdateEnergyLevel(-0.75)

	// generate test rule - energy > 50 and energy < 100, colour == 1
	testRuleMatrix := [][]float64{{1, 0, -COL}, {0, 1, -1}, {0, 1, -0.5}}
	testInputVector := objects.RuleInputs{objects.Colour, objects.Energy}
	testRuleComps := objects.RuleComparators{objects.EQ, objects.LT, objects.GT}

	rule := objects.GenerateRule(objects.Lootbox, "testRule", testInputVector, testRuleMatrix, testRuleComps, false)

	if rule.EvaluateAgentRule(testAgent) != false {
		t.Error("Rule incorrectly evaluated as true")
	}
}

func TestRuleImmutability(t *testing.T) {

	// generate test rule - energy > 50 and energy < 100, colour == 1
	testRuleMatrix := [][]float64{{1, 0, -4}, {0, 1, -1}, {0, 1, -0.5}}
	testInputVector := objects.RuleInputs{objects.Colour, objects.Energy}
	testRuleComps := objects.RuleComparators{objects.EQ, objects.LT, objects.GT}

	rule := objects.GenerateRule(objects.Lootbox, "testRule", testInputVector, testRuleMatrix, testRuleComps, false)

	updatedRuleMatrix := testRuleMatrix
	updatedRuleMatrix[0][2] = -3
	err := rule.UpdateRuleMatrix(updatedRuleMatrix)

	if err == nil {
		t.Error("Rule immutability not obeyed (should be immutable)")
	}

	rule.ToggleRuleMutability()

	newErr := rule.UpdateRuleMatrix(updatedRuleMatrix)

	if newErr != nil {
		t.Error("Rule immutability not obeyed (should not be immutable)")
	}
}

func TestRulePassesAfterMutation(t *testing.T) {
	// generate default agent with colour == 1 and energy == 100
	testServer := server.GenerateServer()
	testAgent := agent.NewAgentSOSA(objects.GetBaseBiker(1, uuid.New(), testServer))
	// make points deterministic
	testAgent.UpdatePoints(20)
	testAgent.UpdateEnergyLevel(-0.25)

	// generate test rule - energy > 50 and energy < 100, colour == 1
	testRuleMatrix := [][]float64{{1, -20, -4}}
	testInputVector := objects.RuleInputs{objects.Points, objects.Energy}
	testRuleComps := objects.RuleComparators{objects.EQ}

	rule := objects.GenerateRule(objects.Lootbox, "testRule", testInputVector, testRuleMatrix, testRuleComps, true)

	if rule.EvaluateAgentRule(testAgent) != false {
		t.Error("Rule incorrectly evaluated as true")
	}

	fixedRuleMat := testRuleMatrix
	fixedRuleMat[0][2] = -5

	if rule.EvaluateAgentRule(testAgent) != true {
		t.Error("Rule incorrectly evaluated as false")
	}

}

func TestDefaultRuleAlwaysPasses(t *testing.T) {
	testServer := server.GenerateServer()
	testServer.Initialize(5)
	testAgent := agent.NewAgentSOSA(objects.GetBaseBiker(1, uuid.New(), testServer))
	testServer.AddAgent(testAgent)
	testServer.FoundingInstitutions()
	rule := objects.GenerateNullPassingRule()

	if !rule.EvaluateAgentRule(testAgent) {
		t.Error("Default rule evaluated as false")
	}
}
