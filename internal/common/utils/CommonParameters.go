package utils

import "github.com/google/uuid"

// ----- Physical Parameters -----

const GridHeight float64 = 250.0
const GridWidth float64 = 250.0
const CollisionThreshold float64 = 7.0 // how close two objects have to be for a collision to be detected
const BikersOnBike = 8
const MassBike float64 = 1.0
const MassBiker float64 = 1.0
const MassAwdi float64 = 7.0
const BikerMaxForce float64 = 0.8
const AwdiMaxForce float64 = 1.0
const DragCoefficient float64 = 0.5
const MovingDepletion float64 = 0.01              // constant of proportionality for energy loss when pedalling
const LimboEnergyPenalty float64 = 0.05           // energy lost per round when off a bike
const DecisionPenalty float64 = 0.01			  // energy lost when an agent makes a decision
const PointsFromSameColouredLootBox = 500.0

// ----- Round and Iteration Parameters -----

const Rounds = 100
const ReplenishEnergyEveryIteration = true
const ResetPointsEveryIteration = true
const RespawnEveryIteration = false
const ReplenishLootBoxes bool = true
const ReplenishMegaBikes bool = true

// ----- Awdi Behavior -----

const AwdiTargetsEmptyMegaBike bool = false
const AwdiOnlyTargetsStationaryMegaBike bool = false // if false, targeting slowest
const AwdiRemovesMegaBike bool = false

// ----- Misc ------

const Epsilon float64 = 0.01 // tolerance for FP rounding and checking if == 1.0

// ----- Agent Names -----

var (
	names = []string{
		"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Hannah",
		"Isaac", "Jack", "Karen", "Leo", "Mona", "Nina", "Oscar", "Paul",
		"Quinn", "Rachel", "Steve", "Tina", "Uma", "Victor", "Wendy", "Xander",
		"Yara", "Zane", "Amelia", "Benjamin", "Clara", "Daniel", "Elena", "Felix",
		"Gina", "Harry", "Ivy", "Jacob", "Kylie", "Liam", "Mia", "Noah",
		"Olivia", "Peter", "Quincy", "Rita", "Samuel", "Tara", "Ursula", "Vincent",
	}
	nameMap   = make(map[uuid.UUID]string)
	nameIndex = 0
)
