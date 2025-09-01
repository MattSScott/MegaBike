package ServerTests

import (
	"MegabikeFYPVersion/internal/server"
	"fmt"
	"testing"
)

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
