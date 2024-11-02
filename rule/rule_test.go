package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRuleContext(t *testing.T) {
	rc := NewRuleContext()
	assert.NotNil(t, rc)
	assert.Nil(t, rc.Get("nonexistent"))
}

func TestRuleContext_SetAndGet(t *testing.T) {
	rc := NewRuleContext()
	rc.Set("key", "value")
	assert.Equal(t, "value", rc.Get("key"))
}

func TestBaseRule_SetAndGetRuleContext(t *testing.T) {
	rc := NewRuleContext()
	r := &BaseRule[int]{}
	r.SetRuleContext(rc)
	assert.Equal(t, rc, r.GetRuleContext())
}

func TestBaseRule_AddChild(t *testing.T) {
	r := &BaseRule[int]{}
	child := &BaseRule[int]{}
	r.AddChildren(child)
	assert.Equal(t, 1, len(r.GetChildren()))
	assert.Equal(t, child, r.GetChildren()[0])
}

func TestBaseRule_AddChildren(t *testing.T) {
	r := &BaseRule[int]{}
	r.ruleType = bestFirstRuleType
	child1 := &BaseRule[int]{}
	child2 := &BaseRule[int]{}
	r.AddChildren(child1, child2)
	assert.Equal(t, 2, len(r.GetChildren()))
	assert.Equal(t, child1, r.GetChildren()[0])
	assert.Equal(t, child2, r.GetChildren()[1])
}

func TestBaseRule_ChainRule_Panics_When_More_Than_One_Child(t *testing.T) {
	r := &BaseRule[int]{}
	r.ruleType = chainRuleType
	child1 := &BaseRule[int]{}
	child2 := &BaseRule[int]{}
	assert.Panics(t, func() {
		r.AddChildren(child1, child2)
	})
}

func TestBaseRule_FireChainRuleType(t *testing.T) {
	r := &BaseRule[int]{ruleType: chainRuleType}
	r.OnEval(func(ctx Context) bool { return true })
	r.OnPreExecute(func(ctx Context) {})
	r.OnExecute(func(ctx Context) {})
	r.OnPostExecute(func(ctx Context) {})
	assert.True(t, r.fire())
}

func TestBaseRule_FireBestFirstRuleType(t *testing.T) {
	r := &BaseRule[int]{ruleType: bestFirstRuleType}
	r.OnEval(func(ctx Context) bool { return true })
	r.OnPreExecute(func(ctx Context) {})
	r.OnExecute(func(ctx Context) {})
	r.OnPostExecute(func(ctx Context) {})
	assert.False(t, r.fire())
}
