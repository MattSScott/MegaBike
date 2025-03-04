package server

import (
	"SOMAS2023/internal/clients/teamSOSA"
	"SOMAS2023/internal/common/globals"
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"

	baseserver "github.com/MattSScott/basePlatformSOMAS/BaseServer"
	"github.com/google/uuid"
	"fmt"
)

type AgentInitFunction func(baseBiker *objects.BaseBiker) objects.IBaseBiker

var AgentInitFunctions = []AgentInitFunction{
	teamSOSA.GetBiker, // Team SOSA
}

// BASEBIKER EXPERIMENTS (uncomment this and comment out the above to run base biker experiments)
// var AgentInitFunctions = []AgentInitFunction{
// 	nil,
// }

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

func (s *Server) spawnLootBox() {
	lootBox := objects.GetLootBox()
	s.lootBoxes[lootBox.GetID()] = lootBox
}

// replenishes lootboxes up to the externally set count
func (s *Server) replenishLootBoxes() {
	count := globals.LootBoxCount - len(s.lootBoxes)
	for i := 0; i < count; i++ {
		s.spawnLootBox()
	}
}

func (s *Server) spawnMegaBike(governance utils.Governance) {
	megaBike := objects.GetMegaBike(s, governance)
	s.megaBikes[megaBike.GetID()] = megaBike
	megaBike.InitialiseRuleMap()
}

func (s *Server) spawnInitialMegaBikesAndRiders() {
	for i := 0; i < globals.MegaBikeCount; i++ {
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

func (s *Server) replenishMegaBikes() {
	neededBikes := globals.MegaBikeCount - len(s.megaBikes)
	for i := 0; i < neededBikes; i++ {
		s.spawnMegaBike(utils.PerfectDemocracy) // needs changing but currently not called anywhere anyway
		fmt.Println("replenishing megabike!!!")
	}
}
