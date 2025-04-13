package agent

import (
	obj "SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/voting"
	"slices"

	"github.com/MattSScott/basePlatformSOMAS/messaging"
	"github.com/google/uuid"
)

const (
	positiveMessage = 0.05
	negativeMessage = -0.05
	
)

// ----- Round  -----

func (a *AgentSOSA) HandleProposedLootboxMessage(msg obj.ProposedLootboxMessage) {
	
}

func (a *AgentSOSA) HandleNextLootboxMessage(msg obj.NextLootboxMessage) {
	
}

func (a *AgentSOSA) HandleForcesMessage(msg obj.ForcesMessage) {
	
}

func (a *AgentSOSA) HandleVoteAllocationMessage(msg obj.VoteAllocationMessage) {
	
}

func (a *AgentSOSA) GetAllRoundMessages([]obj.IBaseBiker) []messaging.IMessage[obj.IBaseBiker] {
	// TODO: add logic to decide when to send these messages

	proposedLootboxMessage := a.CreateProposedLootboxMessage()
	nextLootboxMsg := a.CreateNextLootboxMessage()
	forcesMsg := a.CreateForcesMessage()
	voteAllocationMessage := a.CreateVoteAllocationMessage()
	
	return []messaging.IMessage[obj.IBaseBiker]{nextLootboxMsg, forcesMsg, proposedLootboxMessage, voteAllocationMessage}
}

func (a *AgentSOSA) CreateProposedLootboxMessage() obj.ProposedLootboxMessage {
	// Currently this returns a default/meaningless message
	// For team's agent, add your own logic to communicate with other agents
	return obj.ProposedLootboxMessage{
		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikers()),
		Lootbox:     uuid.Nil,
	}
}

func (a *AgentSOSA) CreateNextLootboxMessage() obj.NextLootboxMessage {
	// Currently this returns a default message which sends to all bikers on the biker agent's bike
	// For team's agent, add your own logic to communicate with other agents
	return obj.NextLootboxMessage{
		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikers()),
		Lootbox:   uuid.Nil,
	}
}

func (a *AgentSOSA) CreateForcesMessage() obj.ForcesMessage {
	// Currently this returns a default message which sends to all bikers on the biker agent's bike
	// For team's agent, add your own logic to communicate with other agents
	return obj.ForcesMessage{
		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikers()),
		AgentId:     uuid.Nil,
		AgentForces: a.GetForces(),
	}
}

func (a *AgentSOSA) CreateVoteAllocationMessage() obj.VoteAllocationMessage {
	// Currently this returns a default/meaningless message
	// For team's agent, add your own logic to communicate with other agents
	return obj.VoteAllocationMessage{
		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikers()),
		VoteMap:     make(voting.IdVoteMap),
	}
}


// Iteration

// Done
func (a *AgentSOSA) HandleKickOutMessage(msg obj.KickoutAgentMessage) {

	senderID := msg.GetSender().GetID()
	agentId := msg.AgentId

	// if they want to kick us out, lower their trust score. 
	// if they want to kick out the agent who we trust least, i.e. same opinion, increase their trust score. 
	
	if agentId == a.GetID() {
		a.Modules.AgentParameters.UpdateTrustValue(senderID, negativeMessage)
	} else if agentId == a.GetTeammateWithMinTrust().ID {
		a.Modules.AgentParameters.UpdateTrustValue(senderID, positiveMessage)
	} 
}

// Done
func (a *AgentSOSA) HandleChangeBikeMessage(msg obj.ChangeBikeMessage) {
	sender := msg.GetSender()
	
	// if sender is a teammate...
	if slices.Contains(a.GetFellowBikers(), sender) {
		// if we think poorly of the bike, trust them more for leaving. otherwise trust them less
		if a.GetAverageTrustOnBike() < 0.5 {
			a.Modules.AgentParameters.UpdateTrustValue(sender.GetID(), positiveMessage)
		} else {
			a.Modules.AgentParameters.UpdateTrustValue(sender.GetID(), negativeMessage)
		}
	}
}

