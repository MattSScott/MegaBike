package objects

import (
    "SOMAS2023/internal/clients/teamSOSA/agent"
    obj "SOMAS2023/internal/common/objects"
    "SOMAS2023/internal/server"
    "SOMAS2023/internal/common/utils"
    "testing"
    "github.com/google/uuid"
    "github.com/MattSScott/basePlatformSOMAS/messaging"
)

func TestRoundMessaging(t *testing.T) {
    // Setup
    s := server.GenerateServer()
    s.Initialize(3)

    agent1 := agent.NewAgentSOSA(obj.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), s))
    agent2 := agent.NewAgentSOSA(obj.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), s))

    t.Run("Test Forces Message", func(t *testing.T) {
        // Get initial trust value
        initialTrust := agent1.GetTrustOfAgent(agent2)

        // Create and send a forces message with high pedal force
        forces := utils.Forces{Pedal: 0.9}
        msg := obj.ForcesMessage{
            BaseMessage: messaging.CreateMessage[obj.IBaseBiker](agent2, []obj.IBaseBiker{agent1}),
            AgentForces: forces,
        }

        // Handle the message
        agent1.HandleForcesMessage(msg)

        // Check if trust increased
        newTrust := agent1.GetTrustOfAgent(agent2)
        if newTrust <= initialTrust {
            t.Errorf("Trust should have increased for high pedal force. Initial: %f, New: %f", initialTrust, newTrust)
        }
    })

    t.Run("Test Proposed Lootbox Message", func(t *testing.T) {
        // Get initial trust value
        initialTrust := agent1.GetTrustOfAgent(agent2)

        // Set same round direction for both agents
        lootboxID := uuid.New()
        agent1.SetRoundDirection(lootboxID)

        // Create and send proposed lootbox message
        msg := obj.ProposedLootboxMessage{
            BaseMessage: messaging.CreateMessage[obj.IBaseBiker](agent2, []obj.IBaseBiker{agent1}),
            Lootbox: lootboxID,
        }

        // Handle the message
        agent1.HandleProposedLootboxMessage(msg)

        // Check if trust increased
        newTrust := agent1.GetTrustOfAgent(agent2)
        if newTrust <= initialTrust {
            t.Errorf("Trust should have increased for agreeing on lootbox. Initial: %f, New: %f", initialTrust, newTrust)
        }
    })
}

func TestIterationMessaging(t *testing.T) {
    // Setup
    s := server.GenerateServer()
    s.Initialize(3)

    agent1 := agent.NewAgentSOSA(obj.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), s))
    agent2 := agent.NewAgentSOSA(obj.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), s))


    // Set agents as fellow bikers
    bikeID := uuid.New()
    agent1.SetBike(bikeID)
    agent2.SetBike(bikeID)
    agent1.ToggleOnBike()
    agent2.ToggleOnBike()

    t.Run("Test Kickout Message", func(t *testing.T) {
        // Get initial trust value
        initialTrust := agent1.GetTrustOfAgent(agent2)

        // Create and send kickout message targeting agent1
        msg := obj.KickoutAgentMessage{
            BaseMessage: messaging.CreateMessage[obj.IBaseBiker](agent2, []obj.IBaseBiker{agent1}),
            AgentId: agent1.GetID(),
        }

        // Handle the message
        agent1.HandleKickOutMessage(msg)

        // Check if trust decreased
        newTrust := agent1.GetTrustOfAgent(agent2)
        if newTrust >= initialTrust {
            t.Errorf("Trust should have decreased when receiving kickout message. Initial: %f, New: %f", initialTrust, newTrust)
        }
    })
}