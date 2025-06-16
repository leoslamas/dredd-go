// Package rule provides a flexible rules engine with support for chain and best-first execution strategies.
// It allows building decision trees with configurable evaluation and execution hooks.
package rule

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// RuleType represents the execution strategy for rules.
type RuleType int

const (
	// ChainRuleType executes rules in a linear sequence.
	ChainRuleType RuleType = iota
	// BestFirstRuleType executes rules in a tree-based manner.
	BestFirstRuleType
)

// String implements the fmt.Stringer interface for RuleType.
func (rt RuleType) String() string {
	switch rt {
	case ChainRuleType:
		return "ChainRule"
	case BestFirstRuleType:
		return "BestFirstRule"
	default:
		return "UnknownRuleType"
	}
}

// Common errors for the rules engine.
var (
	ErrChainRuleMultipleChildren = errors.New("chain rule can only have one child")
	ErrChainRuleMultipleRules    = errors.New("chain rule runner only supports one rule")
	ErrNilRuleContext            = errors.New("rule context cannot be nil")
	ErrNilRule                   = errors.New("rule cannot be nil")
	ErrInvalidRuleType           = errors.New("invalid rule type")
)

// RuleContext represents a thread-safe context for storing typed key-value pairs.
type RuleContext[T any] struct {
	mu      sync.RWMutex
	context map[string]T
}

// NewRuleContext creates a new RuleContext with an initialized map.
func NewRuleContext[T any]() *RuleContext[T] {
	return &RuleContext[T]{context: make(map[string]T)}
}

// NewRuleContextWithCapacity creates a new RuleContext with a pre-allocated map of the given capacity.
func NewRuleContextWithCapacity[T any](capacity int) *RuleContext[T] {
	return &RuleContext[T]{context: make(map[string]T, capacity)}
}

// Get retrieves a value from the context by its key.
// Returns the value and a boolean indicating if the key was found.
func (rc *RuleContext[T]) Get(key string) (T, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	value, ok := rc.context[key]
	return value, ok
}

// MustGet retrieves a value from the context by its key.
// Panics if the key is not found.
func (rc *RuleContext[T]) MustGet(key string) T {
	value, ok := rc.Get(key)
	if !ok {
		panic(fmt.Sprintf("key '%s' not found in rule context", key))
	}
	return value
}

// Set adds or updates a key-value pair in the context.
func (rc *RuleContext[T]) Set(key string, value T) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.context[key] = value
}

// Delete removes a key-value pair from the context.
func (rc *RuleContext[T]) Delete(key string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	delete(rc.context, key)
}

// Exists checks if a key exists in the context.
func (rc *RuleContext[T]) Exists(key string) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	_, ok := rc.context[key]
	return ok
}

// Keys returns all keys in the context.
func (rc *RuleContext[T]) Keys() []string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	keys := make([]string, 0, len(rc.context))
	for k := range rc.context {
		keys = append(keys, k)
	}
	return keys
}

// Size returns the number of key-value pairs in the context.
func (rc *RuleContext[T]) Size() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return len(rc.context)
}

// Context defines the interface for rule execution context.
type Context[T any] interface {
	GetRuleContext() *RuleContext[T]
	SetRuleContext(*RuleContext[T])
	GetGoContext() context.Context
}

// EvaluationResult represents the result of rule evaluation.
type EvaluationResult struct {
	ShouldExecute bool
	Error         error
}

// ExecutionResult represents the result of rule execution.
type ExecutionResult struct {
	Error error
}

// BaseRule represents a generic rule with a context and various lifecycle hooks.
type BaseRule[T, C any] struct {
	ruleType      RuleType
	context       *RuleContext[C]
	goContext     context.Context
	children      []*BaseRule[T, C]
	onEval        func(Context[C]) EvaluationResult
	onExecute     func(Context[C]) ExecutionResult
	onPreExecute  func(Context[C]) ExecutionResult
	onPostExecute func(Context[C]) ExecutionResult
}

// GetRuleContext returns the RuleContext associated with the rule.
func (r *BaseRule[T, C]) GetRuleContext() *RuleContext[C] {
	return r.context
}

// SetRuleContext sets the RuleContext for the rule.
func (r *BaseRule[T, C]) SetRuleContext(context *RuleContext[C]) {
	r.context = context
}

// GetGoContext returns the Go context associated with the rule.
func (r *BaseRule[T, C]) GetGoContext() context.Context {
	return r.goContext
}

// SetGoContext sets the Go context for the rule.
func (r *BaseRule[T, C]) SetGoContext(ctx context.Context) {
	r.goContext = ctx
}

func (r *BaseRule[T, C]) eval() EvaluationResult {
	if r.onEval == nil {
		return EvaluationResult{ShouldExecute: true, Error: nil}
	}
	return r.onEval(r)
}

// OnEval sets the evaluation function for the rule with a simple boolean return.
func (r *BaseRule[T, C]) OnEval(f func(Context[C]) bool) *BaseRule[T, C] {
	r.onEval = func(ctx Context[C]) EvaluationResult {
		return EvaluationResult{ShouldExecute: f(ctx), Error: nil}
	}
	return r
}

