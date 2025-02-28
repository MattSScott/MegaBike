package utils

import (
	"math/rand"
)

// GenerateRandomCoordinates creates random X and Y coordinates within the grid boundaries.
func GenerateRandomCoordinates() Coordinates {
	// Generate random coordinates
	return Coordinates{
		X: rand.Float64() * GridWidth,
		Y: rand.Float64() * GridHeight,
	}
}

// GenerateRandomColour returns a random colour for the agents
func GenerateRandomColour() Colour {
	// Generate a random index between 0 and the number of colours - 1.

	// STANDARD VERSION BELOW
	randomIndex := rand.Intn(int(NumOfColours))
	return Colour(randomIndex)

	// // // experiment - make them all the same colour
	// return Colour(2)
}

func GenerateRandomFloat(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
