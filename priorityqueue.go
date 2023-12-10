package workqueue

import (
	"context"
	"sync"
)

type PriorityInterface interface {
	Interface
	// AddWeight adds an item to the workqueue with the given priority. priority low is better, 0 is imddiatly process
	AddWeight(item interface{}, priority int)
}

type PriorityCallback interface {
	Callback
	OnWeight(item any, priority int)
}

type PriorityQ struct {
	Q           // 继承 Q
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	waitingHeap *heap // 基于对象的权重堆
	cb          PriorityCallback
}

// 创建一个 PriorityQueue 对象
// Create a new PriorityQueue object.
func NewPriorityQueue(cb PriorityCallback) *PriorityQ {
	return newPriorityQ("", cb)
}

// 创建一个带名称的 PriorityQueue 对象
// Create a new named PriorityQueue object.
func NewNamedPriorityQueue(name string, cb PriorityCallback) *PriorityQ {
	return newPriorityQ(name, cb)
}

func newPriorityQ(name string, cb PriorityCallback) *PriorityQ {
	q := &PriorityQ{
		Q:           *newQ(name, cb),
		wg:          sync.WaitGroup{},
		waitingHeap: &heap{data: make([]*waitingFor, 0, defaultQueueCap)},
		cb:          cb,
	}
	q.ctx, q.cancel = context.WithCancel(context.Background())
	q.wg.Add(1)
	go q.waitingLoop()
	return q
}

// 添加一个带权重的任务到队列中
// Add a weighted task to the queue
func (q *PriorityQ) AddWeight(item interface{}, priority int) {
	if q.ShuttingDown() {
		return
	}
	q.cb.OnWeight(item, priority)
	if priority <= 0 {
		q.Add(item)
		return
	}
	q.waitingHeap.Push(&waitingFor{
		data:  item,
		value: int64(priority),
	})
}

// / 关闭 Queue
// Close the queue
func (q *PriorityQ) ShutDown() {
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
func (q *PriorityQ) ShutDownWithDrain() {
	q.once.Do(func() {
		q.cond.L.Lock()
		defer q.cond.L.Unlock()
		q.drain = true
		q.shutdown()
		for q.processing.len() > 0 && q.drain {
			q.cond.Wait()
		}
		q.cancel()
		q.wg.Wait()
		q.waitingHeap.Reset()
	})
}

// 从堆中读取 WaitingFor，如果对象没有超时，就重新放回堆
// read from heap, if the object has not timed out, put it back in the heap
func (q *PriorityQ) waitingLoop() {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			entry := q.waitingHeap.Pop()
			if entry == nil {
				continue
			}
			q.Add(entry.data)
		}
	}
}
