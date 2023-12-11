package workqueue

import (
	"testing"

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

func TestPriorityQueueStandard(t *testing.T) {
	t.Run("Priority", func(t *testing.T) {
		first := "foo"
		second := "bar"
		third := "baz"
		q := NewPriorityQueue(DeafultQueueSortWindows, nil)
		defer q.ShutDown()
		q.AddWeight(first, 10)
		q.AddWeight(second, 30)
		q.AddWeight(third, 5)
		item, closed := q.Get()
		assert.False(t, closed)
		assert.Equal(t, third, item)
		q.Done(item)
		item, closed = q.Get()
		assert.False(t, closed)
		assert.Equal(t, first, item)
		q.Done(item)
		item, closed = q.Get()
		assert.False(t, closed)
		assert.Equal(t, second, item)
		q.Done(item)
	})
	t.Run("TwoFireEarly", func(t *testing.T) {
		first := "foo"
		second := "bar"
		third := "baz"
		q := NewPriorityQueue(DeafultQueueSortWindows*4, nil)
		q.AddWeight(first, 10)
		q.AddWeight(second, 30)
		item, closed := q.Get()
		assert.False(t, closed)
		assert.Equal(t, first, item)
		q.Done(item)
		q.AddWeight(third, 5)
		item, closed = q.Get()
		assert.False(t, closed)
		assert.Equal(t, second, item)
		q.Done(item)
		item, closed = q.Get()
		assert.False(t, closed)
		assert.Equal(t, third, item)
		q.Done(item)
	})
	t.Run("CallbackFuncs", func(t *testing.T) {
		// c := &pcb{}
		// q := NewPriorityQueue(DeafultQueueSortWindows, c)
		// q.AddWeight("foo", 4)
		// q.AddWeight("bar", 5)
		// assert.Equal(t, []any{"foo", "bar"}, c.p0)
		// time.Sleep(100 * time.Millisecond)
		// assert.NotEqual(t, []any{"foo", "bar"}, c.a0)
		// time.Sleep(time.Second)
		// assert.Equal(t, []any{"foo", "bar"}, c.a0)
		// item, closed := q.Get()
		// assert.Equal(t, []any{"foo"}, c.g0)
		// assert.Equal(t, "foo", item)
		// assert.False(t, closed)
		// q.Done(item)
		// item, _ = q.Get()
		// q.Done(item)
		// q.ShutDown()
	})
}
