package rule

type BestFirstRule struct {
	*BaseRule[BestFirstRule]
}

func NewBestFirstRule() *BaseRule[BestFirstRule] {
	return &BaseRule[BestFirstRule]{
		ruleType:      bestFirstRuleType,
		context:       NewRuleContext(),
		children:      make([]*BaseRule[BestFirstRule], 0),
		onEval:        func(r Context) bool { return true },
		onPreExecute:  func(r Context) {},
		onExecute:     func(r Context) {},
		onPostExecute: func(r Context) {},
	}
}

// BestFirstRuleRunner executes a list of BestFirstRule rules within a given RuleContext.
// It iterates over the provided rules, setting the RuleContext for each rule and firing it.
// If a rule's fire method returns false, the iteration stops.
//
// Parameters:
//   - ruleContext: A pointer to the RuleContext in which the rules will be executed.
//   - rules: A slice of pointers to BestFirstRule objects to be executed.
func BestFirstRuleRunner[T any](ruleContext *RuleContext, rules ...*BaseRule[T]) {
	RuleRunner(bestFirstRuleType, ruleContext, rules...)
}
