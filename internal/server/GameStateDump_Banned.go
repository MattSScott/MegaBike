package server

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/common/voting"

	"github.com/MattSScott/basePlatformSOMAS/messaging"
	"github.com/google/uuid"
)

// in order to limit access to certain functions that are part of the IBaseBiker and IPhysicsObject interfaces
// bikers have access to a gamestate which implements the interface but imposes some
// restrictions on the functions that can be called

const bannedFunctionErrorMessage = "you're not allowed to call this"

func (o PhysicsObjectDump) SetPhysicalState(state utils.PhysicalState) {
	panic(bannedFunctionErrorMessage)
}

func (o PhysicsObjectDump) UpdateForce() {
	panic(bannedFunctionErrorMessage)
}

func (o PhysicsObjectDump) UpdateOrientation() {
	panic(bannedFunctionErrorMessage)
}

func (o PhysicsObjectDump) CheckForCollision(objects.IPhysicsObject) bool {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) GetAllMessages([]objects.IBaseBiker) []messaging.IMessage[objects.IBaseBiker] {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) UpdateAgentInternalState() {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) DecideGovernance() utils.Governance {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) DecideAction() objects.BikerAction {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) DecideForce(uuid.UUID) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) GetForces() utils.Forces {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) DecideJoining([]uuid.UUID) map[uuid.UUID]bool {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) ChangeBike() uuid.UUID {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) ProposeDirection() uuid.UUID {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) ProposeNewRadius(float64) float64 {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) ProposeDirectionFromSubset(subset map[uuid.UUID]objects.ILootBox) uuid.UUID {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) FinalDirectionVote(map[uuid.UUID]uuid.UUID) voting.LootboxVoteMap {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) DecideAllocation() voting.IdVoteMap {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) VoteForKickout() map[uuid.UUID]int {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) VoteDictator() voting.IdVoteMap {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) VoteLeader() voting.IdVoteMap {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) DictateDirection() uuid.UUID {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) LeadDirection() uuid.UUID {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) SetBike(uuid.UUID) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) SetForces(utils.Forces) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) UpdateColour(utils.Colour) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) SetDeterministicColour(utils.Colour) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) UpdatePoints(int) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) UpdateEnergyLevel(float64) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) ToggleOnBike() {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) QueryReputation(uuid.UUID) float64 {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) SetReputation(uuid.UUID, float64) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) HandleGovernanceMessage(msg objects.GovernanceMessage) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) HandleJoiningMessage(msg objects.JoiningAgentMessage) {
	panic(bannedFunctionErrorMessage)
}
func (a AgentDump) HandleReputationMessage(msg objects.ReputationOfAgentMessage) {
	panic(bannedFunctionErrorMessage)
}
func (a AgentDump) HandleLootboxMessage(msg objects.LootboxMessage) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) HandleKickoutMessage(msg objects.KickoutAgentMessage) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) HandleForcesMessage(msg objects.ForcesMessage) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) HandleVoteGovernanceMessage(msg objects.VoteGoveranceMessage) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) HandleVoteLootboxDirectionMessage(msg objects.VoteLootboxDirectionMessage) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) HandleVoteRulerMessage(msg objects.VoteRulerMessage) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) HandleVoteKickoutMessage(msg objects.VoteKickoutMessage) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) HandleVoteAllocationMessage(msg objects.VoteAllocationMessage) {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) DecideDictatorAllocation() voting.IdVoteMap {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) DecideKickOut() []uuid.UUID {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) DecideWeights(action utils.Action) map[uuid.UUID]float64 {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) ResetPoints() {
	panic(bannedFunctionErrorMessage)
}

func (a AgentDump) GetTrustworthiness() float64 {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) AddAgent(objects.IBaseBiker) {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) RemoveAgent(uuid.UUID) {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) UpdateMass() {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) KickOutAgent(weights map[uuid.UUID]float64) []uuid.UUID {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) SetGovernance(utils.Governance) {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) SetRuler(uuid.UUID) {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) AddToRuleMap(rule *objects.Rule) {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) GetActiveRulesForAction(action objects.Action) []*objects.Rule {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) ViewLocalRuleMap() map[objects.Action][]*objects.Rule {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) ClearRuleMap() {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) ActionIsValidForRuleset(action objects.Action) bool {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) ActionCompliesWithLinearRuleset() bool {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) GetCurrentPool() float64 {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) ResetCurrentPool() {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) UpdateCurrentPool(float64) {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) GetKickedOutCount() int {
	panic(bannedFunctionErrorMessage)
}

func (b BikeDump) ResetKickedOutCount() {
	panic(bannedFunctionErrorMessage)
}
