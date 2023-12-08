package workqueue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type DelayingInterface interface {
	Interface
	// AddAfter adds an item to the workqueue after the indicated duration has passed
	AddAfter(item interface{}, duration time.Duration)
}

type DelayingCallback interface {
	Callback
	OnRetry(item any, duration time.Duration)
}

type WaitingFor struct {
	data     any
	expireAt int64
	index    int
}

type DelayingQ struct {
	Q           // 继承 Q
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	waitingHeap *heap        // 基于对象的时间堆
	now         atomic.Int64 // 当前时间
	cb          DelayingCallback
}

func NewDelayingQueue(cb DelayingCallback) *DelayingQ {
	return newDelayingQ("", cb)
}

func NewNamedDelayingQueue(name string, cb DelayingCallback) *DelayingQ {
	return newDelayingQ(name, cb)
}

func newDelayingQ(name string, cb DelayingCallback) *DelayingQ {
	q := &DelayingQ{
		Q:           *newQ(name, cb),
		wg:          sync.WaitGroup{},
		waitingHeap: &heap{data: make([]*WaitingFor, 0, defaultQueueCap)},
		now:         atomic.Int64{},
		cb:          cb,
	}
	q.ctx, q.cancel = context.WithCancel(context.Background())
	q.wg.Add(2)
	go q.waitingLoop()
	go q.syncNow()
	return q
}

// 添加一个延迟任务到队列中
// Add a delayed task to the queue
func (q *DelayingQ) AddAfter(item interface{}, duration time.Duration) {
	if q.ShuttingDown() {
		return
	}
	q.cb.OnRetry(item, duration)
	if duration <= 0 {
		q.Add(item)
		return
	}
	q.waitingHeap.Push(&WaitingFor{
		data:     item,
		expireAt: time.Now().Add(duration).UnixNano(),
	})
}

// 关闭 Queue
// Close the queue
func (q *DelayingQ) ShutDown() {
	q.once.Do(func() {
		q.cond.L.Lock()
		defer q.cond.L.Unlock()
		q.drain = false
		q.shutdown()
		q.cancel()
		q.wg.Wait()
		q.waitingHeap.Reset()
	})
}

// 关闭 Queue 并且等待所有的任务都被处理完
// Close the Queue and wait for all tasks to be processed
func (q *DelayingQ) ShutDownWithDrain() {
	q.once.Do(func() {
		q.cond.L.Lock()
		defer q.cond.L.Unlock()
		q.drain = true
		q.shutdown()
		for q.processing.len() > 0 && q.drainS() {
			q.cond.Wait()
		}
		q.cancel()
		q.wg.Wait()
		q.waitingHeap.Reset()
	})
}

func (q *DelayingQ) syncNow() {
	heartbeat := time.NewTicker(time.Second)
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
	heartbeat := time.NewTicker(5 * time.Second)
	defer func() {
		q.wg.Done()
		heartbeat.Stop()
	}()

	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			entry := q.waitingHeap.Pop()
			if entry == nil {
				continue
			}
			if entry.expireAt < q.now.Load() {
				q.Add(entry.data)
			} else {
				q.waitingHeap.Push(entry)
				<-heartbeat.C // 已经到读到没有超时的对象，等待一段时间再读
			}
		}
	}
}
