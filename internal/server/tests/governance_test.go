package server_test

import (
	// "SOMAS2023/internal/common/objects"
	"SOMAS2023/internal/common/utils"
	"SOMAS2023/internal/server"
	"fmt"
	"testing"

	// "github.com/google/uuid"
	"github.com/google/uuid"
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

func TestRunRepresentativeDirectionDecision(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)

	// perform role assignment, then get them to decide on a direction. if nil, something wrong. otherwise pass!
	for _, bike := range s.GetMegaBikes(){
		governance := bike.GetGovernance()
		if governance == utils.Some || governance == utils.One {
			s.PerformRoleAssignment(bike)
			lootbox := s.RunRepresentativeDirectionDecision(bike)
			fmt.Printf("\n bike targetting lootbox with id %v", lootbox)
			if lootbox == uuid.Nil {
				t.Error("targetting nil lootbox")
			}
		}
	}

	fmt.Println("Representative direction decision process runs successfully")
}

func TestRunDemocraticDirectionDecision(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)


	for _, bike := range s.GetMegaBikes(){
		governance := bike.GetGovernance()
		if governance == utils.Many {
			lootbox := s.RunDemocraticDirectionDecision(bike)
			fmt.Printf("\n bike targetting lootbox with id %v", lootbox) 
			if lootbox == uuid.Nil {
				t.Error("targetting nil lootbox")
			}
		}
	}

	fmt.Println("Democratic direction decision process runs successfully")
}

// needs work
func TestHandleDepartingRepresentative(t *testing.T) {
	iterations := 3
	s := server.GenerateServer()
	s.Initialize(iterations)


	
}
