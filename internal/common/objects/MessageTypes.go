package objects

import (
	"SOMAS2023/internal/common/utils"

	"github.com/MattSScott/basePlatformSOMAS/messaging"
	"github.com/google/uuid"
)

// ----- Round (plus InvokerHandlers) -----

// "I proposed this lootbox direction in this round"
type ProposedLootboxMessage struct {
	messaging.BaseMessage[IBaseBiker]
	Lootbox uuid.UUID // the lootbox ID you proposed this round
}
// "I applied this force in this round"
type ForcesMessage struct {
	messaging.BaseMessage[IBaseBiker]
	AgentForces utils.Forces // the forces
}

// "I did or did not conform to the group chosen direction this round"
type ConformMessage struct {
	messaging.BaseMessage[IBaseBiker]
	DidConform bool// whether you conformed or not
}

func (msg ProposedLootboxMessage) InvokeMessageHandler(agent IBaseBiker) {
	agent.HandleProposedLootboxMessage(msg)
}

func (msg ForcesMessage) InvokeMessageHandler(agent IBaseBiker) {
	agent.HandleForcesMessage(msg)
}

func (msg ConformMessage) InvokeMessageHandler(agent IBaseBiker) {
	agent.HandleConformMessage(msg)
}

// ----- Iteration (plus InvokeHandlers) -----

// "I want to kick out this agent"
type KickoutAgentMessage struct {
	messaging.BaseMessage[IBaseBiker]
	AgentId uuid.UUID // agent who you want to kick off
}

// "I want to move to this bike"
type ChangeBikeMessage struct {
	messaging.BaseMessage[IBaseBiker]
	BikeId  uuid.UUID // the bike this agent wants to join
}

func (msg KickoutAgentMessage) InvokeMessageHandler(agent IBaseBiker) {
	agent.HandleKickoutMessage(msg)
}

func (msg ChangeBikeMessage) InvokeMessageHandler(agent IBaseBiker) {
	agent.HandleChangeBikeMessage(msg)
}
