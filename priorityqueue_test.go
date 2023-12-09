package workqueue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)



type pcb struct {
	a0, g0, d0, p0 []any
}

func (c *pcb) OnAdd(item any) {
	c.a0 = append(c.a0, item)
}

func (c *pcb) OnGet(item any) {
	c.g0 = append(c.g0, item)
}

func (c *pcb) OnDone(item any) {
	c.d0 = append(c.d0, item)
}

func (c *pcb) OnWeight(item any, _ int) {
	c.p0 = append(c.p0, item)
}

func TestPriorityQueueWithCallback(t *testing.T) {
	c := &pcb{}
	q := NewPriorityQueue(c)
	q.AddWeight("foo", 3)
	q.AddWeight("bar", 2)
	assert.Equal(t, []any{"foo", "bar"}, c.p0)
	time.Sleep(3 * time.Second)
	assert.Equal(t, []any{"bar", "foo"}, c.a0)
	item, closed := q.Get()
	assert.Equal(t, []any{"bar"}, c.g0)
	assert.Equal(t, "bar", item)
	assert.False(t, closed)
	q.Done("foo")
	q.Done("bar")
	assert.Equal(t, []any{"foo", "bar"}, c.d0)
	q.ShutDown()
}
