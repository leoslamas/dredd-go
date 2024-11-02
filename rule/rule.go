package rule

type ruleType int

const (
	chainRuleType ruleType = iota
	bestFirstRuleType
)

// RuleContext represents a context for storing key-value pairs.
type RuleContext struct {
	context map[string]interface{}
}

// NewRuleContext creates a new RuleContext with an initialized map.
func NewRuleContext() *RuleContext {
	return &RuleContext{context: make(map[string]interface{})}
}

// Get retrieves a value from the context by its key.
func (rc *RuleContext) Get(key string) interface{} {
	return rc.context[key]
}

// Set adds or updates a key-value pair in the context.
func (rc *RuleContext) Set(key string, value interface{}) {
	rc.context[key] = value
}

type Context interface {
	GetRuleContext() *RuleContext
	SetRuleContext(*RuleContext)
}

// BaseRule represents a generic rule with a context and various lifecycle hooks.
type BaseRule[T any] struct {
	ruleType      ruleType
	context       *RuleContext
	children      []*BaseRule[T]
	onEval        func(Context) bool
	onExecute     func(Context)
	onPreExecute  func(Context)
	onPostExecute func(Context)
}

// GetRuleContext returns the RuleContext associated with the rule.
func (r *BaseRule[T]) GetRuleContext() *RuleContext {
	return r.context
}

// SetRuleContext sets the RuleContext for the rule.
func (r *BaseRule[T]) SetRuleContext(context *RuleContext) {
	r.context = context
}

func (r *BaseRule[T]) eval() bool {
	return r.onEval(r)
}

// OnEval sets the evaluation function for the rule.
func (r *BaseRule[T]) OnEval(f func(Context) bool) *BaseRule[T] {
	r.onEval = f
	return r
}

func (r *BaseRule[T]) preExecute() {
	r.onPreExecute(r)
}

// OnPreExecute sets the pre-execution function for the rule.
func (r *BaseRule[T]) OnPreExecute(f func(Context)) *BaseRule[T] {
	r.onPreExecute = f
	return r
}

func (r *BaseRule[T]) execute() {
	r.onExecute(r)
}

// OnExecute sets the execution function for the rule.
func (r *BaseRule[T]) OnExecute(f func(Context)) *BaseRule[T] {
	r.onExecute = f
	return r
}

func (r *BaseRule[T]) postExecute() {
	r.onPostExecute(r)
}

// OnPostExecute sets the post-execution function for the rule.
func (r *BaseRule[T]) OnPostExecute(f func(Context)) *BaseRule[T] {
	r.onPostExecute = f
	return r
}

// GetChildren returns the children of the rule.
func (r *BaseRule[T]) GetChildren() []*BaseRule[T] {
	return r.children
}

// AddChildren adds child rules to the rule.
func (r *BaseRule[T]) AddChildren(rules ...*BaseRule[T]) *BaseRule[T] {
	switch r.ruleType {
	case chainRuleType:
		if len(r.children)+len(rules) > 1 {
			panic("ChainRule can only have one child")
		}
	}
	r.children = append(r.children, rules...)
	return r
}

func (r *BaseRule[T]) fire() bool {
	switch r.ruleType {
	case chainRuleType:
		if r.eval() {
			r.preExecute()
			r.execute()
			r.postExecute()
			r.runChildren()
		}
	case bestFirstRuleType:
		if r.eval() {
			r.preExecute()
			r.execute()
			r.postExecute()
			r.runChildren()
			return false
		}
	}
	return true
}

func (r *BaseRule[T]) runChildren() {
	RuleRunner(r.ruleType, r.GetRuleContext(), r.GetChildren()...)
}

// RuleRunner executes a list of rules within a given RuleContext.
func RuleRunner[T any](ruleType ruleType, ruleContext *RuleContext, rules ...*BaseRule[T]) {
	if len(rules) == 0 {
		return
	}

	switch ruleType {
	case chainRuleType:
		if len(rules) > 1 {
			panic("ChainRuleRunner only supports one rule")
		}

		var r = rules[0]
		r.SetRuleContext(ruleContext)
		r.fire()

	case bestFirstRuleType:
		for _, r := range rules {
			r.SetRuleContext(ruleContext)
			if !r.fire() {
				break
			}
		}
	}
}
