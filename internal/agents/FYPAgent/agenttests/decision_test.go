package AgentTests

import (
	"MegabikeFYPVersion/internal/server"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestDecideDirection(t *testing.T) {

	// setup
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	iterationDump := s.GenerateIterationDump()
	s.RunVoluntaryReassociation(iterationDump)

	for _, agent := range s.GetAgentMap() {
		if agent.GetBikeStatus() {
			chosenLootbox := agent.DecideDirection()
			if chosenLootbox == uuid.Nil {
				t.Error("Agent suggesting a nil lootbox")
			}
		}
	}

	fmt.Println("All agents successfully suggesting a lootbox")

}
