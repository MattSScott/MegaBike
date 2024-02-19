package objects

import (
	utils "SOMAS2023/internal/common/utils"
	"math"

	"github.com/google/uuid"
)

type IMegaBike interface {
	IPhysicsObject
	AddAgent(biker IBaseBiker)
	RemoveAgent(bikerId uuid.UUID)
	GetAgents() []IBaseBiker
	UpdateMass()
	KickOutAgent(weights map[uuid.UUID]float64) []uuid.UUID
	GetGovernance() utils.Governance
	GetRuler() uuid.UUID
	SetGovernance(governance utils.Governance)
	SetRuler(ruler uuid.UUID)
	GetActiveRulesForAction(action Action) []*Rule
	AddToRuleMap(rule *Rule)
	ClearRuleMap()
	ViewLocalRuleMap() map[Action][]*Rule
	ActionIsValidForRuleset(action Action) bool
	ActionCompliesWithLinearRuleset() bool
}

// MegaBike will have the following forces
type MegaBike struct {
	*PhysicsObject
	agents              []IBaseBiker
	kickedOutCount      int
	governance          utils.Governance
	ruler               uuid.UUID
	globalRuleCacheView RuleCacheOperations
	activeRuleMap       map[Action][]*Rule
	linearRuleList      []*Rule
}

// GetMegaBike is a constructor for MegaBike that initializes it with a new UUID and default position.
func GetMegaBike(ruleCache RuleCacheOperations) *MegaBike {
	return &MegaBike{
		PhysicsObject:       GetPhysicsObject(utils.MassBike),
		governance:          utils.Democracy,
		ruler:               uuid.Nil,
		globalRuleCacheView: ruleCache,
		activeRuleMap:       make(map[Action][]*Rule),
		linearRuleList:      make([]*Rule, 0),
	}
}

// adds
func (mb *MegaBike) AddAgent(biker IBaseBiker) {
	mb.agents = append(mb.agents, biker)
}

// Remove agent from bike, given its ID
func (mb *MegaBike) RemoveAgent(bikerId uuid.UUID) {
	// Create a new slice to store the updated agents
	var updatedAgents []IBaseBiker

	// Iterate through the agents and copy them to the updatedAgents slice
	for _, agent := range mb.agents {
		if agent.GetID() != bikerId {
			updatedAgents = append(updatedAgents, agent)
		}
	}

	// Replace the mb.agents slice with the updatedAgents slice
	mb.agents = updatedAgents
}

func (mb *MegaBike) GetAgents() []IBaseBiker {
	return mb.agents
}

// Calculate the mass of the bike with all it's agents
func (mb *MegaBike) UpdateMass() {
	mass := utils.MassBike
	mass += float64(len(mb.agents))
	mb.mass = mass
}

// Calculates and returns the total force of the Megabike based on the Biker's force
func (mb *MegaBike) UpdateForce() {
	if len(mb.agents) == 0 {
		mb.force = 0.0
	}
	totalPedal := 0.0
	totalBrake := 0.0
	for _, agent := range mb.agents {
		force := agent.GetForces()

		if force.Pedal != 0 {
			totalPedal += float64(force.Pedal)
		} else {
			totalBrake += float64(force.Brake)
		}
	}
	mb.force = (float64(totalPedal) - float64(totalBrake))
}

// Calculates the final orientation of the Megabike, between -1 and 1 (-180° to 180°), given the Biker's Turning forces
func (mb *MegaBike) UpdateOrientation() {
	var xSum, ySum float64
	numOfSteeringAgents := 0

	for _, agent := range mb.agents {
		turningDecision := agent.GetForces().Turning
		if turningDecision.SteerBike {
			numOfSteeringAgents += 1

			// Ensure input is between -1 and 1 (wrap around if in excess)
			steeringForce := math.Mod(turningDecision.SteeringForce+1.0, 2.0) - 1.0

			// Convert steering force to cartesian coordinates and sum up
			angle := math.Pi * float64(steeringForce)
			xSum += math.Cos(angle) // X component of the vector
			ySum += math.Sin(angle) // Y component of the vector
		}
	}

	// Average x and y components and return polar form
	if numOfSteeringAgents > 0 {
		avgX := xSum / float64(numOfSteeringAgents)
		avgY := ySum / float64(numOfSteeringAgents)
		mb.orientation += math.Atan2(avgY, avgX) / math.Pi // Converts back to -1 to 1 range
	}
}

