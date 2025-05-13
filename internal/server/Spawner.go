package server

import (
	"SOMAS2023/internal/clients/teamSOSA"
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"

	"fmt"
	baseserver "github.com/MattSScott/basePlatformSOMAS/BaseServer"
	"github.com/google/uuid"
)

type AgentInitFunction func(baseBiker *objects.BaseBiker) objects.IBaseBiker

// replace teamSOSA.GetBiker with nil to run basebiker experiments
var AgentInitFunctions = []AgentInitFunction{
	teamSOSA.GetBiker, 
}

// returns: a slice of agent generator-count pairs, i.e. a slice in which each element is a 2-tuple of the form (agent generator function, number to spawn in)
func (s *Server) GetAgentGenerators() []baseserver.AgentGeneratorCountPair[objects.IBaseBiker] {

	bikersPerTeam := *globals.BikerAgentCount / (len(AgentInitFunctions))
	extraBaseBikers := *globals.BikerAgentCount % (len(AgentInitFunctions))

	agentGenerators := []baseserver.AgentGeneratorCountPair[objects.IBaseBiker]{
		// Spawn base bikers
		baseserver.MakeAgentGeneratorCountPair(s.BikerAgentGenerator(nil), extraBaseBikers),
	}

	for _, initFunction := range AgentInitFunctions {
		agentGenerators = append(agentGenerators, baseserver.MakeAgentGeneratorCountPair(s.BikerAgentGenerator(initFunction), bikersPerTeam))
	}
	return agentGenerators
}

// helper function to spawn in bikes
func (s *Server) BikerAgentGenerator(initFunc func(baseBiker *objects.BaseBiker) objects.IBaseBiker) func() objects.IBaseBiker {
	return func() objects.IBaseBiker {
		baseBiker := objects.GetBaseBiker(utils.GenerateRandomColour(), uuid.New(), s)
		if initFunc == nil {
			return baseBiker
		} else {
			return initFunc(baseBiker)
		}
	}
}

// ----- Lootboxes -----

func (s *Server) spawnLootBox() {
	lootBox := objects.GetLootBox()
	s.lootBoxes[lootBox.GetID()] = lootBox
}

func (s *Server) replenishLootBoxes() {
	count := globals.LootBoxCount - len(s.lootBoxes)
	for i := 0; i < count; i++ {
		s.spawnLootBox()
	}
}

// ----- Megabikes -----

func (s *Server) spawnInitialMegaBikesAndRiders() {
	for i := 0; i < globals.MegaBikeCount; i++ {
		// increment i each time to spawn a megabike of each governance type
		governance := utils.Governance(i)
		s.spawnMegaBike(governance)
		fmt.Println("Spawning Megabike with Governance", governance)
	}

	bikeArray := make([]objects.IMegaBike, 0)

	for _, bike := range s.GetMegaBikes() {
		bikeArray = append(bikeArray, bike)
	}

	bikeAssignIdx := 0

	for _, agent := range s.GetAgentMap() {
		bike := bikeArray[bikeAssignIdx]
		if len(bike.GetAgents()) == 8 {
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
	neededBikes := globals.MegaBikeCount - len(s.megaBikes)
	for i := 0; i < neededBikes; i++ {
		s.spawnMegaBike(utils.One)
	}
}
