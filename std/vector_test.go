package std

import (
	"testing"

	"github.com/go-clog/clog"
	"github.com/stretchr/testify/assert"
)

func init() {
	clog.New(clog.CONSOLE, clog.ConsoleConfig{
		Level:      clog.TRACE,
		BufferSize: 1024,
	})
}

func TestNewVector(t *testing.T) {
	cap := 32
	vec := NewVector(cap)
	assert.Equal(t, vec.Len(), 0)
	assert.Equal(t, vec.Cap(), cap)

	for i := 0; i < cap; i++ {
		vec.Push(i)
		assert.Equal(t, vec.Len(), i+1)
	}

	assert.Equal(t, vec.Len(), cap)
	assert.Equal(t, vec.Cap(), cap)

	for i := 0; i < cap; i++ {
		assert.EqualValues(t, i, vec.At(i))
	}
}

func TestVectorPush(t *testing.T) {
	cap := 32
	vec := NewVector(cap)

	for i := 0; i < cap; i++ {
		vec.Push(i)
	}

	for i := 0; i < cap; i++ {
		assert.EqualValues(t, i, vec.At(i))
	}

	vec.Clear()
	assert.Equal(t, nil, vec.At(0))
}

func TestVectorResize(t *testing.T) {
	v := NewVector(5)
	v.Push(1)
	v.Push(2)
	assert.Equal(t, 5, v.Cap())

	v.Resize(10)
	assert.Equal(t, 10, v.Cap())
	assert.EqualValues(t, 1, v.At(0))
}

func TestVectorPop(t *testing.T) {
	cap := 32
	vec := NewVector(cap)
	for i := 0; i < cap; i++ {
		vec.Push(i)
		assert.Equal(t, vec.Len(), i+1)
	}
	for i := 0; i < cap; i++ {
		v, ok := vec.Pop()
		assert.Equal(t, ok, true)
		assert.EqualValues(t, v, i)
	}

	v, ok := vec.Pop()
	assert.Nil(t, v)
	assert.Equal(t, ok, false)

	v, ok = vec.Pop()
	assert.Nil(t, v)
	assert.Equal(t, ok, false)
}

func BenchmarkVectorPush(b *testing.B) {
	cap := 1024
	v := NewVector(cap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Push(i)
	}
}

func BenchmarkVectorPop(b *testing.B) {
	cap := 1024
	v := NewVector(cap)
	for i := 0; i < b.N; i++ {
		v.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Pop()
	}
}
