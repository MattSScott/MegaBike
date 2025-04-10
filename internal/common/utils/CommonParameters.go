package utils

import "github.com/google/uuid"

// ----- Environment Parameters -----

const GridHeight float64 = 250.0
const GridWidth float64 = 250.0
const CollisionThreshold float64 = 7.0
const Epsilon float64 = 0.01 // tolerance for FP rounding and checking if == 1.0
const BikersOnBike = 8
const ReplenishEnergyEveryIteration = true
const ResetPointsEveryIteration = true
const RespawnEveryIteration = false
const Rounds = 100 // usually 100 

// ----- Server Parameters -----
const ReplenishLootBoxes bool = true
const ReplenishMegaBikes bool = true

// ----- Physics Parameters -----
const MassBike float64 = 1.0
const MassBiker float64 = 1.0
const MassAwdi float64 = 7.0
const BikerMaxForce float64 = 0.8 // The max force a biker can pedal
const AwdiMaxForce float64 = 1.0  // The awdi's force is equivalent to that of one biker agent going at maximum speed
const DragCoefficient float64 = 0.5 // Drag coefficient can be optimised in experimentation


// ----- Resources - Points and Energy -----
const PointsFromSameColouredLootBox = 500.0
const MovingDepletion float64 = 0.01 // proportionality of energy loss
const LimboEnergyPenalty float64 = 0.01 // amount of energy lost per round when off a bike
const DeliberativeDemocracyPenalty float64 = 0.01 // amount of energy lost per vote in a deliberative democracy


// ----- Awdi Behavior -----
const AwdiTargetsEmptyMegaBike bool = false
const AwdiOnlyTargetsStationaryMegaBike bool = false // if false, targeting slowest
const AwdiRemovesMegaBike bool = false

// ----- Voting Method Choice ------
type voteMethods int

const (
	PLURALITY voteMethods = iota
	RUNOFF
	BORDACOUNT
	INSTANTRUNOFF
	APPROVAL
	COPELANDSCORING
)

const VoteAction voteMethods = PLURALITY

// ----- Agent Names -----

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


// ----- Deprecated -----
// const LeadershipDemocracyPenalty float64 = 0.025  // amount of energy lost per vote in a leadership democracy