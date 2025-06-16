package rule

import (
	"context"
	"testing"
	"time"
)

func BenchmarkRuleContext_Set(b *testing.B) {
	ctx := NewRuleContext[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx.Set("key", i)
	}
}

func BenchmarkRuleContext_Get(b *testing.B) {
	ctx := NewRuleContext[int]()
	ctx.Set("key", 42)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ctx.Get("key")
	}
}

func BenchmarkRuleContext_SetConcurrent(b *testing.B) {
	ctx := NewRuleContext[int]()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ctx.Set("key", i)
			i++
		}
	})
}

func BenchmarkRuleContext_GetConcurrent(b *testing.B) {
	ctx := NewRuleContext[int]()
	ctx.Set("key", 42)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = ctx.Get("key")
		}
	})
}

func BenchmarkChainRule_SimpleExecution(b *testing.B) {
	rule := NewChainRule[int]()
	rule.OnEval(func(ctx Context[int]) bool {
		return true
	}).OnExecute(func(ctx Context[int]) {
		ctx.GetRuleContext().Set("executed", 1)
	})

	ruleContext := NewRuleContext[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ChainRuleRunner(ruleContext, rule.BaseRule)
	}
}

func BenchmarkBestFirstRule_SimpleExecution(b *testing.B) {
	rule := NewBestFirstRule[int]()
	rule.OnEval(func(ctx Context[int]) bool {
		return true
	}).OnExecute(func(ctx Context[int]) {
		ctx.GetRuleContext().Set("executed", 1)
	})

	ruleContext := NewRuleContext[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = BestFirstRuleRunner(ruleContext, rule.BaseRule)
	}
}

func BenchmarkChainRule_DeepNesting(b *testing.B) {
	// Create a chain of 10 rules
	const depth = 10
	rules := make([]*ChainRule[int], depth)

	for i := 0; i < depth; i++ {
		rules[i] = NewChainRule[int]()
		rules[i].OnEval(func(ctx Context[int]) bool {
			return true
		}).OnExecute(func(ctx Context[int]) {
			count, _ := ctx.GetRuleContext().Get("count")
			ctx.GetRuleContext().Set("count", count+1)
		})

		if i > 0 {
			_ = rules[i-1].AddChildren(rules[i].BaseRule)
		}
	}

	ruleContext := NewRuleContext[int]()
	ruleContext.Set("count", 0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ruleContext.Set("count", 0)
		_ = ChainRuleRunner(ruleContext, rules[0].BaseRule)
	}
}

func BenchmarkBestFirstRule_WideTree(b *testing.B) {
	// Create a tree with 10 siblings
	const width = 10
	rules := make([]*BestFirstRule[int], width)

	for i := 0; i < width; i++ {
		rules[i] = NewBestFirstRule[int]()
		idx := i // Capture loop variable
		rules[i].OnEval(func(ctx Context[int]) bool {
			// Only the last rule should match
			return idx == width-1
		}).OnExecute(func(ctx Context[int]) {
			ctx.GetRuleContext().Set("executed", idx)
		})
	}

	// Convert to BaseRule slice for runner
	baseRules := make([]*BaseRule[BestFirstRule[int], int], width)
	for i, rule := range rules {
		baseRules[i] = rule.BaseRule
	}

	ruleContext := NewRuleContext[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = BestFirstRuleRunner(ruleContext, baseRules...)
	}
}

func BenchmarkFunctionalOptions_Creation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewBaseRule[string, int](
			ChainRuleType,
			WithEvaluation[string, int](func(ctx Context[int]) bool {
				return true
			}),
			WithExecution[string, int](func(ctx Context[int]) {
				ctx.GetRuleContext().Set("result", 42)
			}),
		)
	}
}

func BenchmarkRuleContext_Capacity(b *testing.B) {
	const capacity = 1000

	b.Run("WithoutCapacity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := NewRuleContext[int]()
			for j := 0; j < capacity; j++ {
				ctx.Set(string(rune('a'+j%26)), j)
			}
		}
	})

	b.Run("WithCapacity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := NewRuleContextWithCapacity[int](capacity)
			for j := 0; j < capacity; j++ {
				ctx.Set(string(rune('a'+j%26)), j)
			}
		}
	})
}

func BenchmarkRuleExecution_WithContext(b *testing.B) {
	rule := NewChainRule[int]()
	rule.OnEval(func(ctx Context[int]) bool {
		return true
	}).OnExecute(func(ctx Context[int]) {
		ctx.GetRuleContext().Set("executed", 1)
	})

	ruleContext := NewRuleContext[int]()
	goCtx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ChainRuleRunnerWithContext(goCtx, ruleContext, rule.BaseRule)
	}
}

func BenchmarkRuleExecution_WithTimeout(b *testing.B) {
	rule := NewChainRule[int]()
	rule.OnEval(func(ctx Context[int]) bool {
		return true
	}).OnExecute(func(ctx Context[int]) {
		ctx.GetRuleContext().Set("executed", 1)
	})

	ruleContext := NewRuleContext[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		_ = ChainRuleRunnerWithContext(ctx, ruleContext, rule.BaseRule)
		cancel()
	}
}
