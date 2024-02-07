package agent

import (
	"SOMAS2023/internal/common/objects"

	"github.com/google/uuid"
)

type MockGameState struct{}

func (mgs *MockGameState) GetLootBoxes() map[uuid.UUID]objects.ILootBox {
	return make(map[uuid.UUID]objects.ILootBox)
}

func (mgs *MockGameState) GetMegaBikes() map[uuid.UUID]objects.IMegaBike {
	return make(map[uuid.UUID]objects.IMegaBike)
}

func (mgs *MockGameState) GetAgentMap() map[uuid.UUID]objects.IBaseBiker {
	return make(map[uuid.UUID]objects.IBaseBiker)
}

func (mgs *MockGameState) GetAwdi() objects.IAwdi {
	return objects.GetIAwdi()
}

func (mgs *MockGameState) ViewGlobalRuleCache() map[uuid.UUID]*objects.Rule {
	return make(map[uuid.UUID]*objects.Rule)
}

func (mgs *MockGameState) AddToGlobalRuleCache(*objects.Rule) {}
