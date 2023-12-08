package workqueue

import (
	"testing"
	// queue "github.com/shengyanli1982/workqueue"
)

func TestQueue(t *testing.T) {
	q := NewQueue(nil)
	if q == nil {
		t.Fatal("New() returned nil")
	}
	if q.Len() != 0 {
		t.Fatal("New() returned non-empty queue")
	}
	if q.ShuttingDown() {
		t.Fatal("New() returned shutting down queue")
	}
	q.Add("foo")
	if q.Len() != 1 {
		t.Fatal("Add() failed")
	}
	item, closed := q.Get()
	if closed {
		t.Fatal("Get() returned closed")
	}
	if item != "foo" {
		t.Fatal("Get() returned wrong item")
	}
	q.Done(item)
	if q.Len() != 0 {
		t.Fatal("Done() failed")
	}
	q.Add("foo")
	q.Add("bar")
	q.ShutDown()
	if !q.ShuttingDown() {
		t.Fatal("ShutDown() failed")
	}
	item, closed = q.Get()
	if closed {
		t.Fatal("Get() returned closed")
	}
	if item != "foo" {
		t.Fatal("Get() returned wrong item")
	}
	q.Done(item)
	item, closed = q.Get()
	if closed {
		t.Fatal("Get() returned closed")
	}
	if item != "bar" {
		t.Fatal("Get() returned wrong item")
	}
	q.Done(item)
	item, closed = q.Get()
	if !closed {
		t.Fatal("Get() returned not closed")
	}
	if item != nil {
		t.Fatal("Get() returned wrong item")
	}
	q.Done(item)
	q.ShutDown()
}

type cb struct {
	a0, g0, d0 []any
}

func (c *cb) OnAdd(item any) {
	c.a0 = append(c.a0, item)
}

func (c *cb) OnGet(item any) {
	c.g0 = append(c.g0, item)
}

func (c *cb) OnDone(item any) {
	c.d0 = append(c.d0, item)
}

func TestQueueCallback(t *testing.T) {
	c := &cb{
		a0: make([]any, 0),
		g0: make([]any, 0),
		d0: make([]any, 0),
	}
	q := NewQueue(c)
	q.Add("foo")
	q.Add("bar")
	if len(c.a0) != 2 {
		t.Fatal("OnAdd callback failed")
	}
	_, closed := q.Get()
	if closed {
		t.Fatal("Get() returned closed")
	}
	if len(c.g0) != 1 {
		t.Fatal("OnGet callback failed")
	}
	q.Done("foo")
	q.Done("bar")
	if len(c.d0) != 2 {
		t.Fatal("OnDone callback failed")
	}
	q.ShutDown()
}
