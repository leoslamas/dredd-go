# dredd-go

This is a port of [Dredd](https://github.com/amsterdatech/Dredd) rules engine to Go

## Dredd Rules Engine

*From the original:*

> Dredd was created to be a simple way to detach application business logic in order to create a decision tree model best for visualize and perhaps easy to understand and maintain.

---

## Chain Rule Runner 

When using the `ChainRuleRunner`, the rules will be executed in a linear sequence. When the `OnEval()` of a rule returns true, its child rule will be evaluated, continuing until there are no more child rules.

![ChainRuleRunner](img/chain-runner.png)

## Best First Rule Runner

When using the `BestFirstRuleRunner`, the rules will be executed so that if `OnEval()` returns true, the first child rule will be evaluated. If `OnEval()` returns false, the next sibling rule will be evaluated until there are no more child or sibling rules.

![alt text](img/best-first-runner.png)

## Rules

Here are some useful methods for setting up your rules:

- `OnEval()` sets the condition that determines whether the rule should execute.
- `OnExecute()` contains the main code the rule should execute.
- `OnPreExecute()` any actions the rule needs to perform beforehand.
- `OnPostExecute()` any actions the rule should perform afterward.
- `AddChildren()` helper method to add one or multiple child rules.
  
*Notes:*

* You don't need to provide all the callbacks
* Additionally, you should pass a `RuleContext` during execution, which is a map accessible from within the rules. 
* You can even mix runners and call another runner within the execution of a rule, using a new sequence of different rules from any type.

## Example

```go
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
```

Result:

```
> Eval Chain Rule 1
> Pre Chain Rule 1
> Execute Chain Rule 1
> Post Chain Rule 1
> Eval Chain Rule 2
```

## Todo

- [ ] Async rules

---

# License #

    Copyright 2015 Amsterda Technology, Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.