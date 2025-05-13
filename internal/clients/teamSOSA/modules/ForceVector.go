package modules

import (
	"SOMAS2023/internal/common/utils"
	"math"
)

type ForceVector struct {
	X float64
	Y float64
}

// constructor 
func GetForceVector(force utils.Forces) *ForceVector {
	pi_f64 := float64(math.Pi)
	return &ForceVector{
		X: force.Pedal * math.Cos(pi_f64*force.Turning.SteeringForce),
		Y: force.Pedal * math.Sin(pi_f64*force.Turning.SteeringForce),
	}
}

// returns: dot product of the vector with another
func (fv *ForceVector) Dot(vec ForceVector) float64 {
	return fv.X*vec.X + fv.Y*vec.Y
}

// returns: magnitude of the vector
func (fv *ForceVector) Magnitude() float64 {
	return math.Sqrt(fv.X*fv.X + fv.Y*fv.Y)
}

// returns: cosine similarity of the vector with another
func (fv *ForceVector) CosineSimilarity(vec ForceVector) float64 {
	if fv.Dot(vec) == 0 {
		return 0.0
	}
	return fv.Dot(vec) / (fv.Magnitude() * vec.Magnitude())
}

// returns: a force for the vector
func (fv *ForceVector) ConvertToForce() utils.Forces {
	return utils.Forces{
		Pedal: math.Min(math.Sqrt(math.Pow(fv.X, 2)+math.Pow(fv.Y, 2)), utils.BikerMaxForce),
		Turning: utils.TurningDecision{
			SteeringForce: math.Atan2(fv.Y, fv.X) / math.Pi,
		},
	}
}
