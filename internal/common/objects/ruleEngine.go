package objects

import (
	"SOMAS2023/internal/common/physics"
	"errors"

	"github.com/google/uuid"
	"gonum.org/v1/gonum/mat"
)

func (r *Rule) ToggleRuleMutability() {
	r.isMutable = !r.isMutable
}

func (r *Rule) GetRuleID() uuid.UUID {
	return r.ruleID
}

func (r *Rule) GetRuleName() string {
	return r.ruleName
}

func (r *Rule) GetRuleAction() Action {
	return r.action
}

func (r *Rule) GetRuleInputs() []RuleInput {
	return r.ruleInputs
}

func (r *Rule) GetRuleMatrix() RuleMatrix {
	return r.ruleMatrix
}

func (r *Rule) GetRuleComparators() []Comparator {
	return r.ruleComparators
}

func (r *Rule) UpdateRuleMatrix(newRuleMatrix RuleMatrix) error {
	if !r.isMutable {
		return errors.New("rule is (currently) immutable")
	}

	if len(newRuleMatrix) != len(r.ruleMatrix) {
		return errors.New("new and old matrix dimensions must match")
	}

	if len(newRuleMatrix[0]) != len(r.ruleMatrix[0]) {
		return errors.New("new and old matrix dimensions must match")
	}

	r.ruleMatrix = newRuleMatrix
	return nil
}

func (r *Rule) EvaluateAgentInputs(agent IBaseBiker) []float64 {
	var inputVector []float64 = make([]float64, len(r.ruleInputs))

	for i := range r.ruleInputs {
		ruleType := r.ruleInputs[i]
		inputVector[i] = inputGetter(ruleType, agent)
	}

	// supply constant to vector
	inputVector = append(inputVector, 1)
	return inputVector
}

func (r *Rule) EvaluateTestLootboxRuleInputs(bike IMegaBike, lootbox ILootBox) []float64 {
	bPos := bike.GetPosition()
	lPos := lootbox.GetPosition()

	dPos := physics.ComputeDistance(bPos, lPos)

	return []float64{
		dPos,
		1,
	}
}

func (r *Rule) EvaluateAgentRule(agent IBaseBiker) bool {
	inputVector := r.EvaluateAgentInputs(agent)
	return r.EvaluateRule(inputVector)
}

func (r *Rule) EvaluateLootboxRule(bike IMegaBike, lootbox ILootBox) bool {
	inputVector := r.EvaluateTestLootboxRuleInputs(bike, lootbox)
	return r.EvaluateRule(inputVector)
}

func (r *Rule) EvaluateRule(parsedInputs []float64) bool {
	lMat := r.ruleMatrix
	rMat := mat.NewVecDense(len(parsedInputs), parsedInputs)

	var evalMat mat.Dense
	evalMat.Mul(lMat, rMat)

	for row := 0; row < len(r.ruleMatrix); row++ {
		clauseResult := valueComparator(r.ruleComparators[row], evalMat.At(row, 0))
		if !clauseResult {
			return false
		}
	}
	return true
}

func GenerateRule(action Action, name string, ruleInputs RuleInputs, ruleMatrix RuleMatrix, ruleComps RuleComparators, mutability bool) *Rule {
	return &Rule{
		ruleID:          uuid.New(),
		ruleName:        name,
		isMutable:       mutability,
		action:          action,
		ruleInputs:      ruleInputs,
		ruleMatrix:      ruleMatrix,
		ruleComparators: ruleComps,
	}
}

func GenerateNullPassingRule() *Rule {
	ruleInps := RuleInputs{Location, Energy, Points, Colour}
	ruleMatrix := [][]float64{{0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}}
	ruleComparators := RuleComparators{EQ, EQ, EQ}
	return &Rule{
		ruleID:          uuid.New(),
		ruleName:        "null_passing_rule",
		isMutable:       false,
		action:          AppliesAll,
		ruleInputs:      ruleInps,
		ruleMatrix:      ruleMatrix,
		ruleComparators: ruleComparators,
	}
}

func GenerateNullPassingRule2() *Rule {
	ruleInps := RuleInputs{Location, Energy, Points, Colour}
	rDim := 1000
	ruleMatrix := make([][]float64, rDim)
	ruleComparators := make(RuleComparators, rDim)
	for i := 0; i < rDim; i++ {
		ruleMatrix[i] = make([]float64, 5)
		ruleComparators[i] = EQ
	}
	return &Rule{
		ruleID:          uuid.New(),
		ruleName:        "null_passing_rule",
		isMutable:       false,
		action:          AppliesAll,
		ruleInputs:      ruleInps,
		ruleMatrix:      ruleMatrix,
		ruleComparators: ruleComparators,
	}
}

func GenerateNullPassingRuleForAction(action Action) *Rule {
	ruleInps := RuleInputs{Location, Energy, Points, Colour}
	ruleMatrix := [][]float64{{0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}}
	ruleComparators := RuleComparators{EQ, EQ, EQ}
	return &Rule{
		ruleID:          uuid.New(),
		ruleName:        "null_passing_rule",
		isMutable:       false,
		action:          action,
		ruleInputs:      ruleInps,
		ruleMatrix:      ruleMatrix,
		ruleComparators: ruleComparators,
	}
}

func LinguisticNullRuleCheck(agent IBaseBiker) bool {
	for i := 0; i < 1000; i++ {
		lCheck := 0 * locationGetter(agent)
		eCheck := 0 * energyGetter(agent)
		pCheck := 0 * pointsGetter(agent)
		colCheck := 0 * colourGetter(agent)
		constCheck := 0 * 1.0

		if lCheck+eCheck+pCheck+colCheck+constCheck != 0 {
			return false
		}
	}
	return true
}
