package server_test

import (
	"SOMAS2023/internal/server"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestInitialize(t *testing.T) {

	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	if len(s.GetAgentMap()) != server.BikerAgentCount {
		t.Error("Agents not properly instantiated")
	}

	if len(s.GetMegaBikes()) != server.MegaBikeCount {
		t.Error("mega bikes not properly instantiated")
	}

	if len(s.GetLootBoxes()) != server.LootBoxCount {
		t.Error("Mega bikes not properly instantiated")
	}

	if s.GetAwdi().GetID() == uuid.Nil {
		t.Error("awdi not properly instantiated")
	}

	fmt.Printf("\nInitialize passed \n")
}

func TestRunGame(t *testing.T) {
	iterations := 2
	s := server.GenerateServer()
	s.Initialize(iterations)
	s.Start()
}
