package workqueue

import (
	"sync"
)

const defaultQueueCap = 2048

// Interface is a generic interface to a queue.
type Interface interface {
	Add(item any)
	Len() int
	Get() (item any, closed bool)
	Done(item any)
	ShutDown()
	ShutDownWithDrain()
	ShuttingDown() bool
}

type Callback interface {
	OnDone(any)
	OnAdd(any)
	OnGet(any)
}

type Q struct {
	queue      []any
	dirty      set
	processing set
	closed     bool
	drain      bool
	once       sync.Once
	cond       *sync.Cond
	name       string
	cb         Callback
}

type emptycb struct{}

func (emptycb) OnDone(_ any) {}
func (emptycb) OnAdd(_ any)  {}
func (emptycb) OnGet(_ any)  {}

func NewQueue(cb Callback) *Q {
	return newQ("", cb)
}

func NewNamedQueue(name string, cb Callback) *Q {
	return newQ(name, cb)
}

func newQ(name string, cb Callback) *Q {
	if cb == nil {
		cb = emptycb{}
	}
	return &Q{
		queue:      make([]any, 0, defaultQueueCap),
		dirty:      set{},
		processing: set{},
		closed:     false,
		drain:      false,
		cond:       sync.NewCond(&sync.Mutex{}),
		name:       name,
		cb:         cb,
	}
}

// 执行关闭动作
func (q *Q) shutdown() {
	q.closed = true
	q.cond.Broadcast()
}

func (q *Q) drainS() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.drain
}

// 判断当前的 Queue 是否已经关闭
// Determine if the queue is shutting down.
func (q *Q) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.closed
}

// 关闭 Queue
// Close the queue
func (q *Q) ShutDown() {
	q.once.Do(func() {
		q.cond.L.Lock()
		defer q.cond.L.Unlock()
		q.drain = false
		q.shutdown()
	})
}

// 关闭 Queue 并且等待所有的任务都被处理完
// Close the Queue and wait for all tasks to be processed
func (q *Q) ShutDownWithDrain() {
	q.once.Do(func() {
		q.cond.L.Lock()
		defer q.cond.L.Unlock()
		q.drain = true
		q.shutdown()
		for q.processing.len() > 0 && q.drainS() {
			q.cond.Wait()
		}
	})
}

// 获得 Queue 对象的数量
// Get the number of items in the queue.
func (q *Q) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.queue)
}

// 确认这个对象已经被处理，可以被释放
// Mark an item as done processing.
func (q *Q) Done(item any) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.cb.OnDone(item)
	q.processing.delete(item)
	if q.dirty.has(item) {
		q.queue = append(q.queue, item)
		q.cond.Signal()
	} else if q.processing.len() == 0 {
		q.cond.Signal()
	}
}

// 向队列中添加一个对象
// Add an item to the queue.
func (q *Q) Add(item any) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.closed || q.dirty.has(item) {
		return
	}
	q.cb.OnAdd(item)
	q.dirty.insert(item)
	if q.processing.has(item) {
		return // (去重) already processing, don't add it to the queue
	}
	q.queue = append(q.queue, item)
	q.cond.Signal()
}

// 从队列中获取一个对象
// Get an item from the queue.
func (q *Q) Get() (item any, closed bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for len(q.queue) == 0 && !q.closed {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		// We must be shutting down.
		return nil, true
	}
	item = q.queue[0]
	// The underlying array still exists and reference this object, so the object will not be garbage collected.
	q.queue[0] = nil
	q.queue = q.queue[1:]
	q.cb.OnGet(item)
	q.processing.insert(item)
	q.dirty.delete(item)
	return item, false
}
