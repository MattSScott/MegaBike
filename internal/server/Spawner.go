package server

import (
	"SOMAS2023/internal/clients/teamSOSA"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"

	// "fmt"
	baseserver "github.com/MattSScott/basePlatformSOMAS/BaseServer"
	"github.com/google/uuid"
)

type AgentInitFunction func(baseBiker *objects.BaseBiker, tendency float64) objects.IBaseBiker

// returns: a slice of agent generator-count pairs, i.e. a slice in which each element is a 2-tuple of the form (agent generator function, number to spawn in)
func (s *Server) GetAgentGenerators() []baseserver.AgentGeneratorCountPair[objects.IBaseBiker] {

	numGoodAgents := int(float64(s.config.BikerAgentCount) * s.config.ProportionOfGoodAgents)
	numBadAgents := s.config.BikerAgentCount - numGoodAgents

	agentGenerators := []baseserver.AgentGeneratorCountPair[objects.IBaseBiker]{
		baseserver.MakeAgentGeneratorCountPair(s.BikerAgentGenerator(teamSOSA.GetBiker, s.config.GoodPlatonicTendency), numGoodAgents),
		baseserver.MakeAgentGeneratorCountPair(s.BikerAgentGenerator(teamSOSA.GetBiker, 1-s.config.GoodPlatonicTendency), numBadAgents),
	}

	return agentGenerators
}

// helper function to spawn in bikes
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

func (s *Server) spawnLootBox() {
	lootBox := objects.GetLootBox()
	s.lootBoxes[lootBox.GetID()] = lootBox
}

func (s *Server) replenishLootBoxes() {
	count := s.config.LootBoxCount - len(s.lootBoxes)
	for i := 0; i < count; i++ {
		s.spawnLootBox()
	}
}

// ----- Megabikes -----

// improved so it is capable of spawning multiple bikes for each regime, e.g. 9 bikes total 3 of each
func (s *Server) spawnInitialMegaBikesAndRiders() {
	numRegimes := 3
	numBikesPerRegime := s.config.MegaBikeCount / numRegimes

	for i := 0; i < numRegimes; i++ {
		for j := 0; j < numBikesPerRegime; j++ {
			governance := utils.Governance(i)
			s.spawnMegaBike(governance)
			// fmt.Println("Spawning Megabike with Governance", governance)
		}

	}

	bikeArray := make([]objects.IMegaBike, 0)

	for _, bike := range s.GetMegaBikes() {
		bikeArray = append(bikeArray, bike)
	}

	bikeAssignIdx := 0

	for _, agent := range s.GetAgentMap() {
		bike := bikeArray[bikeAssignIdx]
		if len(bike.GetAgents()) == utils.BikersOnBike {
			bikeAssignIdx += 1
			bike = bikeArray[bikeAssignIdx]
		}
		s.AddAgentToBike(agent, bike)
	}

}

func (s *Server) spawnMegaBike(governance utils.Governance) {
	megaBike := objects.GetMegaBike(s, governance)
	s.megaBikes[megaBike.GetID()] = megaBike
	megaBike.InitialiseRuleMap()
}

func (s *Server) replenishMegaBikes() {
	// currently not being called as awdi doesnt currently remove the megabike.
	// if it does, this function needs changing to spawn new megabikes in with the correct governance
	neededBikes := s.config.MegaBikeCount - len(s.megaBikes)
	for i := 0; i < neededBikes; i++ {
		s.spawnMegaBike(utils.One)
	}
}
