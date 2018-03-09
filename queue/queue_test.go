package queue

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testItem []byte

func (t testItem) Len() int {
	return len(t)
}

func TestByteQueueResize(t *testing.T) {
	initialCapacity := 100
	q := New(initialCapacity, false)
	assert.Equal(t, 0, q.Len())
	assert.Equal(t, initialCapacity, q.Cap())
	assert.Equal(t, false, q.Closed())

	for i := 0; i < initialCapacity; i++ {
		q.Add(testItem([]byte(strconv.Itoa(i))))
	}
	assert.Equal(t, initialCapacity, q.Cap())
	q.Add(testItem([]byte("resize here")))
	assert.Equal(t, initialCapacity*2, q.Cap())
	q.Remove()
	// back to initial capacity
	assert.Equal(t, initialCapacity, q.Cap())

	q.Add(testItem([]byte("new resize here")))
	assert.Equal(t, initialCapacity*2, q.Cap())
	q.Add(testItem([]byte("one more item, no resize must happen")))
	assert.Equal(t, initialCapacity*2, q.Cap())

	assert.Equal(t, initialCapacity+2, q.Len())
}

func TestByteQueueSize(t *testing.T) {
	initialCapacity := 100
	q := New(initialCapacity, false)
	assert.Equal(t, 0, q.Size())
	q.Add(testItem([]byte("1")))
	q.Add(testItem([]byte("2")))
	assert.Equal(t, 2, q.Size())
	q.Remove()
	assert.Equal(t, 1, q.Size())
}

func TestByteQueueWait(t *testing.T) {
	initialCapacity := 100
	q := New(initialCapacity, false)
	q.Add(testItem([]byte("1")))
	q.Add(testItem([]byte("2")))

	s, ok := q.Wait()
	assert.Equal(t, true, ok)
	assert.Equal(t, "1", string(s.(testItem)))

	s, ok = q.Wait()
	assert.Equal(t, true, ok)
	assert.Equal(t, "2", string(s.(testItem)))

	go func() {
		q.Add(testItem([]byte("3")))
	}()

	s, ok = q.Wait()
	assert.Equal(t, true, ok)
	assert.Equal(t, "3", string(s.(testItem)))

}

func TestByteQueueClose(t *testing.T) {
	initialCapacity := 100
	q := New(initialCapacity, false)

	// test removing from empty queue
	_, ok := q.Remove()
	assert.Equal(t, false, ok)

	q.Add(testItem([]byte("1")))
	q.Add(testItem([]byte("2")))
	q.Close()

	ok = q.Add(testItem([]byte("3")))
	assert.Equal(t, false, ok)

	_, ok = q.Wait()
	assert.Equal(t, false, ok)

	_, ok = q.Remove()
	assert.Equal(t, false, ok)

	assert.Equal(t, true, q.Closed())

}

func TestByteQueueCloseRemaining(t *testing.T) {
	q := New(100, false)
	q.Add(testItem([]byte("1")))
	q.Add(testItem([]byte("2")))
	msgs := q.CloseRemaining()
	assert.Equal(t, 2, len(msgs))
	ok := q.Add(testItem([]byte("3")))
	assert.Equal(t, false, ok)
	assert.Equal(t, true, q.Closed())
	msgs = q.CloseRemaining()
	assert.Equal(t, 0, len(msgs))
}

func BenchmarkQueueAdd(b *testing.B) {
	q := New(2, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Add(testItem([]byte("test")))
	}
	b.StopTimer()
	q.Close()
}

func addAndConsume(q Queue, n int, bc bool) {
	// Add to queue and consume in another goroutine.
	done := make(chan struct{})
	go func() {
		count := 0
		for {
			var ok bool
			if bc {
				select {
				case _, ok = <-q.Cout():
					break
				}
			} else {
				_, ok = q.Wait()
			}
			if !ok {
				continue
			}

			count++
			if count == n {
				close(done)
				break
			}
		}
	}()
	for i := 0; i < n; i++ {
		q.Add(testItem([]byte("test")))
	}
	<-done
}

func BenchmarkQueueAddConsume(b *testing.B) {
	bc := false
	q := New(2, bc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addAndConsume(q, 100000, bc)
	}
	b.StopTimer()
	q.Close()
}

func BenchmarkQueueAddConsumeCout(b *testing.B) {
	bc := true
	q := New(2, bc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addAndConsume(q, 10000, bc)
	}
	b.StopTimer()
	q.Close()
}

//BenchmarkCQueueAddConsume-8   	  30	  49766746 ns/op	  841107 B/op	  100006 allocs/op
//BenchmarkCQueueAddConsume-8   	  20	  53149885 ns/op	 5334676 B/op	  105834 allocs/op
//BenchmarkQueueAddConsume-8   	      30	  42166713 ns/op	10448140 B/op	  230483 allocs/op
//BenchmarkQueue3AddConsume-8   	   30	  60566676 ns/op	  800197 B/op	  100000 allocs/op
//BenchmarkQueue4AddConsume-8   	    50	  25420024 ns/op	 1021612 B/op	  100036 allocs/op
