package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChainRule(t *testing.T) {
	rule := NewChainRule[bool]()
	if rule.ruleType != ChainRuleType {
		t.Errorf("Expected ruleType to be %v, got %v", ChainRuleType, rule.ruleType)
	}
	if rule.context == nil {
		t.Error("Expected context to be initialized")
	}
	if len(rule.children) != 0 {
		t.Errorf("Expected children to be empty, got %v", len(rule.children))
	}
	// Note: callbacks can be nil in new API
}

func TestChainRuleRunner(t *testing.T) {
	ruleContext := NewRuleContext[bool]()
	rule := NewChainRule[bool]()

	// Test with no rules
	err := ChainRuleRunner[ChainRule[bool], bool](ruleContext)
	assert.NoError(t, err)

	// Test with one rule
	err = ChainRuleRunner(ruleContext, rule.BaseRule)
	assert.NoError(t, err)
	if rule.context != ruleContext {
		t.Error("Expected rule context to be set")
	}

	// Test with more than one rule - should return error instead of panic
	rule2 := NewChainRule[bool]()
	err = ChainRuleRunner(ruleContext, rule.BaseRule, rule2.BaseRule)
	assert.Error(t, err)
	assert.Equal(t, ErrChainRuleMultipleRules, err)
}

func TestChainRuleContext(t *testing.T) {
	rule := NewChainRule[bool]()
	rule2 := NewChainRule[bool]()

	rule.OnEval(func(ctx Context[bool]) bool {
		ctx.GetRuleContext().Set("eval_1", true)
		start, _ := ctx.GetRuleContext().Get("start")
		return start

	}).OnPreExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("pre_execute_1", true)

	}).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("execute_1", true)

	}).OnPostExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("post_execute_1", true)
	})

	rule2.OnEval(func(ctx Context[bool]) bool {
		ctx.GetRuleContext().Set("eval_2", true)
		return true

	}).OnPreExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("pre_execute_2", true)

	}).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("execute_2", true)

	}).OnPostExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("post_execute_2", true)
	})

	_ = rule.AddChildren(rule2.BaseRule)

	ruleContext := NewRuleContext[bool]()
	ruleContext.Set("start", true)

	err := ChainRuleRunner(ruleContext, rule.BaseRule)
	assert.NoError(t, err)

	val1, _ := ruleContext.Get("eval_1")
	assert.True(t, val1)
	val2, _ := ruleContext.Get("pre_execute_1")
	assert.True(t, val2)
	val3, _ := ruleContext.Get("execute_1")
	assert.True(t, val3)
	val4, _ := ruleContext.Get("post_execute_1")
	assert.True(t, val4)
	val5, _ := ruleContext.Get("eval_2")
	assert.True(t, val5)
	val6, _ := ruleContext.Get("pre_execute_2")
	assert.True(t, val6)
	val7, _ := ruleContext.Get("execute_2")
	assert.True(t, val7)
	val8, _ := ruleContext.Get("post_execute_2")
	assert.True(t, val8)
}

func TestChainRule_ShouldNotPanicOnDefaultCallbacks(t *testing.T) {
	rule := NewChainRule[bool]()
	rule2 := NewChainRule[bool]()

	_ = rule.AddChildren(rule2.BaseRule)

	err := ChainRuleRunner(NewRuleContext[bool](), rule.BaseRule)
	assert.NoError(t, err)
}

func TestChainRule_ShouldRunChildOnEvalTrue(t *testing.T) {
	rule := NewChainRule[bool]()
	rule.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_1", true)
	})

	rule2 := NewChainRule[bool]()
	rule2.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_2", true)
	})

	rule3 := NewChainRule[bool]()
	rule3.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_3", true)
	})

	_ = rule2.AddChildren(rule3.BaseRule)
	_ = rule.AddChildren(rule2.BaseRule)

	ruleContext := NewRuleContext[bool]()
	err := ChainRuleRunner(ruleContext, rule.BaseRule)
	assert.NoError(t, err)

	val1, _ := ruleContext.Get("rule_1")
	assert.True(t, val1)
	val2, _ := ruleContext.Get("rule_2")
	assert.True(t, val2)
	val3, _ := ruleContext.Get("rule_3")
	assert.True(t, val3)
}

func TestChainRule_ShouldNotRunChildOnEvalFalse(t *testing.T) {
	rule := NewChainRule[bool]()
	rule.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_1", true)
	})

	rule2 := NewChainRule[bool]()
	rule2.OnEval(func(ctx Context[bool]) bool { return false }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_2", true)
	})

	rule3 := NewChainRule[bool]()
	rule3.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_3", true)
	})

	_ = rule2.AddChildren(rule3.BaseRule)
	_ = rule.AddChildren(rule2.BaseRule)

	ruleContext := NewRuleContext[bool]()
	err := ChainRuleRunner(ruleContext, rule.BaseRule)
	assert.NoError(t, err)

	val1, _ := ruleContext.Get("rule_1")
	assert.True(t, val1)
	_, exists2 := ruleContext.Get("rule_2")
	assert.False(t, exists2) // rule_2 should not exist because eval returned false
	_, exists3 := ruleContext.Get("rule_3")
	assert.False(t, exists3) // rule_3 should not exist because rule_2 didn't run
}

func TestChainRule_ShouldErrorWhenProvidingSiblingRuleToRunner(t *testing.T) {
	rule := NewChainRule[bool]()
	rule2 := NewChainRule[bool]()

	err := ChainRuleRunner(NewRuleContext[bool](), rule.BaseRule, rule2.BaseRule)
	assert.Error(t, err)
	assert.Equal(t, ErrChainRuleMultipleRules, err)
}

func TestChainRule_ShouldErrorWhenProvidingSiblingRulesToRule(t *testing.T) {
	rule := NewChainRule[bool]()
	rule2 := NewChainRule[bool]()
	rule3 := NewChainRule[bool]()

	err := rule.AddChildren(rule2.BaseRule, rule3.BaseRule)
	assert.Error(t, err)
	assert.Equal(t, ErrChainRuleMultipleChildren, err)

	err = ChainRuleRunner(NewRuleContext[bool](), rule.BaseRule)
	assert.NoError(t, err)
}

func TestChainRule_SingleChildConstraint(t *testing.T) {
	rule := NewChainRule[int]()
	child1 := NewChainRule[int]()
	child2 := NewChainRule[int]()

	// Adding one child should work
	err := rule.AddChildren(child1.BaseRule)
	assert.NoError(t, err)

	// Adding a second child should fail
	err = rule.AddChildren(child2.BaseRule)
	assert.Error(t, err)
	assert.Equal(t, ErrChainRuleMultipleChildren, err)
}

func TestChainRule_NilChildValidation(t *testing.T) {
	rule := NewChainRule[int]()

	// Adding nil child should fail
	err := rule.AddChildren(nil)
	assert.Error(t, err)
	assert.Equal(t, ErrNilRule, err)
}
