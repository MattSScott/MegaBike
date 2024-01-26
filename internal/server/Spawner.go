package server

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"

	baseserver "github.com/MattSScott/basePlatformSOMAS/BaseServer"
	"github.com/google/uuid"
)

type AgentInitFunction func(baseBiker *objects.BaseBiker) objects.IBaseBiker

// COHORT EXPERIMENTS
var AgentInitFunctions = []AgentInitFunction{
	// team2.GetBiker, // Team 2
}

// BASEBIKER EXPERIMENTS (uncomment this and comment out the above to run base biker experiments)
// var AgentInitFunctions = []AgentInitFunction{
// 	nil,
// }

func (s *Server) GetAgentGenerators() []baseserver.AgentGeneratorCountPair[objects.IBaseBiker] {

	bikersPerTeam := BikerAgentCount / (len(AgentInitFunctions) + 1)
	extraBaseBikers := BikerAgentCount % (len(AgentInitFunctions) + 1)
	agentGenerators := []baseserver.AgentGeneratorCountPair[objects.IBaseBiker]{
		// Spawn base bikers
		baseserver.MakeAgentGeneratorCountPair(s.BikerAgentGenerator(nil), bikersPerTeam+extraBaseBikers),
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
	count := LootBoxCount - len(s.lootBoxes)
	for i := 0; i < count; i++ {
		s.spawnLootBox()
	}
}

func (s *Server) spawnMegaBike() {
	megaBike := objects.GetMegaBike()
	s.megaBikes[megaBike.GetID()] = megaBike
}

func (s *Server) replenishMegaBikes() {
	neededBikes := MegaBikeCount - len(s.megaBikes)
	for i := 0; i < neededBikes; i++ {
		s.spawnMegaBike()
	}
}
