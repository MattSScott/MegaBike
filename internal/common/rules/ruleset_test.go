package rules

import "testing"

func TestRegisterRuleToCache(t *testing.T) {
	ruleInputs := RuleInputs{Colour, Energy}
	ruleMat := [][]float64{{1, 0, 1}, {0, 1, -100}}
	ruleComps := RuleComparators{EQ, LEQ}

	rule := GenerateRule(MoveBike, "test_rule", ruleInputs, ruleMat, ruleComps, true)

	grc := GenerateGlobalRuleCache()

	if len(grc.rawRuleSet) > 0 {
		t.Errorf("Failed to generate empty rule cache")
	}

	grc.AddRuleToCache(rule)

	if len(grc.rawRuleSet) != 1 {
		t.Errorf("Failed to add rule to cache")
	}

}
