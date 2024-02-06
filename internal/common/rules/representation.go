package rules

import (
	"github.com/google/uuid"
	"gonum.org/v1/gonum/mat"
)

type Action int
type RuleInput int
type Comparator int
type RuleInputs []RuleInput
type RuleMatrix [][]float64
type RuleComparators []Comparator

func (rm RuleMatrix) Dims() (r, c int) {
	return len(rm), len(rm[0])
}

func (rm RuleMatrix) At(i, j int) float64 {
	return rm[i][j]
}

func (rm RuleMatrix) T() mat.Matrix {
	return rm
}

const (
	MoveBike Action = iota
	KickAgent
	Allocation
	Lootbox
)

const (
	Forces RuleInput = iota
	Colour
	Location
	Energy
	Points
)

const (
	EQ Comparator = iota
	GT
	LT
	GEQ
	LEQ
)

type Rule struct {
	ruleID          uuid.UUID
	ruleName        string
	isMutable       bool
	action          Action
	ruleInputs      RuleInputs
	ruleMatrix      RuleMatrix
	ruleComparators RuleComparators
}
