package server_test

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/server"
	"testing"
)

func getRule(dist float64, mute bool) *objects.Rule {
	ruleInputs := objects.RuleInputs{}
	ruleMat := objects.RuleMatrix{{1, -dist}}
	ruleComps := objects.RuleComparators{objects.LEQ}

	return objects.GenerateRule(objects.Lootbox, "lootbox_dist", ruleInputs, ruleMat, ruleComps, mute)
}

func TestCanPruneLootboxes(t *testing.T) {
	s := &server.Server{}
	s.Initialize(1)
	s.FoundingInstitutions()

	for _, bike := range s.GetMegaBikes() {
		bike.AddToRuleMap(getRule(10000, false))
	}

	for _, bike := range s.GetMegaBikes() {
		rLen := len(bike.GetActiveRulesForAction(objects.Lootbox))
		if rLen != 1 {
			t.Error("Rules not properly added: ruleset size is ", rLen)
		}
	}

	if len(s.GetLootBoxes()) == 0 {
		t.Error("No lootboxes spawned")
	}

	for _, bike := range s.GetMegaBikes() {
		found := s.PruneLootboxes(bike)
		if len(found) == 0 {
			t.Error("No lootboxes passing rule")
		}
	}
}

func TestCanOverPruneLootboxes(t *testing.T) {
	s := &server.Server{}
	s.Initialize(1)
	s.FoundingInstitutions()

	for _, bike := range s.GetMegaBikes() {
		bike.AddToRuleMap(getRule(0, false))
	}

	for _, bike := range s.GetMegaBikes() {
		rLen := len(bike.GetActiveRulesForAction(objects.Lootbox))
		if rLen != 1 {
			t.Error("Rules not properly added: ruleset size is ", rLen)
		}
	}

	if len(s.GetLootBoxes()) == 0 {
		t.Error("No lootboxes spawned")
	}

	for _, bike := range s.GetMegaBikes() {
		found := s.PruneLootboxes(bike)
		if len(found) != 0 {
			t.Error("Rule not sufficiently strict")
		}
	}
}

func TestPositionUnchangedAfterPruning(t *testing.T) {
	s := &server.Server{}
	s.Initialize(1)
	s.FoundingInstitutions()

	
}
