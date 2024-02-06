package rules

import (
	"SOMAS2023/internal/common/objects"
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

func (r *Rule) EvaluateRule(agent objects.IBaseBiker) bool {
	var inputVector []float64 = make([]float64, len(r.ruleInputs))

	for i := range r.ruleInputs {
		ruleType := r.ruleInputs[i]
		inputVector[i] = inputGetter(ruleType, agent)
	}
	lMat := r.ruleMatrix
	rMat := mat.NewVecDense(len(inputVector), inputVector)

	var evalMat mat.Dense
	evalMat.Mul(lMat, rMat)

	for row := 0; row < len(inputVector); row++ {
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
