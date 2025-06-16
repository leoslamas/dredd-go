package rule

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuleContext(t *testing.T) {
	rc := NewRuleContext[string]()
	assert.NotNil(t, rc)
	_, exists := rc.Get("nonexistent")
	assert.False(t, exists)
}

func TestRuleContext_SetAndGet(t *testing.T) {
	rc := NewRuleContext[string]()
	rc.Set("key", "value")
	value, exists := rc.Get("key")
	assert.True(t, exists)
	assert.Equal(t, "value", value)
}

func TestRuleContext_ThreadSafety(t *testing.T) {
	ctx := NewRuleContext[int]()

	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Test concurrent reads and writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				ctx.Set(key, id*j)

				if value, exists := ctx.Get(key); exists {
					assert.Equal(t, id*j, value)
				}

				if j%10 == 0 {
					ctx.Delete(key)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	assert.True(t, ctx.Size() > 0)
	keys := ctx.Keys()
	assert.True(t, len(keys) > 0)
}

func TestRuleContext_TypeSafety(t *testing.T) {
	// Test string context
	strCtx := NewRuleContext[string]()
	strCtx.Set("key", "value")
	value, exists := strCtx.Get("key")
	assert.True(t, exists)
	assert.Equal(t, "value", value)

	// Test int context
	intCtx := NewRuleContext[int]()
	intCtx.Set("key", 42)
	intValue, exists := intCtx.Get("key")
	assert.True(t, exists)
	assert.Equal(t, 42, intValue)

	// Test struct context
	type testStruct struct {
		Name string
		Age  int
	}
	structCtx := NewRuleContext[testStruct]()
	expected := testStruct{Name: "test", Age: 25}
	structCtx.Set("key", expected)
	structValue, exists := structCtx.Get("key")
	assert.True(t, exists)
	assert.Equal(t, expected, structValue)
}

func TestRuleContext_Operations(t *testing.T) {
	ctx := NewRuleContext[string]()

	// Test Set and Get
	ctx.Set("key1", "value1")
	value, exists := ctx.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Test non-existent key
	_, exists = ctx.Get("nonexistent")
	assert.False(t, exists)

	// Test MustGet with existing key
	assert.Equal(t, "value1", ctx.MustGet("key1"))

	// Test MustGet with non-existent key (should panic)
	assert.Panics(t, func() {
		ctx.MustGet("nonexistent")
	})

	// Test Exists
	assert.True(t, ctx.Exists("key1"))
	assert.False(t, ctx.Exists("nonexistent"))

	// Test Delete
	ctx.Delete("key1")
	assert.False(t, ctx.Exists("key1"))

	// Test Keys and Size
	ctx.Set("key1", "value1")
	ctx.Set("key2", "value2")
	assert.Equal(t, 2, ctx.Size())

	keys := ctx.Keys()
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
}

func TestNewBaseRule_ErrorHandling(t *testing.T) {
	rule := NewBaseRule[string, int](ChainRuleType)

	// Test error propagation in evaluation
	rule.OnEvalWithError(func(ctx Context[int]) EvaluationResult {
		return EvaluationResult{
			ShouldExecute: false,
			Error:         errors.New("evaluation error"),
		}
	})

	ctx := NewRuleContext[int]()
	rule.SetRuleContext(ctx)
	rule.SetGoContext(context.Background())

	_, err := rule.fire()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "evaluation error")
}

