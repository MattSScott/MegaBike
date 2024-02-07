package objects

import (
	"github.com/google/uuid"
)

// import (
// 	// "SOMAS2023/internal/common/rules"

// 	"github.com/google/uuid"
// )

type RuleCacheOperations interface {
	ViewGlobalRuleCache() map[uuid.UUID]*Rule
	AddToGlobalRuleCache(*Rule)
}

// /*
// IGameState is an interface for GameState that objects will use to get the current game state
// */
type IGameState interface {
	RuleCacheOperations
	GetLootBoxes() map[uuid.UUID]ILootBox
	GetMegaBikes() map[uuid.UUID]IMegaBike
	GetAgentMap() map[uuid.UUID]IBaseBiker
	GetAwdi() IAwdi
}
