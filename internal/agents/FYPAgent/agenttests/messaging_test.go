package AgentTests

import (
	"MegabikeFYPVersion/internal/agents/FYPAgent/agent"
	obj "MegabikeFYPVersion/internal/common/objects"
	"MegabikeFYPVersion/internal/common/utils"
	"MegabikeFYPVersion/internal/server"
	"fmt"
	"testing"

	"github.com/MattSScott/basePlatformSOMAS/messaging"
	"github.com/google/uuid"
)

func TestRoundMessaging(t *testing.T) {
	// Setup
	s := server.GenerateServer()
	s.Initialize(3)

	agent1 := agent.NewFYPAgent(obj.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), s), 0.5)
	agent2 := agent.NewFYPAgent(obj.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), s), 0.5)

	t.Run("Test Conform Message", func(t *testing.T) {


		msg := obj.ConformMessage{
			BaseMessage: messaging.CreateMessage[obj.IBaseBiker](agent2, []obj.IBaseBiker{agent1}),
			DidConform: true,
		}

		// Handle the message
		agent1.HandleConformMessage(msg) // to get them to 0.5 baseline

			// Get initial trust value
		initialTrust := agent1.GetTrustOfAgent(agent2)
		
		agent1.HandleConformMessage(msg) // to increase it

		// Check if trust increased
		newTrust := agent1.GetTrustOfAgent(agent2)
		fmt.Println("old trust", initialTrust)
		fmt.Println("new trust", newTrust)
		if newTrust <= initialTrust {
			t.Errorf("Trust should have increased for high pedal force. Initial: %f, New: %f", initialTrust, newTrust)
		}
	})

	// t.Run("Test Proposed Lootbox Message", func(t *testing.T) {
	// 	// Get initial trust value
	// 	initialTrust := agent1.GetTrustOfAgent(agent2)

	// 	// Set same round direction for both agents
	// 	lootboxID := uuid.New()
	// 	agent1.SetRoundDirection(lootboxID)

	// 	// Create and send proposed lootbox message
	// 	msg := obj.ProposedLootboxMessage{
	// 		BaseMessage: messaging.CreateMessage[obj.IBaseBiker](agent2, []obj.IBaseBiker{agent1}),
	// 		Lootbox:     lootboxID,
	// 	}

	// 	// Handle the message
	// 	agent1.HandleProposedLootboxMessage(msg)

	// 	// Check if trust increased
	// 	newTrust := agent1.GetTrustOfAgent(agent2)
	// 	if newTrust <= initialTrust {
	// 		t.Errorf("Trust should have increased for agreeing on lootbox. Initial: %f, New: %f", initialTrust, newTrust)
	// 	}
	// })
}

// func TestIterationMessaging(t *testing.T) {
//     // Setup
//     s := server.GenerateServer()
//     s.Initialize(3)

//     agent1 := agent.NewFYPAgent(obj.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), s), 0.5)
//     agent2 := agent.NewFYPAgent(obj.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), s), 0.5)

//     // Set agents as fellow bikers
//     bikeID := uuid.New()
//     agent1.SetBike(bikeID)
//     agent2.SetBike(bikeID)
//     agent1.ToggleOnBike()
//     agent2.ToggleOnBike()

//     t.Run("Test Kickout Message", func(t *testing.T) {
//         // Get initial trust value
//         initialTrust := agent1.GetTrustOfAgent(agent2)

//         // Create and send kickout message targeting agent1
//         msg := obj.KickoutAgentMessage{
//             BaseMessage: messaging.CreateMessage[obj.IBaseBiker](agent2, []obj.IBaseBiker{agent1}),
//             AgentId: agent1.GetID(),
//         }

//         // Handle the message
//         agent1.HandleKickOutMessage(msg)

//         // Check if trust decreased
//         newTrust := agent1.GetTrustOfAgent(agent2)
//         if newTrust >= initialTrust {
//             t.Errorf("Trust should have decreased when receiving kickout message. Initial: %f, New: %f", initialTrust, newTrust)
//         }
//     })
// }
