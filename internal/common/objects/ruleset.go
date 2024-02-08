package objects

import "github.com/google/uuid"

type GlobalRuleCache struct {
	stratifiedRuleSet map[Action]([]*Rule) // rules stratified by relevant action
	rawRuleSet        map[uuid.UUID]*Rule  // hashed lookup of all rules
}

func (grc *GlobalRuleCache) GetRelevantRulesFromAction(action Action) []*Rule {
	return grc.stratifiedRuleSet[action]
}

func (grc *GlobalRuleCache) GetRuleByID(id uuid.UUID) *Rule {
	return grc.rawRuleSet[id]
}

func (grc *GlobalRuleCache) AddRuleToCache(rule *Rule) {
	grc.rawRuleSet[rule.GetRuleID()] = rule
	grc.stratifiedRuleSet[rule.GetRuleAction()] = append(grc.stratifiedRuleSet[rule.GetRuleAction()], rule)

}

func (grc *GlobalRuleCache) ViewGlobalRuleSet() map[uuid.UUID]*Rule {
	return grc.rawRuleSet
}

func GenerateGlobalRuleCache() *GlobalRuleCache {
	return &GlobalRuleCache{
		stratifiedRuleSet: make(map[Action][]*Rule),
		rawRuleSet:        make(map[uuid.UUID]*Rule),
	}
}

// think about if actually needed
// func (grc *GlobalRuleCache) DeleteRuleByID(id uuid.UUID) {
// 	return grc.rawRuleSet[id]
// }
