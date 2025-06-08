package ServerTests

import (
	"MegabikeFYPVersion/internal/server"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// Test 1: Does the server set everything up okay before the game starts?
func TestInitialize(t *testing.T) {

	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	if len(s.GetAgentMap()) != s.GetBikerAgentCount() {
		t.Error("Agents not properly instantiated")
	}

	if len(s.GetMegaBikes()) != s.GetMegabikeCount() {
		t.Error("Megabikes not properly instantiated")
	}

	if len(s.GetLootBoxes()) != s.GetLootboxCount() {
		t.Error("Lootboxes not properly instantiated")
	}

	if s.GetAwdi().GetID() == uuid.Nil {
		t.Error("Awdi not properly instantiated")
	}

	for _, agent := range s.GetAgentMap() {
		if agent.GetBikeStatus() {
			t.Error("Agent shouldn't be on a bike at the start of the game")
		}

		if agent.GetBike() != uuid.Nil {
			t.Error("Agent shouldn't have a bike id at the start of the game")
		}
	}

	fmt.Printf("\nInitialize passed \n")
}
