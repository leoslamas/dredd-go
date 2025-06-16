package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBestFirstRule(t *testing.T) {
	rule := NewBestFirstRule[bool]()
	if rule.ruleType != BestFirstRuleType {
		t.Errorf("Expected ruleType to be %v, got %v", BestFirstRuleType, rule.ruleType)
	}
	if rule.context == nil {
		t.Error("Expected context to be initialized, got nil")
	}

	if len(rule.children) != 0 {
		t.Errorf("Expected children to be empty, got %d", len(rule.children))
	}
	// Note: callbacks can be nil in new API
}

func TestBestFirstRuleRunner(t *testing.T) {
	ruleContext := NewRuleContext[bool]()
	rule1 := NewBestFirstRule[bool]()
	rule2 := NewBestFirstRule[bool]()

	rule1.OnEval(func(r Context[bool]) bool { return true })
	rule2.OnEval(func(r Context[bool]) bool { return false })

	//test with no rules
	err := BestFirstRuleRunner[BestFirstRule[bool], bool](ruleContext)
	assert.NoError(t, err)

	err = BestFirstRuleRunner(ruleContext, rule1.BaseRule, rule2.BaseRule)
	assert.NoError(t, err)

	if rule1.context != ruleContext {
		t.Error("Expected rule1 context to be set to ruleContext")
	}
	if rule2.context == ruleContext {
		t.Error("Expected rule2 context to not be set to ruleContext")
	}
}

func TestBestFirstRuleContext(t *testing.T) {
	rule := NewBestFirstRule[bool]()
	rule2 := NewBestFirstRule[bool]()

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

	err := BestFirstRuleRunner(ruleContext, rule.BaseRule)
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

func TestBestFirstRule_ShouldNotPanicOnDefaultCallbacks(t *testing.T) {
	rule := NewBestFirstRule[bool]()
	rule2 := NewBestFirstRule[bool]()

	_ = rule.AddChildren(rule2.BaseRule)

	err := BestFirstRuleRunner(NewRuleContext[bool](), rule.BaseRule)
	assert.NoError(t, err)
}

func TestBestFirstRule_ShouldRunChildOnEvalTrue(t *testing.T) {
	rule := NewBestFirstRule[bool]()
	rule.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_1", true)
	})

	rule2 := NewBestFirstRule[bool]()
	rule2.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_2", true)
	})

	rule3 := NewBestFirstRule[bool]()
	rule3.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_3", true)
	})

	rule4 := NewBestFirstRule[bool]()
	rule4.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_4", true)
	})

	_ = rule.AddChildren(rule3.BaseRule)
	_ = rule2.AddChildren(rule4.BaseRule)

	ruleContext := NewRuleContext[bool]()

	err := BestFirstRuleRunner(ruleContext, rule.BaseRule, rule2.BaseRule)
	assert.NoError(t, err)

	val1, _ := ruleContext.Get("rule_1")
	assert.True(t, val1)
	_, exists2 := ruleContext.Get("rule_2")
	assert.False(t, exists2) // rule2 should not execute because rule1 executed and stopped iteration
	val3, _ := ruleContext.Get("rule_3")
	assert.True(t, val3)
	_, exists4 := ruleContext.Get("rule_4")
	assert.False(t, exists4) // rule4 should not execute
}

func TestBestFirstRule_ShouldRunSiblingOnEvalFalse(t *testing.T) {
	rule := NewBestFirstRule[bool]()
	rule.OnEval(func(ctx Context[bool]) bool { return false }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_1", true)
	})

	rule2 := NewBestFirstRule[bool]()
	rule2.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_2", true)
	})

	rule3 := NewBestFirstRule[bool]()
	rule3.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_3", true)
	})

	rule4 := NewBestFirstRule[bool]()
	rule4.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_4", true)
	})

	_ = rule.AddChildren(rule3.BaseRule)
	_ = rule2.AddChildren(rule4.BaseRule)

	ruleContext := NewRuleContext[bool]()

	err := BestFirstRuleRunner(ruleContext, rule.BaseRule, rule2.BaseRule)
	assert.NoError(t, err)

	_, exists1 := ruleContext.Get("rule_1")
	assert.False(t, exists1) // rule1 should not execute because eval returned false
	val2, _ := ruleContext.Get("rule_2")
	assert.True(t, val2)
	_, exists3 := ruleContext.Get("rule_3")
	assert.False(t, exists3) // rule3 should not execute because rule1 didn't execute
	val4, _ := ruleContext.Get("rule_4")
	assert.True(t, val4)
}

