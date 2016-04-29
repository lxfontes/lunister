package main

import (
	"errors"
	"sync"
)

var (
	ErrCapacityReached = errors.New("Reached bucket capacity")
	ErrQueueClosed     = errors.New("Queue is shutting down")
)

type LeakyBucket struct {
	sync.Mutex
	waterMark int
	capacity  int
	queue     chan interface{}
}

func NewLeakyBucket(capacity int) *LeakyBucket {
	return &LeakyBucket{
		capacity: capacity,
		queue:    make(chan interface{}, capacity),
	}
}

func (b *LeakyBucket) Push(v interface{}) error {
	if b.waterMark >= b.capacity {
		return ErrCapacityReached
	}

	b.Lock()
	b.waterMark += 1
	b.Unlock()

	b.queue <- v

	return nil
}

func (b *LeakyBucket) Pop() (interface{}, error) {
	v := <-b.queue

	b.Lock()
	b.waterMark -= 1
	b.Unlock()
	return v, nil
}

func (b *LeakyBucket) Queue() <-chan interface{} {
	return b.queue
}
