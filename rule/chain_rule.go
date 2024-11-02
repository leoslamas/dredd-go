package rule

type ChainRule struct {
	*BaseRule[ChainRule]
}

func NewChainRule() *BaseRule[ChainRule] {
	return &BaseRule[ChainRule]{
		ruleType:      chainRuleType,
		context:       NewRuleContext(),
		children:      make([]*BaseRule[ChainRule], 0),
		onEval:        func(r Context) bool { return true },
		onPreExecute:  func(r Context) {},
		onExecute:     func(r Context) {},
		onPostExecute: func(r Context) {},
	}
}

// ChainRuleRunner executes a chain of rules within a given rule context.
// It expects a non-empty slice of ChainRule pointers, but will panic if more than one rule is provided.
// The function sets the rule context for the provided rule and then fires the rule.
//
// Parameters:
//   - ruleContext: A pointer to the RuleContext in which the rules will be executed.
//   - rules: A slice of pointers to ChainRule. Must contain exactly one rule.
//
// Panics:
//   - If the length of the rules slice is greater than one.
func ChainRuleRunner[T any](ruleContext *RuleContext, rules ...*BaseRule[T]) {
	RuleRunner(chainRuleType, ruleContext, rules...)
}
