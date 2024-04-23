package server

import (
	"SOMAS2023/internal/common/objects"
	"fmt"
	"time"
)

func (s *Server) runBenchmarkingSession(newRep bool, iters int) {
	if newRep {
		rule := objects.GenerateNullPassingRule2()
		for i := 0; i < iters; i++ {
			for _, ag := range s.GetAgentMap() {
				rule.EvaluateAgentRule(ag)
			}
		}
	} else {
		for i := 0; i < iters; i++ {
			for _, ag := range s.GetAgentMap() {
				objects.LinguisticNullRuleCheck(ag)
			}
		}
	}
}

func (s *Server) TimeRuleEval(newRep bool, iters int, agents int) {
	start := time.Now()
	s.runBenchmarkingSession(newRep, iters)
	end := time.Now()
	elapsed := end.Sub(start)
	elapsed /= time.Duration(iters)
	elapsed /= time.Duration(agents)
	printer := "Linguistic Rep"
	if newRep {
		printer = "Matrix Rep"
	}
	fmt.Printf("%s: %s per rule\n", printer, elapsed)
}
