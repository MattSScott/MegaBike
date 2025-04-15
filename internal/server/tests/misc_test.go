package server_test

import (
	// "SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/server"
	"fmt"
	"testing"

	// "github.com/google/uuid"
	// "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRepresentativeSelection(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	// for each bike:
	// 1. perform representative selection and setting if that bikes governance needs so
	// 2. check that the bike has the correct number of representatives
	for _, bike := range s.GetMegaBikes() {
		agents := bike.GetAgents()
		governance := bike.GetGovernance()
		if governance == utils.Some || governance == utils.One {
			reps := s.RepresentativeSelection(agents, governance)
			bike.SetRepresentatives(reps)
		}

		if governance == utils.Many {
			// shouldnt be any reps for "many" bikes
			assert.Equal(t, 0, len(bike.GetRepresentatives()))
		} else if governance == utils.Some {
			assert.Equal(t, 3, len(bike.GetRepresentatives()))
		} else if governance == utils.One {
			assert.Equal(t, 1, len(bike.GetRepresentatives()))
		} else{
			panic("dealing with invalid form of governance")
		}
	}
	
	fmt.Printf("\nRepresentative Selection Successful \n")

}

func TestGetRandomBikeID(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	bike := s.GetRandomBikeId()
	_, exists := s.GetMegaBikes()[bike]
	if !exists {
		t.Error("returned bike is not in ")
	}
	fmt.Printf("\nGet random ID passed \n")
}