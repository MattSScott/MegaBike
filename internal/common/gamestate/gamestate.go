package gamestate

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/rules"

	"github.com/google/uuid"
)

type RuleCacheOperations interface {
	ViewGlobalRuleCache() rules.GlobalRuleCache
	AddToGlobalRuleCache(*rules.Rule)
}

/*
IGameState is an interface for GameState that objects will use to get the current game state
*/
type IGameState interface {
	RuleCacheOperations
	GetLootBoxes() map[uuid.UUID]objects.ILootBox
	GetMegaBikes() map[uuid.UUID]objects.IMegaBike
	GetAgentMap() map[uuid.UUID]objects.IBaseBiker
	GetAwdi() objects.IAwdi
}
