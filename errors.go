package workqueue

import "errors"

var (
	// RingArray
	ErrRingArrayInvalidArgument = errors.New("[workqueue] RingArray: Invalid argument")
	ErrRingArrayOutOfRange      = errors.New("[workqueue] RingArray: Out of range")
	// Queue
	ErrorQueueClosed    = errors.New("[workqueue] Queue: Is closed")
	ErrorQueueFull      = errors.New("[workqueue] Queue: Is full")
	ErrorQueueEmpty     = errors.New("[workqueue] Queue: Queue is empty")
	ErrorQueueItemExist = errors.New("[workqueue] Queue: Item already exists in queue")
)
