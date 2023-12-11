package workqueue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultSyncTimeWindow = 500 * time.Millisecond

type DelayingInterface interface {
	Interface
	// AddAfter adds an item to the workqueue after the indicated duration has passed
	AddAfter(item interface{}, duration time.Duration)
}

type DelayingCallback interface {
	Callback
	OnAfter(item any, duration time.Duration)
}

type DelayingQ struct {
	Q           // 继承 Q
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	waitingHeap *heap        // 基于对象的时间堆
	now         atomic.Int64 // 当前时间
	cb          DelayingCallback
	lockHeap    sync.Mutex
}

// 创建一个 DelayingQueue 对象
// Create a new DelayingQueue object.
func NewDelayingQueue(cb DelayingCallback) *DelayingQ {
	return newDelayingQ("", cb)
}

// 创建一个带名称的 DelayingQueue 对象
// Create a new named DelayingQueue object.
func NewNamedDelayingQueue(name string, cb DelayingCallback) *DelayingQ {
	return newDelayingQ(name, cb)
}

func newDelayingQ(name string, cb DelayingCallback) *DelayingQ {
	if cb == nil {
		cb = emptycb{}
	}
	q := &DelayingQ{
		Q:           *newQ(name, cb),
		wg:          sync.WaitGroup{},
		waitingHeap: &heap{data: make([]*waitingFor, 0, defaultQueueCap)},
		now:         atomic.Int64{},
		cb:          cb,
		lockHeap:    sync.Mutex{},
	}
	q.ctx, q.cancel = context.WithCancel(context.Background())
	q.wg.Add(2)
	q.now.Store(time.Now().UnixNano())
	go q.waitingLoop()
	go q.syncNow()
	return q
}

// 添加一个延迟任务到队列中
// Add a delayed task to the queue
func (q *DelayingQ) AddAfter(item any, duration time.Duration) {
	if q.ShuttingDown() {
		return
	}
	q.cb.OnAfter(item, duration)
	if duration <= 0 {
		q.Add(item)
		return
	}
	q.lockHeap.Lock()
	q.waitingHeap.Push(&waitingFor{
		data:  item,
		value: time.Now().Add(duration).UnixNano(),
	})
	q.lockHeap.Unlock()
}

// 关闭 Queue
// Close the queue
func (q *DelayingQ) ShutDown() {
	q.Q.ShutDown()
	q.cancel()
	q.wg.Wait()
	q.lockHeap.Lock()
	q.waitingHeap.Reset()
	q.lockHeap.Unlock()
}

// 关闭 Queue 并且等待所有的任务都被处理完
// Close the Queue and wait for all tasks to be processed
func (q *DelayingQ) ShutDownWithDrain() {
	q.Q.ShutDownWithDrain()
	q.cancel()
	q.wg.Wait()
	q.lockHeap.Lock()
	q.waitingHeap.Reset()
	q.lockHeap.Unlock()
}

// 同步当前时间, 每 500 毫秒钟同步一次, 用于更新超时标尺
// sync current time, sync once per 500 milliseconds, used to update timeout scale
func (q *DelayingQ) syncNow() {
	heartbeat := time.NewTicker(DefaultSyncTimeWindow)
	defer func() {
		q.wg.Done()
		heartbeat.Stop()
	}()

	for {
		select {
		case <-q.ctx.Done():
			return
		case <-heartbeat.C:
			q.now.Store(time.Now().UnixNano())
		}
	}
}

// 从堆中读取 WaitingFor，如果对象没有超时，就重新放回堆
// read from heap, if the object has not timed out, put it back in the heap
func (q *DelayingQ) waitingLoop() {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			q.lockHeap.Lock()
			entry := q.waitingHeap.Pop()
			q.lockHeap.Unlock()
			if entry == nil {
				continue
			}
			if entry.value < q.now.Load() {
				q.Add(entry.data)
			} else {
				q.lockHeap.Lock()
				q.waitingHeap.Push(entry)
				q.lockHeap.Unlock()
			}
		}
	}
}
