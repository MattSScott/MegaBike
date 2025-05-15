package utils

import (
	"math/rand"
	"github.com/google/uuid"
)

// returns: random X and Y coordinates within the grid boundaries.
func GenerateRandomCoordinates() Coordinates {
	return Coordinates{
		X: rand.Float64() * GridWidth,
		Y: rand.Float64() * GridHeight,
	}
}

// returns: a random colour for the agents
func GenerateRandomColour() Colour {
	// Generate a random index between 0 and the number of colours - 1.
	randomIndex := rand.Intn(int(NumOfColours))
	return Colour(randomIndex)
}

// returns: a random float between two floats
func GenerateRandomFloat(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// returns: the name of an agent given its uuid
func TranslateToName(id uuid.UUID) string {
	if name, exists := nameMap[id]; exists {
		return name
	}

	if nameIndex >= len(names) {
		return "Unknown" // Fallback if all names are used
	}

	name := names[nameIndex]
	nameMap[id] = name
	nameIndex++
	return name
}

// returns: a value for the platonic tendency of an agent
func GeneratePlatonicTendency(goodPlatonicTendency float64, proportionOfGoodAgents float64) float64 {

	r := rand.Float64()

	if r < proportionOfGoodAgents {
		// return the good value
		return goodPlatonicTendency
	} else {
		// return the bad value
		return 1 - goodPlatonicTendency
	}
}