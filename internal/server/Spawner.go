package server

import (
	fypagent "MegabikeFYPVersion/internal/agents/FYPAgent"
	"MegabikeFYPVersion/internal/common/objects"
	"MegabikeFYPVersion/internal/common/utils"

	"fmt"

	baseserver "github.com/MattSScott/basePlatformSOMAS/BaseServer"
	"github.com/google/uuid"
)

// ----- Agents -----

type AgentInitFunction func(baseBiker *objects.BaseBiker, tendency float64) objects.IBaseBiker

// returns: a slice of agent generator-count pairs, i.e. a slice in which each element is a 2-tuple of the form (agent generator function, number to spawn in)
func (s *Server) GetAgentGenerators() []baseserver.AgentGeneratorCountPair[objects.IBaseBiker] {

	numGoodAgents := int(float64(s.config.BikerAgentCount) * s.config.ProportionOfGoodAgents)
	numBadAgents := s.config.BikerAgentCount - numGoodAgents

	agentGenerators := []baseserver.AgentGeneratorCountPair[objects.IBaseBiker]{
		baseserver.MakeAgentGeneratorCountPair(s.BikerAgentGenerator(fypagent.GetBiker, s.config.GoodPlatonicTendency), numGoodAgents),
		baseserver.MakeAgentGeneratorCountPair(s.BikerAgentGenerator(fypagent.GetBiker, 1-s.config.GoodPlatonicTendency), numBadAgents),
	}

	return agentGenerators
}

// helper function to spawn in agents
func (s *Server) BikerAgentGenerator(initFunc AgentInitFunction, tendency float64) func() objects.IBaseBiker {
	return func() objects.IBaseBiker {
		baseBiker := objects.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), s)
		if initFunc == nil {
			return baseBiker
		} else {
			return initFunc(baseBiker, tendency)
		}
	}
}

// ----- Lootboxes -----

// spawns a lootbox
func (s *Server) spawnLootBox() {
	lootBox := objects.GetLootBox()
	s.lootBoxes[lootBox.GetID()] = lootBox
}

// replenish the number of lootboxes up to the max count
func (s *Server) replenishLootBoxes() {
	count := s.config.LootBoxCount - len(s.lootBoxes)
	for i := 0; i < count; i++ {
		s.spawnLootBox()
	}
}

// ----- Megabikes -----

// spawns all of the initial megabikes, an equal number per regime.
func (s *Server) spawnInitialMegaBikes() {

	numRegimes := 3
	numBikesPerRegime := s.config.MegaBikeCount / numRegimes

	for i := 0; i < numRegimes; i++ {
		for j := 0; j < numBikesPerRegime; j++ {
			governance := utils.Governance(i)
			s.spawnMegaBike(governance)
			fmt.Println("Spawning Megabike with Governance", governance)
		}
	}
}

// spawns a megabike
func (s *Server) spawnMegaBike(governance utils.Governance) {
	megaBike := objects.GetMegaBike(governance)
	s.megaBikes[megaBike.GetID()] = megaBike
}

// replenish the number of megabikes up to the max count
func (s *Server) replenishMegaBikes() {

	// currently not being called as awdi doesnt currently remove the megabike.
	// if it does, this function needs changing to spawn new megabikes in with the correct governance

	neededBikes := s.config.MegaBikeCount - len(s.megaBikes)
	for i := 0; i < neededBikes; i++ {
		s.spawnMegaBike(utils.One)
	}
}
