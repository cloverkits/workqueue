package workqueue

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

func TestQueueStandard(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		tests := []struct {
			queue         *Q
			queueShutDown func(Interface)
		}{
			{
				queue:         NewQueue(nil),
				queueShutDown: Interface.ShutDown,
			},
			{
				queue:         NewQueue(nil),
				queueShutDown: Interface.ShutDownWithDrain,
			},
		}

		for _, test := range tests {
			// Start producers
			const producers = 50
			producerWG := sync.WaitGroup{}
			producerWG.Add(producers)
			for i := 0; i < producers; i++ {
				go func(i int) {
					defer producerWG.Done()
					for j := 0; j < 50; j++ {
						test.queue.Add(i)
						time.Sleep(time.Millisecond)
					}
				}(i)
			}

			// Start consumers
			const consumers = 10
			consumerWG := sync.WaitGroup{}
			consumerWG.Add(consumers)
			for i := 0; i < consumers; i++ {
				go func(i int) {
					defer consumerWG.Done()
					for {
						item, quit := test.queue.Get()
						if item == "added after shutdown!" {
							t.Errorf("Got an item added after shutdown.")
						}
						if quit {
							return
						}
						t.Logf("Worker %v: begin processing %v", i, item)
						time.Sleep(3 * time.Millisecond)
						t.Logf("Worker %v: done processing %v", i, item)
						test.queue.Done(item)
					}
				}(i)
			}

			producerWG.Wait()
			test.queueShutDown(test.queue)
			test.queue.Add("added after shutdown!")
			consumerWG.Wait()
			if test.queue.Len() != 0 {
				t.Errorf("Expected the queue to be empty, had: %v items", test.queue.Len())
			}
		}
	})
	t.Run("AddWhileProcessing", func(t *testing.T) {
		tests := []struct {
			queue         *Q
			queueShutDown func(Interface)
		}{
			{
				queue:         NewQueue(nil),
				queueShutDown: Interface.ShutDown,
			},
			{
				queue:         NewQueue(nil),
				queueShutDown: Interface.ShutDownWithDrain,
			},
		}

		for _, test := range tests {
			// Start producers
			const producers = 50
			producerWG := sync.WaitGroup{}
			producerWG.Add(producers)
			for i := 0; i < producers; i++ {
				go func(i int) {
					defer producerWG.Done()
					test.queue.Add(i)
				}(i)
			}

			// Start consumers
			const consumers = 10
			consumerWG := sync.WaitGroup{}
			consumerWG.Add(consumers)
			for i := 0; i < consumers; i++ {
				go func(i int) {
					defer consumerWG.Done()
					// Every worker will re-add every item up to two times.
					// This tests the dirty-while-processing case.
					counters := map[interface{}]int{}
					for {
						item, quit := test.queue.Get()
						if quit {
							return
						}
						counters[item]++
						if counters[item] < 2 {
							test.queue.Add(item)
						}
						test.queue.Done(item)
					}
				}(i)
			}

			producerWG.Wait()
			test.queueShutDown(test.queue)
			consumerWG.Wait()
			if test.queue.Len() != 0 {
				t.Errorf("Expected the queue to be empty, had: %v items", test.queue.Len())
			}
		}
	})
	t.Run("ReInsert", func(t *testing.T) {
		q := NewQueue(nil)

		q.Add("foo")

		// Start processing
		i, _ := q.Get()
		assert.Equal(t, "foo", i)

		// Add it back while processing
		q.Add(i)

		// Finish it up
		q.Done(i)

		// It should be back on the queue
		i, _ = q.Get()
		assert.Equal(t, "foo", i)

		// Finish that one up
		q.Done(i)
		assert.Equal(t, 0, q.Len())

		q.shutdown()
	})
	t.Run("GarbageCollection", func(t *testing.T) {
		type bigObject struct {
			data []byte
		}
		leakQueue := NewQueue(nil)
		t.Cleanup(func() {
			// Make sure leakQueue doesn't go out of scope too early
			runtime.KeepAlive(leakQueue)
		})
		c := &bigObject{data: []byte("hello")}
		mustGarbageCollect(t, c)
		leakQueue.Add(c)
		o, _ := leakQueue.Get()
		leakQueue.Done(o)
	})
	t.Run("CallbackFuncs", func(t *testing.T) {
		c := &cb{}
		q := NewQueue(c)
		q.Add("foo")
		assert.Equal(t, []any{"foo"}, c.a0)
		item, closed := q.Get()
		assert.Equal(t, []any{"foo"}, c.g0)
		assert.Equal(t, "foo", item)
		assert.False(t, closed)
		q.Done(item)
		assert.Equal(t, []any{"foo"}, c.d0)
		q.ShutDown()
	})
}

