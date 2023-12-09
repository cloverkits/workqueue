package workqueue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type dcb struct {
	a0, g0, d0, r0 []any
}

func (c *dcb) OnAdd(item any) {
	c.a0 = append(c.a0, item)
}

func (c *dcb) OnGet(item any) {
	c.g0 = append(c.g0, item)
}

func (c *dcb) OnDone(item any) {
	c.d0 = append(c.d0, item)
}

func (c *dcb) OnAfter(item any, _ time.Duration) {
	c.r0 = append(c.r0, item)
}

func TestDelayingQueueWithCallback(t *testing.T) {
	c := &dcb{}
	q := NewDelayingQueue(c)
	q.AddAfter("foo", time.Second)
	q.AddAfter("bar", time.Second)
	assert.Equal(t, []any{"foo", "bar"}, c.r0)
	time.Sleep(3 * time.Second)
	assert.Equal(t, []any{"foo", "bar"}, c.a0)
	item, closed := q.Get()
	assert.Equal(t, []any{"foo"}, c.g0)
	assert.Equal(t, "foo", item)
	assert.False(t, closed)
	q.Done("foo")
	q.Done("bar")
	assert.Equal(t, []any{"foo", "bar"}, c.d0)
	q.ShutDown()
}
