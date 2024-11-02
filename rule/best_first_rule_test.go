package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBestFirstRule(t *testing.T) {
	rule := NewBestFirstRule()
	if rule.ruleType != bestFirstRuleType {
		t.Errorf("Expected ruleType to be %v, got %v", bestFirstRuleType, rule.ruleType)
	}
	if rule.context == nil {
		t.Error("Expected context to be initialized, got nil")
	}

	if len(rule.children) != 0 {
		t.Errorf("Expected children to be empty, got %d", len(rule.children))
	}
	if rule.onEval == nil || rule.onPreExecute == nil || rule.onExecute == nil || rule.onPostExecute == nil {
		t.Error("Expected all callbacks to be initialized")
	}
}

func TestBestFirstRuleRunner(t *testing.T) {
	ruleContext := NewRuleContext()
	rule1 := NewBestFirstRule()
	rule2 := NewBestFirstRule()

	rule1.onEval = func(r Context) bool { return true }
	rule2.onEval = func(r Context) bool { return false }

	rules := []*BaseRule[BestFirstRule]{rule1, rule2}

	//test with no rules
	BestFirstRuleRunner[BestFirstRule](ruleContext)

	BestFirstRuleRunner(ruleContext, rules...)

	if rule1.context != ruleContext {
		t.Error("Expected rule1 context to be set to ruleContext")
	}
	if rule2.context == ruleContext {
		t.Error("Expected rule2 context to not be set to ruleContext")
	}
}

func TestBestFirstRuleContext(t *testing.T) {
	rule := NewBestFirstRule()
	rule2 := NewBestFirstRule()

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

	BestFirstRuleRunner(ruleContext, rule)

	assert.True(t, ruleContext.Get("eval_1").(bool))
	assert.True(t, ruleContext.Get("pre_execute_1").(bool))
	assert.True(t, ruleContext.Get("execute_1").(bool))
	assert.True(t, ruleContext.Get("post_execute_1").(bool))
	assert.True(t, ruleContext.Get("eval_2").(bool))
	assert.True(t, ruleContext.Get("pre_execute_2").(bool))
	assert.True(t, ruleContext.Get("execute_2").(bool))
	assert.True(t, ruleContext.Get("post_execute_2").(bool))
}

func TestBestFirstRule_ShouldNotPanicOnDefaultCallbacks(t *testing.T) {
	rule := NewBestFirstRule()
	rule2 := NewBestFirstRule()

	rule.AddChildren(rule2)

	BestFirstRuleRunner(NewRuleContext(), rule)
}

func TestBestFirstRule_ShouldRunChildOnEvalTrue(t *testing.T) {
	rule := NewBestFirstRule()
	rule.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_1", true)
	})

	rule2 := NewBestFirstRule()
	rule2.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_2", true)
	})

	rule3 := NewBestFirstRule()
	rule3.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_3", true)
	})

	rule4 := NewBestFirstRule()
	rule4.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_4", true)
	})

	rule.AddChildren(rule3)
	rule2.AddChildren(rule4)

	ruleContext := NewRuleContext()

	BestFirstRuleRunner(ruleContext, rule, rule2)

	assert.True(t, ruleContext.Get("rule_1").(bool))
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_2").(bool) })
	assert.True(t, ruleContext.Get("rule_3").(bool))
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_4").(bool) })
}

func TestBestFirstRule_ShouldRunSiblingOnEvalFalse(t *testing.T) {
	rule := NewBestFirstRule()
	rule.OnEval(func(ctx Context) bool { return false }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_1", true)
	})

	rule2 := NewBestFirstRule()
	rule2.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_2", true)
	})

	rule3 := NewBestFirstRule()
	rule3.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_3", true)
	})

	rule4 := NewBestFirstRule()
	rule4.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_4", true)
	})

	rule.AddChildren(rule3)
	rule2.AddChildren(rule4)

	ruleContext := NewRuleContext()

	BestFirstRuleRunner(ruleContext, rule, rule2)

	assert.Panics(t, func() { var _ = ruleContext.Get("rule_1").(bool) })
	assert.True(t, ruleContext.Get("rule_2").(bool))
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_3").(bool) })
	assert.True(t, ruleContext.Get("rule_4").(bool))
}

func TestBestFirstRule_ReadmeFlow(t *testing.T) {
	// Rule1      Rule2      Rule3
	//   |
	// Rule4  ->  Rule5      Rule6
	//              |
	// Rule7      Rule8      Rule9
	//              |
	// Rule10     Rule11  ->  Rule12

	rule := NewBestFirstRule()
	rule.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_1", true)
	})

	rule2 := NewBestFirstRule()
	rule2.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_2", true)
	})

	rule3 := NewBestFirstRule()
	rule3.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_3", true)
	})

	rule4 := NewBestFirstRule()
	rule4.OnEval(func(ctx Context) bool { return false }).OnExecute(func(ctx Context) { // FALSE
		ctx.GetRuleContext().Set("rule_4", true)
	})

	rule5 := NewBestFirstRule()
	rule5.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_5", true)
	})

	rule6 := NewBestFirstRule()
	rule6.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_6", true)
	})

	rule7 := NewBestFirstRule()
	rule7.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_7", true)
	})

	rule8 := NewBestFirstRule()
	rule8.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_8", true)
	})

	rule9 := NewBestFirstRule()
	rule9.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_9", true)
	})

	rule10 := NewBestFirstRule()
	rule10.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_10", true)
	})

	rule11 := NewBestFirstRule()
	rule11.OnEval(func(ctx Context) bool { return false }).OnExecute(func(ctx Context) { // FALSE
		ctx.GetRuleContext().Set("rule_11", true)
	})

	rule12 := NewBestFirstRule()
	rule12.OnEval(func(ctx Context) bool { return true }).OnExecute(func(ctx Context) {
		ctx.GetRuleContext().Set("rule_12", true)
	})

	rule.AddChildren(rule4, rule5, rule6)
	rule4.AddChildren(rule7)
	rule5.AddChildren(rule8, rule9)
	rule7.AddChildren(rule10)
	rule8.AddChildren(rule11, rule12)

	ruleContext := NewRuleContext()

	BestFirstRuleRunner(ruleContext, rule, rule2, rule3)

	assert.True(t, ruleContext.Get("rule_1").(bool))
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_2").(bool) })
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_3").(bool) })
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_4").(bool) })
	assert.True(t, ruleContext.Get("rule_5").(bool))
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_6").(bool) })
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_7").(bool) })
	assert.True(t, ruleContext.Get("rule_8").(bool))
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_9").(bool) })
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_10").(bool) })
	assert.Panics(t, func() { var _ = ruleContext.Get("rule_11").(bool) })
	assert.True(t, ruleContext.Get("rule_12").(bool))
}