// Done
func (a *AgentSOSA) GetAllIterationMessages([]obj.IBaseBiker) []messaging.IMessage[obj.IBaseBiker] {

	wantToSendKickoutMessage := false
	wantToSendChangeBikeMessage := false

	if len(a.DecideKickOut()) != 0 {
		wantToSendKickoutMessage = true
	}

	if a.DecideAction() == 1 {
		wantToSendChangeBikeMessage = true
	}


	if wantToSendChangeBikeMessage && wantToSendKickoutMessage {

		kickoutMessage := a.CreatekickoutMessage()
		changeBikeMessage := a.CreateChangeBikeMessage()
		return []messaging.IMessage[obj.IBaseBiker]{kickoutMessage, changeBikeMessage}

	} else if wantToSendChangeBikeMessage {

		changeBikeMessage := a.CreateChangeBikeMessage()
		return []messaging.IMessage[obj.IBaseBiker]{changeBikeMessage}

	} else if wantToSendKickoutMessage {

		kickoutMessage := a.CreatekickoutMessage()
		return []messaging.IMessage[obj.IBaseBiker]{kickoutMessage}

	} else {

		return []messaging.IMessage[obj.IBaseBiker]{}

	}
}

// Done
func (a *AgentSOSA) CreatekickoutMessage() obj.KickoutAgentMessage {
	// this is only called when we decide we do want to kick someone out

	agentId := a.DecideKickOut()[0]

	return obj.KickoutAgentMessage{
		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikers()),
		AgentId:     agentId,
	}
}

// Done
func (a *AgentSOSA) CreateChangeBikeMessage() obj.ChangeBikeMessage {
	// only called when we have decided to change bikes

	// for now just tell fellow bikers which we want to move and mention a bike we want to go to. decidechangebike may need a bit of investigating.
	return obj.ChangeBikeMessage{
		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikers()),
		BikeId: a.DecideChangeBike(),
	}
}






















// -- LEGACY - dont know why it was added as seesm exact same as inherited basebikers one.

// // returns: slice of messages the agent is going to send
// func (a *AgentSOSA) GetAllMessages([]obj.IBaseBiker) []messaging.IMessage[obj.IBaseBiker] {
// 	// For team's agent add your own logic on chosing when your biker should send messages and which ones to send (return)
// 	wantToSendMsg := true
// 	if wantToSendMsg {
// 		kickoutMsg := a.CreatekickoutMessage()
// 		nextLootboxMsg := a.CreateNextLootboxMessage()
// 		changeBikeMessage := a.CreateChangeBikeMessage()
// 		forcesMsg := a.CreateForcesMessage()
// 		proposedLootboxMsg := a.CreateProposedLootboxMessage()
// 		return []messaging.IMessage[obj.IBaseBiker]{kickoutMsg, nextLootboxMsg, changeBikeMessage, forcesMsg, proposedLootboxMsg}
// 	}
// 	return []messaging.IMessage[obj.IBaseBiker]{}
// }

// func (a *AgentSOSA) HandleForcesMessage(msg obj.ForcesMessage) {
	// 	agentId := msg.AgentId
	// 	if agentId == uuid.Nil {
	// 		return
	// 	}
	
	// 	// fmt.Printf("[HandleForcesMessage] Received message from Agent %v\n", agentId)
	
	// 	// agentPosition := a.GetLocation()
	// 	// optimalLootbox := a.Modules.Environment.GetNearestLootboxByColor(agentId, a.GetColour())
	// 	// lootboxPosition := a.Modules.Environment.GetLootboxPos(optimalLootbox)
	// 	// optimalForces := a.Modules.Utils.GetForcesToTarget(agentPosition, lootboxPosition)
	// 	// eventValue := a.Modules.Utils.ProjectForce(optimalForces, msg.AgentForces)
	
	// 	// fmt.Printf("Agent Social Network Before: %v\n", a.Modules.SocialCapital.SocialNetwork)
	// 	a.Modules.AgentParameters.UpdateTrustValue(agentId, SocialEventValue_AgentSentMsg)
	// 	// a.Modules.SocialCapital.UpdateInstitution(agentId, InstitutionEventWeight_Adhereance, eventValue)
	// 	// fmt.Printf("Agent Social Network After: %v\n", a.Modules.SocialCapital.SocialNetwork)
	// }

	// func (a *AgentSOSA) CreateKickOffMessage() obj.KickoutAgentMessage {
// 	minTrustAgentStruct := a.Modules.AgentParameters.GetMinimumTrust()
// 	minTrustagentId := minTrustAgentStruct.ID
// 	kickOff := false
// 	if minTrustagentId != a.GetID() {
// 		kickOff = true
// 	}

// 	return obj.KickoutAgentMessage{
// 		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikers()),
// 		AgentId:     minTrustagentId,
// 		Kickout:     kickOff,
// 	}
// }