func TestNewBaseRule_ContextCancellation(t *testing.T) {
	rule := NewBaseRule[string, int](ChainRuleType)

	rule.OnEval(func(ctx Context[int]) bool {
		return true
	})
	rule.OnExecute(func(ctx Context[int]) {
		// This should not execute due to cancellation
		t.Error("Should not execute due to cancellation")
	})

	// Create a cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	ctx := NewRuleContext[int]()
	rule.SetRuleContext(ctx)
	rule.SetGoContext(cancelledCtx)

	_, err := rule.fire()
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestNewBaseRule_NilSafety(t *testing.T) {
	rule := NewBaseRule[string, int](ChainRuleType)

	// Test with nil callbacks (should not panic)
	ctx := NewRuleContext[int]()
	rule.SetRuleContext(ctx)
	rule.SetGoContext(context.Background())

	_, err := rule.fire()
	assert.NoError(t, err)
}

func TestRuleRunner_Validation(t *testing.T) {
	// Test with nil context
	rule := NewChainRule[int]()
	err := ChainRuleRunner[ChainRule[int], int](nil, rule.BaseRule)
	assert.Error(t, err)
	assert.Equal(t, ErrNilRuleContext, err)

	// Test with nil rule
	ctx := NewRuleContext[int]()
	err = ChainRuleRunner[ChainRule[int], int](ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule at index 0 is nil")

	// Test chain rule with multiple rules
	rule1 := NewChainRule[int]()
	rule2 := NewChainRule[int]()
	err = ChainRuleRunner(ctx, rule1.BaseRule, rule2.BaseRule)
	assert.Error(t, err)
	assert.Equal(t, ErrChainRuleMultipleRules, err)
}

func TestFunctionalOptions(t *testing.T) {
	executed := false

	rule := NewBaseRule[string, int](
		ChainRuleType,
		WithEvaluation[string, int](func(ctx Context[int]) bool {
			return true
		}),
		WithExecution[string, int](func(ctx Context[int]) {
			executed = true
			ctx.GetRuleContext().Set("result", 42)
		}),
	)

	ctx := NewRuleContext[int]()
	err := RuleRunner(ChainRuleType, context.Background(), ctx, rule)
	assert.NoError(t, err)
	assert.True(t, executed)

	result, exists := ctx.Get("result")
	assert.True(t, exists)
	assert.Equal(t, 42, result)
}

func TestRuleString(t *testing.T) {
	rule := NewBaseRule[string, int](ChainRuleType)
	str := rule.String()
	assert.Contains(t, str, "ChainRule")
	assert.Contains(t, str, "children: 0")

	child := NewBaseRule[string, int](BestFirstRuleType)
	_ = rule.AddChildren(child)
	str = rule.String()
	assert.Contains(t, str, "children: 1")
}

func TestRuleType_String(t *testing.T) {
	assert.Equal(t, "ChainRule", ChainRuleType.String())
	assert.Equal(t, "BestFirstRule", BestFirstRuleType.String())
	assert.Equal(t, "UnknownRuleType", RuleType(999).String())
}

func TestComplexRuleExecution(t *testing.T) {
	// Create a complex rule tree - all best first rules for consistency
	rootRule := NewBestFirstRule[string]()
	child1 := NewBestFirstRule[string]()
	child2 := NewBestFirstRule[string]()
	grandchild := NewBestFirstRule[string]()

	executionOrder := make([]string, 0)

	rootRule.OnEval(func(ctx Context[string]) bool {
		executionOrder = append(executionOrder, "root_eval")
		return true
	}).OnExecute(func(ctx Context[string]) {
		executionOrder = append(executionOrder, "root_execute")
	})

	child1.OnEval(func(ctx Context[string]) bool {
		executionOrder = append(executionOrder, "child1_eval")
		return true
	}).OnExecute(func(ctx Context[string]) {
		executionOrder = append(executionOrder, "child1_execute")
	})

	child2.OnEval(func(ctx Context[string]) bool {
		executionOrder = append(executionOrder, "child2_eval")
		return false // Should not execute
	}).OnExecute(func(ctx Context[string]) {
		executionOrder = append(executionOrder, "child2_execute")
	})

	grandchild.OnEval(func(ctx Context[string]) bool {
		executionOrder = append(executionOrder, "grandchild_eval")
		return true
	}).OnExecute(func(ctx Context[string]) {
		executionOrder = append(executionOrder, "grandchild_execute")
	})

	// Build the tree
	require.NoError(t, child1.AddChildren(grandchild.BaseRule))
	require.NoError(t, rootRule.AddChildren(child1.BaseRule, child2.BaseRule))

	// Execute
	ctx := NewRuleContext[string]()
	err := BestFirstRuleRunner(ctx, rootRule.BaseRule)
	assert.NoError(t, err)

	expected := []string{
		"root_eval",
		"root_execute",
		"child1_eval",
		"child1_execute",
		"grandchild_eval",
		"grandchild_execute",
	}
	assert.Equal(t, expected, executionOrder)
}
