package std

import (
	"fmt"
	"runtime/debug"
	"sync"
	"testing"
)

func addAndConsume4(q Queue, total int, round int) {
	// Add to queue and consume in another goroutine.
	var wg sync.WaitGroup
	wg.Add(2)
	//clog.Info("Round: %v, len: %v, cap: %v", round, q.Len(), q.Cap())

	go func() {
		defer wg.Done()
		for i := 0; i < total; i++ {
			if v, ok := q.Wait(); !ok {
				//clog.Info("Round: %d|%d, except: %v, %v", round, i, v, ok)
				break
			} else if v != i {
				//clog.Error(2, "Round %d: invalid %v : %v, len: %v, cap: %v", round, v, i, q.Len(), q.Cap())
			}
		}
		//clog.Info("Round: %d, len: %v, cap: %v", round, q.Len(), q.Cap())
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < total; i++ {
			q.Add(i)
		}
	}()

	wg.Wait()
}

func addAndConsume5(q Queue, total int, round int) {
	// Add to queue and consume in another goroutine.
	var wg sync.WaitGroup
	wg.Add(2)
	//clog.Info("Round: %v, len: %v, cap: %v", round, q.Len(), q.Cap())

	go func() {
		defer wg.Done()
		n := total
		c := q.Cout()

		for n > 0 {
			if _, ok := <-c; !ok {
				break
			} else {
				n--
			}
		}
		//clog.Info("Round: %d, len: %v, cap: %v", round, q.Len(), q.Cap())
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < total; i++ {
			q.Add(i)
		}
	}()

	wg.Wait()
}

func BenchmarkQueue4AddConsume(b *testing.B) {
	q := NewSyncQueue(100)
	b.ResetTimer()

	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			fmt.Printf("panic: %v", err)
		}
	}()

	for i := 0; i < b.N; i++ {
		//clog.Info("Round %d", i)
		addAndConsume5(q, 100000, i)
	}

	b.StopTimer()
	q.Close()
}
