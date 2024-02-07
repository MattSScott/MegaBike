package objects

import (
	"SOMAS2023/internal/common/objects"
	"testing"
)

func TestRegisterRuleToCache(t *testing.T) {
	ruleInputs := objects.RuleInputs{objects.Colour, objects.Energy}
	ruleMat := [][]float64{{1, 0, 1}, {0, 1, -100}}
	ruleComps := objects.RuleComparators{objects.EQ, objects.LEQ}

	rule := objects.GenerateRule(objects.MoveBike, "test_rule", ruleInputs, ruleMat, ruleComps, true)

	grc := objects.GenerateGlobalRuleCache()

	if len(grc.ViewGlobalRuleSet()) > 0 {
		t.Errorf("Failed to generate empty rule cache")
	}

	grc.AddRuleToCache(rule)

	if len(grc.ViewGlobalRuleSet()) != 1 {
		t.Errorf("Failed to add rule to cache")
	}

}
