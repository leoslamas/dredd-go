package rule

import "context"

// ChainRule represents a rule that executes in a linear chain.
// Each ChainRule can have at most one child rule.
type ChainRule[C any] struct {
	*BaseRule[ChainRule[C], C]
}

// NewChainRule creates a new ChainRule with default settings.
func NewChainRule[C any]() *ChainRule[C] {
	baseRule := NewBaseRule[ChainRule[C], C](ChainRuleType)
	return &ChainRule[C]{BaseRule: baseRule}
}

// NewChainRuleWithOptions creates a new ChainRule with the given options.
func NewChainRuleWithOptions[C any](options ...RuleOption[ChainRule[C], C]) *ChainRule[C] {
	baseRule := NewBaseRule[ChainRule[C], C](ChainRuleType, options...)
	return &ChainRule[C]{BaseRule: baseRule}
}

// ChainRuleRunnerWithContext executes a chain of rules within a given rule context.
// It expects exactly one rule and returns an error if constraints are violated.
//
// Parameters:
//   - goCtx: Go context for cancellation and timeouts.
//   - ruleContext: A pointer to the RuleContext in which the rules will be executed.
//   - rules: A slice of pointers to ChainRule. Must contain exactly one rule.
//
// Returns:
//   - An error if execution fails or constraints are violated.
func ChainRuleRunnerWithContext[T, C any](goCtx context.Context, ruleContext *RuleContext[C], rules ...*BaseRule[T, C]) error {
	return RuleRunner(ChainRuleType, goCtx, ruleContext, rules...)
}

// ChainRuleRunner executes a chain rule with a background context.
// This is a convenience method for simple use cases.
func ChainRuleRunner[T, C any](ruleContext *RuleContext[C], rules ...*BaseRule[T, C]) error {
	return ChainRuleRunnerWithContext(context.Background(), ruleContext, rules...)
}
