package server_test

import (
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/server"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestInitialize(t *testing.T) {

	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	if len(s.GetAgentMap()) != *globals.BikerAgentCount {
		t.Error("agents not properly instantiated")
	}

	if len(s.GetMegaBikes()) != globals.MegaBikeCount {
		t.Error("megabikes not properly instantiated")
	}

	if len(s.GetLootBoxes()) != globals.LootBoxCount {
		t.Error("loot boxes not properly instantiated")
	}

	if len(s.ViewGlobalRuleCache()) != *globals.GlobalRuleCount {
		t.Error("ruleset not properly instantiated")
	}

	if s.GetAwdi().GetID() == uuid.Nil {
		t.Error("awdi not properly instantiated")
	}

	fmt.Printf("\nInitialize passed \n")
}

func TestInitialAssignment(t *testing.T) {
	s := server.GenerateServer()
	s.Initialize(0)

	for _, bike := range s.GetMegaBikes() {
		if len(bike.GetAgents()) > 8 {
			t.Error("Bike has too many agents")
		}
	}

	for _, agent := range s.GetAgentMap() {
		if !agent.GetBikeStatus() {
			t.Error("Agent has bike status as false")
		}

		if agent.GetBike() == uuid.Nil {
			t.Error("Agent has null bike ID")
		}
	}
}

// func TestRunGame(t *testing.T) {
// 	iterations := 2
// 	s := server.GenerateServer()
// 	s.Initialize(iterations)
// 	s.Start()
// }
