package main

import (
	"fmt"
	"log"

	"github.com/leoslamas/dredd-go/rule"
)

func main() {
	// Example 1: Using the new API with proper error handling
	fmt.Println("=== Example 1: Chain Rule with New API ===")
	runChainRuleExample()

	fmt.Println("\n=== Example 2: Best First Rule Example ===")
	runBestFirstRuleExample()

	fmt.Println("\n=== Example 3: Functional Options Pattern ===")
	runFunctionalOptionsExample()
}

func runChainRuleExample() {
	// Create rules using the new typed API
	rule1 := rule.NewChainRule[bool]()
	rule2 := rule.NewChainRule[bool]()

	// Configure rule1 with the new API
	rule1.OnEval(func(ctx rule.Context[bool]) bool {
		fmt.Println("Eval Chain Rule 1")
		value, exists := ctx.GetRuleContext().Get("value")
		if !exists {
			fmt.Println("Warning: 'value' key not found in context")
			return false
		}
		return value
	}).OnPreExecute(func(ctx rule.Context[bool]) {
		fmt.Println("Pre Chain Rule 1")
	}).OnExecute(func(ctx rule.Context[bool]) {
		fmt.Println("Execute Chain Rule 1")
	}).OnPostExecute(func(ctx rule.Context[bool]) {
		fmt.Println("Post Chain Rule 1")
	})

	// Configure rule2
	rule2.OnEval(func(ctx rule.Context[bool]) bool {
		fmt.Println("Eval Chain Rule 2")
		return false
	}).OnExecute(func(ctx rule.Context[bool]) {
		fmt.Println("Execute Chain Rule 2") // unreachable
	})

	// Add child rule with error handling
	if err := rule1.AddChildren(rule2.BaseRule); err != nil {
		log.Fatalf("Failed to add child rule: %v", err)
	}

	// Create context and execute
	ruleContext := rule.NewRuleContext[bool]()
	ruleContext.Set("value", true)

	if err := rule.ChainRuleRunner(ruleContext, rule1.BaseRule); err != nil {
		log.Fatalf("Chain rule execution failed: %v", err)
	}
}

func runBestFirstRuleExample() {
	// Create best-first rules
	rule1 := rule.NewBestFirstRule[string]()
	rule2 := rule.NewBestFirstRule[string]()
	rule3 := rule.NewBestFirstRule[string]()

	// Configure rules
	rule1.OnEval(func(ctx rule.Context[string]) bool {
		fmt.Println("Evaluating Rule 1")
		return false // This will cause rule2 to be evaluated
	}).OnExecute(func(ctx rule.Context[string]) {
		fmt.Println("Executing Rule 1")
		ctx.GetRuleContext().Set("executed", "rule1")
	})

	rule2.OnEval(func(ctx rule.Context[string]) bool {
		fmt.Println("Evaluating Rule 2")
		return true // This will execute and stop further evaluation
	}).OnExecute(func(ctx rule.Context[string]) {
		fmt.Println("Executing Rule 2")
		ctx.GetRuleContext().Set("executed", "rule2")
	})

	rule3.OnEval(func(ctx rule.Context[string]) bool {
		fmt.Println("Evaluating Rule 3")
		return true
	}).OnExecute(func(ctx rule.Context[string]) {
		fmt.Println("Executing Rule 3")
		ctx.GetRuleContext().Set("executed", "rule3")
	})

	// Execute best-first rules
	ruleContext := rule.NewRuleContext[string]()
	if err := rule.BestFirstRuleRunner(ruleContext, rule1.BaseRule, rule2.BaseRule, rule3.BaseRule); err != nil {
		log.Fatalf("Best first rule execution failed: %v", err)
	}

	// Check which rule was executed
	if executed, exists := ruleContext.Get("executed"); exists {
		fmt.Printf("Final result: %s was executed\n", executed)
	}
}

func runFunctionalOptionsExample() {
	// Create a rule using functional options
	rule1 := rule.NewChainRuleWithOptions[int](
		rule.WithEvaluation[rule.ChainRule[int], int](func(ctx rule.Context[int]) bool {
			fmt.Println("Functional Options: Evaluating")
			count, exists := ctx.GetRuleContext().Get("count")
			if !exists {
				return false
			}
			return count > 0
		}),
		rule.WithExecution[rule.ChainRule[int], int](func(ctx rule.Context[int]) {
			fmt.Println("Functional Options: Executing")
			count, _ := ctx.GetRuleContext().Get("count")
			ctx.GetRuleContext().Set("result", count*2)
		}),
	)

	// Execute the rule
	ruleContext := rule.NewRuleContext[int]()
	ruleContext.Set("count", 5)

	if err := rule.ChainRuleRunner(ruleContext, rule1.BaseRule); err != nil {
		log.Fatalf("Functional options rule execution failed: %v", err)
	}

	if result, exists := ruleContext.Get("result"); exists {
		fmt.Printf("Functional Options: Result = %d\n", result)
	}
}