func TestQueueShutdown(t *testing.T) {
	t.Run("ShutDownWithDrain", func(t *testing.T) {
		q := NewQueue(nil)

		q.Add("foo")
		q.Add("bar")

		firstItem, _ := q.Get()
		secondItem, _ := q.Get()

		finishedWG := sync.WaitGroup{}
		finishedWG.Add(1)
		go func() {
			defer finishedWG.Done()
			q.ShutDownWithDrain()
		}()

		// This is done as to simulate a sequence of events where ShutDownWithDrain
		// is called before we start marking all items as done - thus simulating a
		// drain where we wait for all items to finish processing.
		shuttingDown := false
		for !shuttingDown {
			_, shuttingDown = q.Get()
		}

		// Mark the first two items as done, as to finish up
		q.Done(firstItem)
		q.Done(secondItem)

		finishedWG.Wait()
	})
	t.Run("ShutDown", func(t *testing.T) {
		q := NewQueue(nil)

		q.Add("foo")
		q.Add("bar")

		q.Get()
		q.Get()

		finishedWG := sync.WaitGroup{}
		finishedWG.Add(1)
		go func() {
			defer finishedWG.Done()
			// Invoke ShutDown: suspending the execution immediately.
			q.ShutDown()
		}()

		// We can now do this and not have the test timeout because we didn't call
		// Done on the first two items before arriving here.
		finishedWG.Wait()
	})
	t.Run("DrainWithDirtyItem", func(t *testing.T) {
		q := NewQueue(nil)

		q.Add("foo")
		gotten, _ := q.Get()
		q.Add("foo")

		finishedWG := sync.WaitGroup{}
		finishedWG.Add(1)
		go func() {
			defer finishedWG.Done()
			q.ShutDownWithDrain()
		}()

		// Ensure that ShutDownWithDrain has started and is blocked.
		shuttingDown := false
		for !shuttingDown {
			_, shuttingDown = q.Get()
		}

		// Finish "working".
		q.Done(gotten)

		// `shuttingDown` becomes false because Done caused an item to go back into
		// the queue.
		again, shuttingDown := q.Get()
		assert.False(t, shuttingDown, "should not have been done")
		q.Done(again)

		// Now we are really done.
		_, shuttingDown = q.Get()
		assert.True(t, shuttingDown, "should have been done")
		finishedWG.Wait()
	})
	t.Run("ForceQueueShutdown", func(t *testing.T) {
		q := NewQueue(nil)

		q.Add("foo")
		q.Add("bar")

		q.Get()
		q.Get()

		finishedWG := sync.WaitGroup{}
		finishedWG.Add(1)
		go func() {
			defer finishedWG.Done()
			q.ShutDownWithDrain()
		}()

		// This is done as to simulate a sequence of events where ShutDownWithDrain
		// is called before ShutDown
		shuttingDown := false
		for !shuttingDown {
			_, shuttingDown = q.Get()
		}

		// Use ShutDown to force the queue to shut down (simulating a caller
		// which can invoke this function on a second SIGTERM/SIGINT)
		q.ShutDown()

		// We can now do this and not have the test timeout because we didn't call
		// done on any of the items before arriving here.
		finishedWG.Wait()
	})

}

// mustGarbageCollect asserts than an object was garbage collected by the end of the test.
// The input must be a pointer to an object.
func mustGarbageCollect(t *testing.T, i interface{}) {
	t.Helper()
	var collected int32 = 0
	runtime.SetFinalizer(i, func(x interface{}) {
		atomic.StoreInt32(&collected, 1)
	})
	t.Cleanup(func() {
		runtime.GC()
		if atomic.LoadInt32(&collected) != 1 {
			assert.Fail(t, "object was not garbage collected")
		}
	})
}