// gets the orientation of the megabike
func (mb *MegaBike) GetOrientation() float64 {
	return mb.orientation
}

// get the count of kicked out agents
func (mb *MegaBike) GetKickedOutCount() int {
	return mb.kickedOutCount
}

// only called for level 0 and level 1
func (mb *MegaBike) KickOutAgent(weights map[uuid.UUID]float64) []uuid.UUID {
	voteCount := make(map[uuid.UUID]float64)
	// Count votes for each agent
	for _, agent := range mb.agents {
		agentVotes := agent.VoteForKickout() // Assuming this now returns map[uuid.UUID]int
		for agentID, votes := range agentVotes {
			agentWeight := weights[agentID]
			if val, ok := voteCount[agentID]; ok {
				voteCount[agentID] = float64(val) + agentWeight*float64(votes)
			} else {
				voteCount[agentID] = float64(votes) * agentWeight
			}
		}
	}

	// Find all agents with votes > half the number of agents
	agentsToKickOut := make([]uuid.UUID, 0)
	for agentID, votes := range voteCount {
		if votes > float64(len(mb.agents))/2.0 {
			agentsToKickOut = append(agentsToKickOut, agentID)
		}
	}

	mb.kickedOutCount += len(agentsToKickOut)

	return agentsToKickOut
}

func (mb *MegaBike) GetGovernance() utils.Governance {
	return mb.governance
}

func (mb *MegaBike) GetRuler() uuid.UUID {
	return mb.ruler
}

func (mb *MegaBike) SetGovernance(governance utils.Governance) {
	mb.governance = governance
}

func (mb *MegaBike) SetRuler(ruler uuid.UUID) {
	mb.ruler = ruler
}

func (mb *MegaBike) GetActiveRulesForAction(action Action) []*Rule {
	output := []*Rule{}
	if action != AppliesAll {
		output = append(output, mb.activeRuleMap[AppliesAll]...)
	}

	output = append(output, mb.activeRuleMap[action]...)
	return output
}

func (mb *MegaBike) AddToRuleMap(rule *Rule) {
	category := rule.GetRuleAction()
	mb.activeRuleMap[category] = append(mb.activeRuleMap[category], rule)
}

func (mb *MegaBike) ClearRuleMap() {
	mb.activeRuleMap = make(map[Action][]*Rule)
}

func (mb *MegaBike) ActivateAllGlobalRules() {
	globalRuleView := mb.globalRuleCacheView
	for _, rule := range globalRuleView.ViewGlobalRuleCache() {
		mb.AddToRuleMap(rule)
		mb.linearRuleList = append(mb.linearRuleList, rule)
	}
}

func (mb *MegaBike) InitialiseRuleMap() {
	dist := 100.0
	mute := false

	ruleInputs := RuleInputs{Energy}
	ruleMat := RuleMatrix{{1, -dist}}
	ruleComps := RuleComparators{LEQ}

	rule := GenerateRule(Lootbox, "lootbox_dist", ruleInputs, ruleMat, ruleComps, mute)

	mb.AddToRuleMap(rule)
}

func (mb *MegaBike) ViewLocalRuleMap() map[Action][]*Rule {
	return mb.activeRuleMap
}

func (mb *MegaBike) ViewLinearRuleList() []*Rule {
	return mb.linearRuleList
}

func (mb *MegaBike) ActionCompliesWithLinearRuleset() bool {
	for _, r := range mb.linearRuleList {
		for _, agent := range mb.agents {
			if !r.EvaluateAgentRule(agent) {
				return false
			}
		}
	}

	return true
}

func (mb *MegaBike) ActionIsValidForRuleset(action Action) bool {
	rulesToTest := mb.activeRuleMap[action]

	for _, r := range rulesToTest {
		for _, agent := range mb.agents {
			if !r.EvaluateAgentRule(agent) {
				return false
			}
		}
	}

	return true
}
