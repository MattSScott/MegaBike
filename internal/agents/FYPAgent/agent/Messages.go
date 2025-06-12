package agent

import (
	"MegabikeFYPVersion/internal/agents/FYPAgent/modules"
	obj "MegabikeFYPVersion/internal/common/objects"

	"github.com/MattSScott/basePlatformSOMAS/messaging"
)

// ----- Round -----

func (a *FYPAgent) HandleProposedLootboxMessage(msg obj.ProposedLootboxMessage) {
	// sender := msg.GetSender()

	// // if they proposed the same lootbox as us during this round, give them a SMALL boost in trust
	// if msg.Lootbox == a.GetRoundDirection() {
	// 	a.Modules.AgentParameters.UpdateTrustValue(sender.GetID(), 0.05*modules.PositiveMessage)
	// }
}

func (a *FYPAgent) HandleForcesMessage(msg obj.ForcesMessage) {

	// if they are pedalling less than 20%, lower trust. if greater than 80%, increase trust
	// sender := msg.GetSender()
	// if msg.AgentForces.Pedal < 0.2 {
	// 	a.Modules.AgentParameters.UpdateTrustValue(sender.GetID(), modules.NegativeMessage)
	// } else if msg.AgentForces.Pedal > 0.8 {
	// 	a.Modules.AgentParameters.UpdateTrustValue(sender.GetID(), modules.PositiveMessage)
	// }
}

func (a *FYPAgent) HandleConformMessage(msg obj.ConformMessage) {
	sender := msg.GetSender()

	// if they conformed, boost trust. if not, decrease trust
	if msg.DidConform {
		a.Modules.AgentParameters.UpdateTrustValue(sender.GetID(), modules.PositiveMessage)
	} else {
		a.Modules.AgentParameters.UpdateTrustValue(sender.GetID(), modules.NegativeMessage)
	}
}

func (a *FYPAgent) GetAllRoundMessages([]obj.IBaseBiker) []messaging.IMessage[obj.IBaseBiker] {

	// if we are on a bike, tell our teammates what lootbox / forces we aimed for this round. otherwise dont message people.
	if a.GetBikeStatus() {
		proposedLootboxMessage := a.CreateProposedLootboxMessage()
		forcesMsg := a.CreateForcesMessage()
		conformMsg := a.CreateConformMessage()

		return []messaging.IMessage[obj.IBaseBiker]{forcesMsg, proposedLootboxMessage, conformMsg}
	} else {
		return []messaging.IMessage[obj.IBaseBiker]{}
	}
}

func (a *FYPAgent) CreateProposedLootboxMessage() obj.ProposedLootboxMessage {
	// tell our fellow bikers what direction we chose this round
	return obj.ProposedLootboxMessage{
		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikersSlice()),
		Lootbox:     a.GetRoundDirection(),
	}
}

func (a *FYPAgent) CreateForcesMessage() obj.ForcesMessage {
	// tell our fellow bikers what forces we chose this round
	return obj.ForcesMessage{
		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikersSlice()),
		AgentForces: a.GetRoundForces(),
	}
}

func (a *FYPAgent) CreateConformMessage() obj.ConformMessage {
	// tell our fellow bikers whether we conformed this round
	return obj.ConformMessage{
		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikersSlice()),
		DidConform:  a.GetRoundDidConform(),
	}
}

// ----- Iteration (to rework) -----

// func (a *FYPAgent) HandleKickOutMessage(msg obj.KickoutAgentMessage) {

// 	senderID := msg.GetSender().GetID()
// 	agentId := msg.AgentId

// 	// if they want to kick us out, lower their trust score.
// 	// if they want to kick out the agent who we trust least, i.e. same opinion, increase their trust score.

// 	if agentId == a.GetID() {
// 		a.Modules.AgentParameters.UpdateTrustValue(senderID, modules.NegativeMessage)
// 	} else if agentId == a.GetTeammateWithMinTrust().ID {
// 		a.Modules.AgentParameters.UpdateTrustValue(senderID, modules.PositiveMessage)
// 	}
// }

// func (a *FYPAgent) HandleChangeBikeMessage(msg obj.ChangeBikeMessage) {
// 	sender := msg.GetSender()

// 	// if sender is a teammate...
// 	if _, ok := a.GetFellowBikers()[sender.GetID()]; ok {
// 		// if we think poorly of the bike, trust them more for leaving. otherwise trust them less
// 		if a.GetAverageTrustOnBike() < 0.5 {
// 			a.Modules.AgentParameters.UpdateTrustValue(sender.GetID(), modules.PositiveMessage)
// 		} else {
// 			a.Modules.AgentParameters.UpdateTrustValue(sender.GetID(), modules.NegativeMessage)
// 		}
// 	}
// }

// func (a *FYPAgent) GetAllIterationMessages([]obj.IBaseBiker) []messaging.IMessage[obj.IBaseBiker] {

// 	wantToSendKickoutMessage := false
// 	wantToSendChangeBikeMessage := false

// 	if len(a.DecideKickOut()) != 0 {
// 		wantToSendKickoutMessage = true
// 	}

// 	// to change
// 	if true {
// 		wantToSendChangeBikeMessage = true
// 	}

// 	if wantToSendChangeBikeMessage && wantToSendKickoutMessage {

// 		kickoutMessage := a.CreatekickoutMessage()
// 		changeBikeMessage := a.CreateChangeBikeMessage()
// 		return []messaging.IMessage[obj.IBaseBiker]{kickoutMessage, changeBikeMessage}

// 	} else if wantToSendChangeBikeMessage {

// 		changeBikeMessage := a.CreateChangeBikeMessage()
// 		return []messaging.IMessage[obj.IBaseBiker]{changeBikeMessage}

// 	} else if wantToSendKickoutMessage {

// 		kickoutMessage := a.CreatekickoutMessage()
// 		return []messaging.IMessage[obj.IBaseBiker]{kickoutMessage}

// 	} else {

// 		return []messaging.IMessage[obj.IBaseBiker]{}

// 	}
// }

// // func (a *FYPAgent) CreatekickoutMessage() obj.KickoutAgentMessage {
// // 	// this is only called when we decide we do want to kick someone out

// // 	agentId := a.DecideKickOut()[0]

// // 	return obj.KickoutAgentMessage{
// // 		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikersSlice()),
// // 		AgentId:     agentId,
// // 	}
// // }

// // func (a *FYPAgent) CreateChangeBikeMessage() obj.ChangeBikeMessage {

// // 	return obj.ChangeBikeMessage{
// // 		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](a, a.GetFellowBikersSlice()),
// // 		BikeId:      a.GetBike(),
// // 	}
// // }
