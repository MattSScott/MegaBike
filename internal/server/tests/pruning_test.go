package server_test

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/physics"
	"SOMAS2023/internal/server"
	"fmt"
	"testing"

	"github.com/google/uuid"
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
		bike.ClearRuleMap()
		bike.AddToRuleMap(getRule(10000, false))
		rLen := len(bike.GetActiveRulesForAction(objects.Lootbox))
		if rLen != 1 {
			t.Error("Rules not properly added: ruleset size is", rLen)
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
		bike.ClearRuleMap()
		bike.AddToRuleMap(getRule(0, false))
		rLen := len(bike.GetActiveRulesForAction(objects.Lootbox))
		if rLen != 1 {
			t.Error("Rules not properly added: ruleset size is", rLen)
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

	for _, bike := range s.GetMegaBikes() {
		bike.ClearRuleMap()
		bike.AddToRuleMap(getRule(0, false))
		rLen := len(bike.GetActiveRulesForAction(objects.Lootbox))
		if rLen != 1 {
			t.Error("Rules not properly added: ruleset size is", rLen)
		}
	}

	for _, bike := range s.GetMegaBikes() {
		startPos := bike.GetPosition()
		weights := make(map[uuid.UUID]float64)
		for _, agent := range bike.GetAgents() {
			weights[agent.GetID()] = 1.0
		}
		direction := s.RunDemocraticAction(bike, weights)
		fmt.Println(direction)
		for _, agent := range bike.GetAgents() {
			// fmt.Println(direction)
			agent.DecideForce(direction)
		}
		s.MovePhysicsObject(bike)
		endPos := bike.GetPosition()
		distTrav := physics.ComputeDistance(startPos, endPos)
		fmt.Println(distTrav)
		if distTrav > 0.1 {
			t.Error("Bike moved despite no voted direction")
		}
	}
}

func TestPositionChangedAfterPruning(t *testing.T) {
	s := &server.Server{}
	s.Initialize(1)
	s.FoundingInstitutions()

	for _, bike := range s.GetMegaBikes() {
		bike.ClearRuleMap()
		bike.AddToRuleMap(getRule(10000, false))
		rLen := len(bike.GetActiveRulesForAction(objects.Lootbox))
		if rLen != 1 {
			t.Error("Rules not properly added: ruleset size is", rLen)
		}
	}

	for _, bike := range s.GetMegaBikes() {
		startPos := bike.GetPosition()
		weights := make(map[uuid.UUID]float64)
		for _, agent := range bike.GetAgents() {
			weights[agent.GetID()] = 1.0
		}
		direction := s.RunDemocraticAction(bike, weights)
		for _, agent := range bike.GetAgents() {
			// fmt.Println(direction)
			agent.DecideForce(direction)
		}
		s.MovePhysicsObject(bike)
		endPos := bike.GetPosition()
		distTrav := physics.ComputeDistance(startPos, endPos)
		fmt.Println(distTrav)
		if len(bike.GetAgents()) > 0 && distTrav <= 0.1 {
			t.Error("Bike not moved despite no pruning")
		}
	}
}
