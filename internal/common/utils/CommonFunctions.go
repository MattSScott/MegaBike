package utils

import (
	"math/rand"
	"github.com/google/uuid"
)

// names for the agents
var (
	names = []string{
		"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Hannah",
		"Isaac", "Jack", "Karen", "Leo", "Mona", "Nina", "Oscar", "Paul",
		"Quinn", "Rachel", "Steve", "Tina", "Uma", "Victor", "Wendy", "Xander",
		"Yara", "Zane", "Amelia", "Benjamin", "Clara", "Daniel", "Elena", "Felix",
		"Gina", "Harry", "Ivy", "Jacob", "Kylie", "Liam", "Mia", "Noah",
		"Olivia", "Peter", "Quincy", "Rita", "Samuel", "Tara", "Ursula", "Vincent",
	}
	nameMap  = make(map[uuid.UUID]string)
	nameIndex = 0
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