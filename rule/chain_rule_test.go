package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChainRule(t *testing.T) {
	rule := NewChainRule()
	if rule.ruleType != chainRuleType {
		t.Errorf("Expected ruleType to be %v, got %v", chainRuleType, rule.ruleType)
	}
	if rule.context == nil {
		t.Error("Expected context to be initialized")
	}
	if len(rule.children) != 0 {
		t.Errorf("Expected children to be empty, got %v", len(rule.children))
	}
	if rule.onEval == nil || rule.onPreExecute == nil || rule.onExecute == nil || rule.onPostExecute == nil {
		t.Error("Expected all callbacks to be initialized")
	}
}

func TestChainRuleRunner(t *testing.T) {
	ruleContext := NewRuleContext()
	rule := NewChainRule()

	// Test with no rules
	ChainRuleRunner[ChainRule](ruleContext)
	// No panic expected

	// Test with one rule
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()
	ChainRuleRunner(ruleContext, rule)
	if rule.context != ruleContext {
		t.Error("Expected rule context to be set")
	}

	// Test with more than one rule
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic but did not get one")
		}
	}()
	ChainRuleRunner(ruleContext, rule, rule)
}

func TestChainRuleContext(t *testing.T) {
	rule := NewChainRule()
	rule2 := NewChainRule()

	rule.OnEval(func(ctx Context) bool {
		ctx.GetRuleContext().Set("eval_1", true)
		return ctx.GetRuleContext().Get("start").(bool)

	}).OnPreExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("pre_execute_1", true)

	}).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("execute_1", true)

	}).OnPostExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("post_execute_1", true)

	}).AddChildren(
		rule2.OnEval(func(ctx Context) bool {
			ctx.GetRuleContext().Set("eval_2", true)
			return true

		}).OnPreExecute(func(ctx Context) {
			ctx.GetRuleContext().Set("pre_execute_2", true)

		}).OnExecute(func(ctx Context) {
			ctx.GetRuleContext().Set("execute_2", true)

		}).OnPostExecute(func(ctx Context) {
			ctx.GetRuleContext().Set("post_execute_2", true)
		}))

	ruleContext := NewRuleContext()
	ruleContext.Set("start", true)

	ChainRuleRunner(ruleContext, rule)

	assert.True(t, ruleContext.Get("eval_1").(bool))
	assert.True(t, ruleContext.Get("pre_execute_1").(bool))
	assert.True(t, ruleContext.Get("execute_1").(bool))
	assert.True(t, ruleContext.Get("post_execute_1").(bool))
	assert.True(t, ruleContext.Get("eval_2").(bool))
	assert.True(t, ruleContext.Get("pre_execute_2").(bool))
	assert.True(t, ruleContext.Get("execute_2").(bool))
	assert.True(t, ruleContext.Get("post_execute_2").(bool))
}

func TestChainRule_ShouldNotPanicOnDefaultCallbacks(t *testing.T) {
	rule := NewChainRule()
	rule2 := NewChainRule()

	rule.AddChildren(rule2)

	ChainRuleRunner(NewRuleContext(), rule)
}

func TestChainRule_ShouldRunChildOnEvalTrue(t *testing.T) {
	rule := NewChainRule()
	rule.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_1", true)
	})

	rule2 := NewChainRule()
	rule2.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_2", true)
	})

	rule3 := NewChainRule()
	rule3.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_3", true)
	})

	rule.AddChildren(rule2.AddChildren(rule3))

	ruleContext := NewRuleContext()
	ChainRuleRunner(ruleContext, rule)

	assert.True(t, ruleContext.Get("rule_1").(bool))
	assert.True(t, ruleContext.Get("rule_2").(bool))
	assert.True(t, ruleContext.Get("rule_3").(bool))
}

func TestChainRule_ShouldNotRunChildOnEvalFalse(t *testing.T) {
	rule := NewChainRule()
	rule.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_1", true)
	})

	rule2 := NewChainRule()
	rule2.OnEval(func(ctx Context) bool { return false }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_2", true)
	})

	rule3 := NewChainRule()
	rule3.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_3", true)
	})

	rule.AddChildren(rule2.AddChildren(rule3))

	ruleContext := NewRuleContext()
	ChainRuleRunner(ruleContext, rule)

	assert.True(t, ruleContext.Get("rule_1").(bool))
	assert.PanicsWithError(t, "interface conversion: interface {} is nil, not bool", func() {
		var _ = ruleContext.Get("rule_2").(bool)
	})
	assert.PanicsWithError(t, "interface conversion: interface {} is nil, not bool", func() {
		var _ = ruleContext.Get("rule_3").(bool)
	})
}

func TestChainRule_ShouldPanicWhenProvidingSiblingRuleToRunner(t *testing.T) {
	rule := NewChainRule()
	rule2 := NewChainRule()

	assert.PanicsWithValue(t, "ChainRuleRunner only supports one rule", func() {
		ChainRuleRunner(NewRuleContext(), rule, rule2)
	})
}

func TestChainRule_ShouldPanicWhenProvidingSiblingRulesToRule(t *testing.T) {
	rule := NewChainRule()
	rule2 := NewChainRule()
	rule3 := NewChainRule()

	assert.PanicsWithValue(t, "ChainRule can only have one child", func() {
		rule.AddChildren(rule2, rule3)
	})

	ChainRuleRunner(NewRuleContext(), rule)
}
