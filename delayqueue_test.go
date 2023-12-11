package workqueue

import (
	"errors"
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

func TestDelayingQueueStandard(t *testing.T) {
	t.Run("Delaying", func(t *testing.T) {
		q := NewDelayingQueue(nil)
		defer q.ShutDown()
		q.AddAfter(time.Now().Local().UnixMilli(), 2*time.Second)
		item, closed := q.Get()
		assert.False(t, closed)
		q.Done(item)
		if item.(int64) > time.Now().UnixMilli() {
			assert.Error(t, errors.New("item should not be ready yet"))
		} else {
			assert.Equal(t, (time.Now().UnixMilli()-item.(int64))/1000, int64(2)) // 2秒后被操作。 2 seconds after the operation.
			return
		}
	})
	// 因为时间同步是 500 毫秒一次，所以这里延迟小于 500 毫秒的延迟任务会被延迟到 500 毫秒后执行。
	t.Run("DelayingSmallTimeSlice", func(t *testing.T) {
		q := NewDelayingQueue(nil)
		defer q.ShutDown()
		q.AddAfter(time.Now().Local().UnixMilli(), 2*time.Millisecond)
		item, closed := q.Get()
		assert.False(t, closed)
		q.Done(item)
		if item.(int64) > time.Now().UnixMilli() {
			assert.Error(t, errors.New("item should not be ready yet"))
		} else {
			assert.Equal(t, (time.Now().UnixMilli()-item.(int64))/100, int64(5)) // 500 毫秒后被操作。 500 milliseconds after the operation.
			return
		}
	})
	t.Run("TwoFireEarly", func(t *testing.T) {
		first := "foo"
		second := "bar"
		third := "baz"
		q := NewDelayingQueue(nil)
		defer q.ShutDown()
		q.AddAfter(first, time.Second)
		q.AddAfter(second, 50*time.Millisecond)
		time.Sleep(600 * time.Millisecond)
		item, closed := q.Get()
		assert.False(t, closed)
		assert.Equal(t, item, second)
		q.Done(item)
		q.AddAfter(third, 2*time.Second)
		time.Sleep(time.Second)
		item, closed = q.Get()
		assert.False(t, closed)
		assert.Equal(t, item, first)
		q.Done(item)
		time.Sleep(2 * time.Second)
		item, closed = q.Get()
		assert.False(t, closed)
		assert.Equal(t, item, third)
		q.Done(item)
	})
	t.Run("CopyShifting", func(t *testing.T) {
		first := "foo"
		second := "bar"
		third := "baz"
		q := NewDelayingQueue(nil)
		defer q.ShutDown()
		q.AddAfter(first, 1*time.Second)
		q.AddAfter(second, 500*time.Millisecond)
		q.AddAfter(third, 250*time.Millisecond)
		time.Sleep(time.Second)
		item, closed := q.Get()
		assert.False(t, closed)
		assert.Equal(t, item, third)
		q.Done(item)
		item, closed = q.Get()
		assert.False(t, closed)
		assert.Equal(t, item, second)
		q.Done(item)
		item, closed = q.Get()
		assert.False(t, closed)
		assert.Equal(t, item, first)
		q.Done(item)
	})
	t.Run("CallbackFuncs", func(t *testing.T) {
		c := &dcb{}
		q := NewDelayingQueue(c)
		q.AddAfter("foo", 50*time.Millisecond)
		q.AddAfter("bar", 50*time.Millisecond)
		assert.Equal(t, []any{"foo", "bar"}, c.r0)
		time.Sleep(100 * time.Millisecond)
		assert.NotEqual(t, []any{"foo", "bar"}, c.a0)
		time.Sleep(time.Second)
		assert.Equal(t, []any{"foo", "bar"}, c.a0)
		item, closed := q.Get()
		assert.Equal(t, []any{"foo"}, c.g0)
		assert.Equal(t, "foo", item)
		assert.False(t, closed)
		q.Done(item)
		item, _ = q.Get()
		q.Done(item)
		q.ShutDown()
	})
}
