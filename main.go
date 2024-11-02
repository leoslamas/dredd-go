package main

import "github.com/leoslamas/dredd-go/rule"

func main() {
	var rule1 = rule.NewChainRule()
	var rule2 = rule.NewChainRule()

	rule1.OnEval(func(ctx rule.Context) bool {
		println("Eval Chain Rule 1")
		return ctx.GetRuleContext().Get("value").(bool) // true

	}).OnPreExecute(func(ctx rule.Context) {
		println("Pre Chain Rule 1")

	}).OnExecute(func(ctx rule.Context) {
		println("Execute Chain Rule 1")

	}).OnPostExecute(func(ctx rule.Context) {
		println("Post Chain Rule 1")

	}).AddChildren(
		rule2.OnEval(func(ctx rule.Context) bool {
			println("Eval Chain Rule 2")
			return false

		}).OnExecute(func(ctx rule.Context) {
			println("Execute Chain Rule 2") // unreachable
		}),
	)

	var ruleContext = rule.NewRuleContext()
	ruleContext.Set("value", true)

	rule.ChainRuleRunner(ruleContext, rule1)
}
