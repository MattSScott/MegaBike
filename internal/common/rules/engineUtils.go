package rules

import (
	"SOMAS2023/internal/common/objects"
	"math"
)

func forceGetter(agent objects.IBaseBiker) float64 {
	return agent.GetForces().Pedal
}

func colourGetter(agent objects.IBaseBiker) float64 {
	return float64(agent.GetColour())
}

func locationGetter(agent objects.IBaseBiker) float64 {
	loco := agent.GetLocation()
	dx := loco.X * loco.X
	dy := loco.Y * loco.Y
	return math.Sqrt(dx + dy)
}

func energyGetter(agent objects.IBaseBiker) float64 {
	return agent.GetEnergyLevel()
}

func pointsGetter(agent objects.IBaseBiker) float64 {
	return float64(agent.GetPoints())
}

// generate numerical output of interface rule
func inputGetter(rule RuleInput, agent objects.IBaseBiker) float64 {
	switch rule {
	case Forces:
		return forceGetter(agent)
	case Colour:
		return colourGetter(agent)
	case Location:
		return locationGetter(agent)
	case Energy:
		return energyGetter(agent)
	case Points:
		return pointsGetter(agent)
	default:
		return 0.0
	}
}

func valueComparator(cmp Comparator, input float64) bool {
	switch cmp {
	case EQ:
		return input == 0
	case GT:
		return input > 0
	case LT:
		return input < 0
	case GEQ:
		return input >= 0
	case LEQ:
		return input <= 0
	default:
		return true
	}
}
