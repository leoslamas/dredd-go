package rule

import "context"

// BestFirstRule represents a rule that executes in a best-first manner.
// BestFirstRules can have multiple children and execute the first matching one.
type BestFirstRule[C any] struct {
	*BaseRule[BestFirstRule[C], C]
}

// NewBestFirstRule creates a new BestFirstRule with default settings.
func NewBestFirstRule[C any]() *BestFirstRule[C] {
	baseRule := NewBaseRule[BestFirstRule[C], C](BestFirstRuleType)
	return &BestFirstRule[C]{BaseRule: baseRule}
}

// NewBestFirstRuleWithOptions creates a new BestFirstRule with the given options.
func NewBestFirstRuleWithOptions[C any](options ...RuleOption[BestFirstRule[C], C]) *BestFirstRule[C] {
	baseRule := NewBaseRule[BestFirstRule[C], C](BestFirstRuleType, options...)
	return &BestFirstRule[C]{BaseRule: baseRule}
}

// BestFirstRuleRunnerWithContext executes a list of BestFirstRule rules within a given RuleContext.
// It iterates over the provided rules, setting the RuleContext for each rule and firing it.
// If a rule's fire method returns false, the iteration stops.
//
// Parameters:
//   - goCtx: Go context for cancellation and timeouts.
//   - ruleContext: A pointer to the RuleContext in which the rules will be executed.
//   - rules: A slice of pointers to BestFirstRule objects to be executed.
//
// Returns:
//   - An error if execution fails.
func BestFirstRuleRunnerWithContext[T, C any](goCtx context.Context, ruleContext *RuleContext[C], rules ...*BaseRule[T, C]) error {
	return RuleRunner(BestFirstRuleType, goCtx, ruleContext, rules...)
}

// BestFirstRuleRunner executes best-first rules with a background context.
// This is a convenience method for simple use cases.
func BestFirstRuleRunner[T, C any](ruleContext *RuleContext[C], rules ...*BaseRule[T, C]) error {
	return BestFirstRuleRunnerWithContext(context.Background(), ruleContext, rules...)
}