// OnEvalWithError sets the evaluation function for the rule with full error handling.
func (r *BaseRule[T, C]) OnEvalWithError(f func(Context[C]) EvaluationResult) *BaseRule[T, C] {
	r.onEval = f
	return r
}

func (r *BaseRule[T, C]) preExecute() ExecutionResult {
	if r.onPreExecute == nil {
		return ExecutionResult{Error: nil}
	}
	return r.onPreExecute(r)
}

// OnPreExecute sets the pre-execution function for the rule with no return value.
func (r *BaseRule[T, C]) OnPreExecute(f func(Context[C])) *BaseRule[T, C] {
	r.onPreExecute = func(ctx Context[C]) ExecutionResult {
		f(ctx)
		return ExecutionResult{Error: nil}
	}
	return r
}

// OnPreExecuteWithError sets the pre-execution function for the rule with error handling.
func (r *BaseRule[T, C]) OnPreExecuteWithError(f func(Context[C]) ExecutionResult) *BaseRule[T, C] {
	r.onPreExecute = f
	return r
}

func (r *BaseRule[T, C]) execute() ExecutionResult {
	if r.onExecute == nil {
		return ExecutionResult{Error: nil}
	}
	return r.onExecute(r)
}

// OnExecute sets the execution function for the rule with no return value.
func (r *BaseRule[T, C]) OnExecute(f func(Context[C])) *BaseRule[T, C] {
	r.onExecute = func(ctx Context[C]) ExecutionResult {
		f(ctx)
		return ExecutionResult{Error: nil}
	}
	return r
}

// OnExecuteWithError sets the execution function for the rule with error handling.
func (r *BaseRule[T, C]) OnExecuteWithError(f func(Context[C]) ExecutionResult) *BaseRule[T, C] {
	r.onExecute = f
	return r
}

func (r *BaseRule[T, C]) postExecute() ExecutionResult {
	if r.onPostExecute == nil {
		return ExecutionResult{Error: nil}
	}
	return r.onPostExecute(r)
}

// OnPostExecute sets the post-execution function for the rule with no return value.
func (r *BaseRule[T, C]) OnPostExecute(f func(Context[C])) *BaseRule[T, C] {
	r.onPostExecute = func(ctx Context[C]) ExecutionResult {
		f(ctx)
		return ExecutionResult{Error: nil}
	}
	return r
}

// OnPostExecuteWithError sets the post-execution function for the rule with error handling.
func (r *BaseRule[T, C]) OnPostExecuteWithError(f func(Context[C]) ExecutionResult) *BaseRule[T, C] {
	r.onPostExecute = f
	return r
}

// GetChildren returns the children of the rule.
func (r *BaseRule[T, C]) GetChildren() []*BaseRule[T, C] {
	return r.children
}

// HasChildren returns true if the rule has any children.
func (r *BaseRule[T, C]) HasChildren() bool {
	return len(r.children) > 0
}

// ChildrenCount returns the number of children.
func (r *BaseRule[T, C]) ChildrenCount() int {
	return len(r.children)
}

// AddChildren adds child rules to the rule.
// Returns an error if the operation violates rule constraints.
func (r *BaseRule[T, C]) AddChildren(rules ...*BaseRule[T, C]) error {
	if r.ruleType == ChainRuleType && len(r.children)+len(rules) > 1 {
		return ErrChainRuleMultipleChildren
	}

	for _, rule := range rules {
		if rule == nil {
			return ErrNilRule
		}
	}

	r.children = append(r.children, rules...)
	return nil
}

// MustAddChildren adds child rules to the rule, panicking on error.
// This is a convenience method for backward compatibility.
func (r *BaseRule[T, C]) MustAddChildren(rules ...*BaseRule[T, C]) *BaseRule[T, C] {
	if err := r.AddChildren(rules...); err != nil {
		panic(err)
	}
	return r
}

func (r *BaseRule[T, C]) fire() (bool, error) {
	// Check context cancellation
	if r.goContext != nil {
		select {
		case <-r.goContext.Done():
			return false, r.goContext.Err()
		default:
		}
	}

	switch r.ruleType {
	case ChainRuleType:
		evalResult := r.eval()
		if evalResult.Error != nil {
			return false, evalResult.Error
		}
		if evalResult.ShouldExecute {
			if result := r.preExecute(); result.Error != nil {
				return false, result.Error
			}
			if result := r.execute(); result.Error != nil {
				return false, result.Error
			}
			if result := r.postExecute(); result.Error != nil {
				return false, result.Error
			}
			if err := r.runChildren(); err != nil {
				return false, err
			}
		}
	case BestFirstRuleType:
		evalResult := r.eval()
		if evalResult.Error != nil {
			return false, evalResult.Error
		}
		if evalResult.ShouldExecute {
			if result := r.preExecute(); result.Error != nil {
				return false, result.Error
			}
			if result := r.execute(); result.Error != nil {
				return false, result.Error
			}
			if result := r.postExecute(); result.Error != nil {
				return false, result.Error
			}
			if err := r.runChildren(); err != nil {
				return false, err
			}
			return false, nil
		}
	default:
		return false, ErrInvalidRuleType
	}
	return true, nil
}

