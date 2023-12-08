package workqueue

import (
	"testing"
	"time"
)

func TestPriorityQueue(t *testing.T) {
	q := NewPriorityQueue(nil)
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
}

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
	q.AddWeight("foo", 4)
	q.AddWeight("bar", 2)
	if len(c.p0) != 2 {
		t.Fatal("OnDone callback failed")
	}
	q.Done("foo")
	q.Done("bar")
	time.Sleep(3 * time.Second)
	q.ShutDown()
}
