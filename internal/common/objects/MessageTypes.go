package objects

import (
	"SOMAS2023/internal/common/utils"

	voting "SOMAS2023/internal/common/voting"

	"github.com/MattSScott/basePlatformSOMAS/messaging"
	"github.com/google/uuid"
)

// ----- Round (plus InvokerHandlers) -----

// "I proposed this lootbox direction in this round"
type ProposedLootboxMessage struct {
	messaging.BaseMessage[IBaseBiker]
	Lootbox uuid.UUID // the lootbox ID you proposed this round
}

// "I want to go to this lootbox next round"
type NextLootboxMessage struct {
	messaging.BaseMessage[IBaseBiker]
	Lootbox uuid.UUID // the lootbox that agent wants next round
}

// "I applied this force in this round"
type ForcesMessage struct {
	messaging.BaseMessage[IBaseBiker]
	AgentId     uuid.UUID    // the agent whose forces are shared
	AgentForces utils.Forces // the forces
}

// "I voted for this resource allocation in this round"
type VoteAllocationMessage struct {
	messaging.BaseMessage[IBaseBiker]
	VoteMap voting.IdVoteMap // the vote map that you voted for (if you are telling the truth)
}

func (msg ProposedLootboxMessage) InvokeMessageHandler(agent IBaseBiker) {
	agent.HandleProposedLootboxMessage(msg)
}

func (msg NextLootboxMessage) InvokeMessageHandler(agent IBaseBiker) {
	agent.HandleNextLootboxMessage(msg)
}

func (msg ForcesMessage) InvokeMessageHandler(agent IBaseBiker) {
	agent.HandleForcesMessage(msg)
}

func (msg VoteAllocationMessage) InvokeMessageHandler(agent IBaseBiker) {
	agent.HandleVoteAllocationMessage(msg)
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
