package workqueue

import (
	"context"
	"sync"
	"time"
)

const DeafultQueueSortWindows = 500 * time.Millisecond

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
	Q            // 继承 Q
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	priorityHeap *heap // 基于对象的权重堆
	cb           PriorityCallback
	lockHeap     sync.Mutex
}

// 创建一个 PriorityQueue 对象
// Create a new PriorityQueue object.
func NewPriorityQueue(win time.Duration, cb PriorityCallback) *PriorityQ {
	return newPriorityQ("", win, cb)
}

// 创建一个带名称的 PriorityQueue 对象
// Create a new named PriorityQueue object.
func NewNamedPriorityQueue(name string, win time.Duration, cb PriorityCallback) *PriorityQ {
	return newPriorityQ(name, win, cb)
}

func newPriorityQ(name string, win time.Duration, cb PriorityCallback) *PriorityQ {
	if cb == nil {
		cb = emptycb{}
	}
	if win <= DeafultQueueSortWindows {
		win = DeafultQueueSortWindows
	}
	q := &PriorityQ{
		Q:            *newQ(name, cb),
		wg:           sync.WaitGroup{},
		priorityHeap: &heap{data: make([]*waitingFor, 0, defaultQueueCap)},
		cb:           cb,
		lockHeap:     sync.Mutex{},
	}
	q.ctx, q.cancel = context.WithCancel(context.Background())
	q.wg.Add(1)
	go q.waitingLoop()
	return q
}

// 添加一个带权重的任务到队列中
// Add a weighted task to the queue
func (q *PriorityQ) AddWeight(item any, priority int) {
	if q.ShuttingDown() {
		return
	}
	q.cb.OnWeight(item, priority)
	if priority <= 0 {
		q.Add(item)
		return
	}
	q.lockHeap.Lock()
	q.priorityHeap.Push(&waitingFor{
		data:  item,
		value: int64(priority),
	})
	q.lockHeap.Unlock()
}

// / 关闭 Queue
// Close the queue
func (q *PriorityQ) ShutDown() {
	q.Q.ShutDown()
	q.cancel()
	q.wg.Wait()
	q.lockHeap.Lock()
	q.priorityHeap.Reset()
	q.lockHeap.Unlock()

}

// 关闭 Queue 并且等待所有的任务都被处理完
// Close the Queue and wait for all tasks to be processed
func (q *PriorityQ) ShutDownWithDrain() {
	q.Q.ShutDownWithDrain()
	q.cancel()
	q.wg.Wait()
	q.lockHeap.Lock()
	q.priorityHeap.Reset()
	q.lockHeap.Unlock()
}

// 从堆中读取已经在事件窗口中排序好的 WaitingFor, 放入 Q 中
// Read the WaitingFor that has been sorted in the event window from the heap and put it into Q
func (q *PriorityQ) waitingLoop() {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			for {
				q.lockHeap.Lock()
				entry := q.priorityHeap.Pop()
				q.lockHeap.Unlock()
				if entry == nil {
					break
				}
				q.Add(entry.data)
			}
		}
	}
}