func TestBestFirstRule_ReadmeFlow(t *testing.T) {
	// Rule1      Rule2      Rule3
	//   |
	// Rule4  ->  Rule5      Rule6
	//              |
	// Rule7      Rule8      Rule9
	//              |
	// Rule10     Rule11  ->  Rule12

	rule := NewBestFirstRule[bool]()
	rule.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_1", true)
	})

	rule2 := NewBestFirstRule[bool]()
	rule2.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_2", true)
	})

	rule3 := NewBestFirstRule[bool]()
	rule3.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_3", true)
	})

	rule4 := NewBestFirstRule[bool]()
	rule4.OnEval(func(ctx Context[bool]) bool { return false }).OnExecute(func(ctx Context[bool]) { // FALSE
		ctx.GetRuleContext().Set("rule_4", true)
	})

	rule5 := NewBestFirstRule[bool]()
	rule5.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_5", true)
	})

	rule6 := NewBestFirstRule[bool]()
	rule6.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_6", true)
	})

	rule7 := NewBestFirstRule[bool]()
	rule7.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_7", true)
	})

	rule8 := NewBestFirstRule[bool]()
	rule8.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_8", true)
	})

	rule9 := NewBestFirstRule[bool]()
	rule9.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_9", true)
	})

	rule10 := NewBestFirstRule[bool]()
	rule10.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_10", true)
	})

	rule11 := NewBestFirstRule[bool]()
	rule11.OnEval(func(ctx Context[bool]) bool { return false }).OnExecute(func(ctx Context[bool]) { // FALSE
		ctx.GetRuleContext().Set("rule_11", true)
	})

	rule12 := NewBestFirstRule[bool]()
	rule12.OnEval(func(ctx Context[bool]) bool { return true }).OnExecute(func(ctx Context[bool]) {
		ctx.GetRuleContext().Set("rule_12", true)
	})

	_ = rule.AddChildren(rule4.BaseRule, rule5.BaseRule, rule6.BaseRule)
	_ = rule4.AddChildren(rule7.BaseRule)
	_ = rule5.AddChildren(rule8.BaseRule, rule9.BaseRule)
	_ = rule7.AddChildren(rule10.BaseRule)
	_ = rule8.AddChildren(rule11.BaseRule, rule12.BaseRule)

	ruleContext := NewRuleContext[bool]()

	err := BestFirstRuleRunner(ruleContext, rule.BaseRule, rule2.BaseRule, rule3.BaseRule)
	assert.NoError(t, err)

	val1, _ := ruleContext.Get("rule_1")
	assert.True(t, val1)
	_, exists2 := ruleContext.Get("rule_2")
	assert.False(t, exists2)
	_, exists3 := ruleContext.Get("rule_3")
	assert.False(t, exists3)
	_, exists4 := ruleContext.Get("rule_4")
	assert.False(t, exists4)
	val5, _ := ruleContext.Get("rule_5")
	assert.True(t, val5)
	_, exists6 := ruleContext.Get("rule_6")
	assert.False(t, exists6)
	_, exists7 := ruleContext.Get("rule_7")
	assert.False(t, exists7)
	val8, _ := ruleContext.Get("rule_8")
	assert.True(t, val8)
	_, exists9 := ruleContext.Get("rule_9")
	assert.False(t, exists9)
	_, exists10 := ruleContext.Get("rule_10")
	assert.False(t, exists10)
	_, exists11 := ruleContext.Get("rule_11")
	assert.False(t, exists11)
	val12, _ := ruleContext.Get("rule_12")
	assert.True(t, val12)
}

func TestBestFirstRule_MultipleChildren(t *testing.T) {
	rule := NewBestFirstRule[int]()
	child1 := NewBestFirstRule[int]()
	child2 := NewBestFirstRule[int]()
	child3 := NewBestFirstRule[int]()

	// Adding multiple children should work for BestFirstRule
	err := rule.AddChildren(child1.BaseRule, child2.BaseRule, child3.BaseRule)
	assert.NoError(t, err)
	assert.Equal(t, 3, rule.ChildrenCount())
	assert.True(t, rule.HasChildren())
}
