package std

import (
	"sync"
)

// Queue interface
//
type Queue interface {
	Cap() int
	Len() int

	// Close the queue, and signal all wait operation
	Close()

	// Closed check if queue has been closed.
	Closed() bool

	// Add element into queue without wait,
	// return false, if the queue closed.
	Add(value interface{}) bool

	// Remove element from queue without wait,
	// return <nil, false>, if there is no element or queue closed.
	Remove() (interface{}, bool)

	// Wait get element from queue synchronized,
	// return <nil, false> only if the queue closed.
	Wait() (interface{}, bool)

	// Cout get element from output channel
	Cout() <-chan interface{}
}

type syncQueue struct {
	lck    sync.RWMutex
	cod    *sync.Cond
	buf    Vector
	cout   chan interface{}
	cap    int
	closed bool
}

// NewSyncQueue create an new synchronized queue with given initial capacity,
// and it will automatically expand and shrink the queue capacity.
//
func NewSyncQueue(initCap int) Queue {
	q := &syncQueue{
		buf: NewVector(initCap),
		cap: initCap,
	}
	q.cod = sync.NewCond(&q.lck)
	return q
}

func (q *syncQueue) Cout() <-chan interface{} {
	q.lck.Lock()
	if q.cout == nil {
		q.cout = make(chan interface{}, 128)
		go func() {
			defer func() {
				close(q.cout)
				q.cout = nil
			}()
			for {
				if v, _ := q.Wait(); v != nil {
					q.cout <- v
				} else {
					break
				}
			}
		}()
	}
	q.lck.Unlock()
	return q.cout
}

func (q *syncQueue) Add(value interface{}) bool {
	q.lck.Lock()
	if q.closed {
		q.lck.Unlock()
		return false
	}
	q.buf.Push(value)
	q.cod.Signal()
	q.lck.Unlock()
	return true
}

func (q *syncQueue) Remove() (interface{}, bool) {
	q.lck.Lock()
	if q.closed {
		q.lck.Unlock()
		return nil, false
	}
	v, ok := q.buf.Pop()
	q.lck.Unlock()
	return v, ok
}

func (q *syncQueue) Wait() (interface{}, bool) {
	q.lck.Lock()
	if q.buf.Len() > 0 {
		q.lck.Unlock()
		return q.Remove()
	}

	q.cod.Wait()
	q.lck.Unlock()
	return q.Remove()
}

func (q *syncQueue) Close() {
	q.lck.Lock()
	q.closed = true
	q.cod.Broadcast()
	q.lck.Unlock()
}

func (q *syncQueue) Closed() bool {
	q.lck.RLock()
	c := q.closed
	q.lck.RUnlock()
	return c
}

func (q *syncQueue) Cap() int {
	q.lck.RLock()
	c := q.buf.Cap()
	q.lck.RUnlock()
	return c
}

func (q *syncQueue) Len() int {
	q.lck.RLock()
	l := q.buf.Len()
	q.lck.RUnlock()
	return l
}
