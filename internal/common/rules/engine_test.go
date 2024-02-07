package rules

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
	testInputVector := RuleInputs{Colour, Energy}
	testRuleComps := RuleComparators{EQ, LT, GT}

	rule := GenerateRule(Lootbox, "testRule", testInputVector, testRuleMatrix, testRuleComps, false)

	if rule.EvaluateRule(testAgent) != true {
		t.Error("Rule incorrectly evaluated as false")
	}
}
