package server

import (
	"MegabikeFYPVersion/internal/common/objects"
	"MegabikeFYPVersion/internal/common/utils"

	"github.com/google/uuid"
)

type SimplifiedGameStateDump struct {
	AgentUUIDS []uuid.UUID                `json:"agentUUIDS"`
	BikeUUIDS  []uuid.UUID                `json:"bikeUUIDS"`
	Iterations []*SimplifiedIterationDump `json:"iterations"`
}

type SimplifiedIterationDump struct {
	Rounds           []*SimplifiedRoundDump    `json:"rounds"`
	KickOffs         map[uuid.UUID]int         `json:"kickOffs"`
	AverageKickOffs  float64                   `json:"avgKickOffs"`
	JoiningDecisions map[uuid.UUID]bool        `json:"joiningDecisions"` // agent -> flag indicating whether their joining decision had b>r
	AgentColleagues  map[uuid.UUID][]uuid.UUID `json:"agentColleagues"`  // maps agents to which people they have on their bike each iteration
	AgentBikes       map[uuid.UUID]uuid.UUID   `json:"agentBikes"`       // map agents to which bike they end up on
}

type SimplifiedRoundDump struct {
	Bikes map[uuid.UUID]SimplfiedBikeDump `json:"bikes"`
}

type SimplfiedBikeDump struct {
	Agents        map[uuid.UUID]SimplfiedAgentDump `json:"agents"`
	BikeDirection utils.Coordinates                `json:"bikeDirection"`
	LootGained    float64                          `json:"lootGained"`
}

type SimplfiedAgentDump struct {
	OnBike         bool              `json:"on_bike"`
	AgentDirection utils.Coordinates `json:"agentDirection"`
}

func NewSimplifiedGameStateDump() *SimplifiedGameStateDump {
	return &SimplifiedGameStateDump{
		AgentUUIDS: make([]uuid.UUID, 0),
		BikeUUIDS:  make([]uuid.UUID, 0),
		Iterations: make([]*SimplifiedIterationDump, 0),
	}
}

func (gsd *SimplifiedGameStateDump) AddIterationToGameState(iterationDump *SimplifiedIterationDump) {
	gsd.Iterations = append(gsd.Iterations, iterationDump)
}

func (sid *SimplifiedIterationDump) AddRoundToIteration(roundDump *SimplifiedRoundDump) {
	sid.Rounds = append(sid.Rounds, roundDump)
}

func (s *Server) GenerateBikeDump(bike objects.IMegaBike) SimplfiedBikeDump {

	agentArray := make(map[uuid.UUID]SimplfiedAgentDump)
	for _, ag := range bike.GetAgents() {
		agentArray[ag.GetID()] = s.GenerateAgentDump(ag)
	}

	bikeOrientationData := utils.Forces{}
	bikeOrientationData.Brake = 0
	bikeOrientationData.Pedal = bike.GetForce()
	bikeOrientationData.Turning = utils.TurningDecision{SteerBike: true, SteeringForce: bike.GetOrientation()}

	return SimplfiedBikeDump{
		Agents:        agentArray,
		BikeDirection: bikeOrientationData.Force2Vec(),
	}
}

// returns: a zero-valued iteration dump
func (s *Server) GenerateIterationDump() *SimplifiedIterationDump {
	return &SimplifiedIterationDump{
		Rounds:           make([]*SimplifiedRoundDump, 0),
		KickOffs:         make(map[uuid.UUID]int),
		AverageKickOffs:  0.0,
		JoiningDecisions: make(map[uuid.UUID]bool),
		AgentColleagues:  make(map[uuid.UUID][]uuid.UUID),
		AgentBikes:       make(map[uuid.UUID]uuid.UUID),
	}
}

func (s *Server) GenerateAgentDump(agent objects.IBaseBiker) SimplfiedAgentDump {
	return SimplfiedAgentDump{
		OnBike:         agent.GetBikeStatus(),
		AgentDirection: agent.GetForces().Force2Vec(),
	}
}

func (s *Server) GenerateRoundDump() *SimplifiedRoundDump {
	bikeArray := make(map[uuid.UUID]SimplfiedBikeDump)

	for id, bike := range s.GetMegaBikes() {
		bikeArray[id] = s.GenerateBikeDump(bike)
	}

	return &SimplifiedRoundDump{
		Bikes: bikeArray,
	}
}