func (r *BaseRule[T, C]) runChildren() error {
	return RuleRunner(r.ruleType, r.goContext, r.GetRuleContext(), r.GetChildren()...)
}

// RuleRunner executes a list of rules within a given RuleContext.
// Returns an error if execution fails.
func RuleRunner[T, C any](ruleType RuleType, goCtx context.Context, ruleContext *RuleContext[C], rules ...*BaseRule[T, C]) error {
	if ruleContext == nil {
		return ErrNilRuleContext
	}

	if len(rules) == 0 {
		return nil
	}

	// Validate all rules are not nil
	for i, rule := range rules {
		if rule == nil {
			return fmt.Errorf("rule at index %d is nil: %w", i, ErrNilRule)
		}
	}

	switch ruleType {
	case ChainRuleType:
		if len(rules) > 1 {
			return ErrChainRuleMultipleRules
		}

		r := rules[0]
		r.SetRuleContext(ruleContext)
		r.SetGoContext(goCtx)
		_, err := r.fire()
		return err

	case BestFirstRuleType:
		for _, r := range rules {
			r.SetRuleContext(ruleContext)
			r.SetGoContext(goCtx)
			continueLoop, err := r.fire()
			if err != nil {
				return err
			}
			if !continueLoop {
				break
			}
		}
		return nil

	default:
		return ErrInvalidRuleType
	}
}

// RuleOption represents a functional option for configuring rules.
type RuleOption[T, C any] func(*BaseRule[T, C])

// WithEvaluationWithError sets the evaluation function for the rule with full error handling.
func WithEvaluationWithError[T, C any](f func(Context[C]) EvaluationResult) RuleOption[T, C] {
	return func(r *BaseRule[T, C]) {
		r.onEval = f
	}
}

// WithEvaluation sets the evaluation function for the rule with boolean return.
func WithEvaluation[T, C any](f func(Context[C]) bool) RuleOption[T, C] {
	return func(r *BaseRule[T, C]) {
		r.onEval = func(ctx Context[C]) EvaluationResult {
			return EvaluationResult{ShouldExecute: f(ctx), Error: nil}
		}
	}
}

// WithExecutionWithError sets the execution function for the rule with error handling.
func WithExecutionWithError[T, C any](f func(Context[C]) ExecutionResult) RuleOption[T, C] {
	return func(r *BaseRule[T, C]) {
		r.onExecute = f
	}
}

// WithExecution sets the execution function for the rule with no return.
func WithExecution[T, C any](f func(Context[C])) RuleOption[T, C] {
	return func(r *BaseRule[T, C]) {
		r.onExecute = func(ctx Context[C]) ExecutionResult {
			f(ctx)
			return ExecutionResult{Error: nil}
		}
	}
}

// WithPreExecutionWithError sets the pre-execution function for the rule with error handling.
func WithPreExecutionWithError[T, C any](f func(Context[C]) ExecutionResult) RuleOption[T, C] {
	return func(r *BaseRule[T, C]) {
		r.onPreExecute = f
	}
}

// WithPreExecution sets the pre-execution function for the rule with no return.
func WithPreExecution[T, C any](f func(Context[C])) RuleOption[T, C] {
	return func(r *BaseRule[T, C]) {
		r.onPreExecute = func(ctx Context[C]) ExecutionResult {
			f(ctx)
			return ExecutionResult{Error: nil}
		}
	}
}

// WithPostExecutionWithError sets the post-execution function for the rule with error handling.
func WithPostExecutionWithError[T, C any](f func(Context[C]) ExecutionResult) RuleOption[T, C] {
	return func(r *BaseRule[T, C]) {
		r.onPostExecute = f
	}
}

// WithPostExecution sets the post-execution function for the rule with no return.
func WithPostExecution[T, C any](f func(Context[C])) RuleOption[T, C] {
	return func(r *BaseRule[T, C]) {
		r.onPostExecute = func(ctx Context[C]) ExecutionResult {
			f(ctx)
			return ExecutionResult{Error: nil}
		}
	}
}

// WithChildren sets the children for the rule.
func WithChildren[T, C any](children ...*BaseRule[T, C]) RuleOption[T, C] {
	return func(r *BaseRule[T, C]) {
		if err := r.AddChildren(children...); err != nil {
			panic(err) // For functional options, we panic on configuration errors
		}
	}
}

// NewBaseRule creates a new BaseRule with the given type and options.
func NewBaseRule[T, C any](ruleType RuleType, options ...RuleOption[T, C]) *BaseRule[T, C] {
	rule := &BaseRule[T, C]{
		ruleType:  ruleType,
		context:   NewRuleContext[C](),
		goContext: context.Background(),
		children:  make([]*BaseRule[T, C], 0),
	}

	for _, option := range options {
		option(rule)
	}

	return rule
}

// String implements the fmt.Stringer interface for BaseRule.
func (r *BaseRule[T, C]) String() string {
	return fmt.Sprintf("BaseRule{type: %s, children: %d}", r.ruleType, len(r.children))
}
