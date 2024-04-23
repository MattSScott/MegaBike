package server

import (
	"SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"

	"github.com/google/uuid"
)

type SimplifiedGameStateDump struct {
	Iterations []*SimplifiedIterationDump `json:"iteration"`
}

type SimplifiedIterationDump struct {
	Rounds          []*SimplifiedRoundDump `json:"round"`
	KickOffs        map[uuid.UUID]int      `json:"kickOffs"`
	AverageKickOffs float64                `json:"avgKickOffs"`
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
	OnBike          bool              `json:"on_bike"`
	AgentDirection  utils.Coordinates `json:"agentDirection"`
	Trustworthiness float64           `json:"trustworthiness"`
}

func NewSimplifiedGameStateDump() *SimplifiedGameStateDump {
	return &SimplifiedGameStateDump{
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
		LootGained:    bike.GetCurrentPool(),
	}
}

func (s *Server) GenerateIterationDump() *SimplifiedIterationDump {
	return &SimplifiedIterationDump{
		Rounds:          make([]*SimplifiedRoundDump, 0),
		KickOffs:        make(map[uuid.UUID]int),
		AverageKickOffs: 0.0,
	}
}

func (s *Server) GenerateAgentDump(agent objects.IBaseBiker) SimplfiedAgentDump {
	return SimplfiedAgentDump{
		OnBike:          agent.GetBikeStatus(),
		AgentDirection:  agent.GetForces().Force2Vec(),
		Trustworthiness: agent.GetTrustworthiness(),
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
