package std

// Vector golang version std::vector from C++
//
type Vector interface {
	Cap() int
	Len() int

	// Clear all element in the vetor
	Clear()

	// Resize vector's capacity to given size
	Resize(n int)

	// Push an element to the end of the vector
	Push(value interface{})

	// Pop the first element in the vector,
	// return false when the vector is empty.
	Pop() (interface{}, bool)

	// At return the element at given index.
	At(index int) interface{}
}

type vector struct {
	node       []interface{}
	cap, cnt   int
	head, tail int
	initCap    int
}

//
// NewVector create vector which is not thread safety, caller should guarantee this.
// The capacity will automatically expend at 2x or shrink by 1/2 util to initial capacity.
func NewVector(initialCap int) Vector {
	if initialCap < 2 {
		initialCap = 2
	}
	return &vector{
		initCap: initialCap,
		cap:     initialCap,
		node:    make([]interface{}, initialCap),
	}
}

// Size return the current size of vector
func (v *vector) Len() int {
	return v.cnt
}

// Cap return the capacity of vector
func (v *vector) Cap() int {
	return v.cap
}

// Resize capacity of vector to given size
func (v *vector) Resize(n int) {
	node := make([]interface{}, n)
	//fmt.Printf("cap: %d, len: %d\n", v.cap, v.cnt)
	//fmt.Printf("resize %d => %d\n", v.cap, n)

	if v.head < v.tail {
		copy(node, v.node[v.head:v.tail])
	} else {
		copy(node[:], v.node[v.head:])
		copy(node[v.cap-v.head:], v.node[:v.tail])
	}

	v.cap, v.head = n, 0
	v.tail = v.cnt % n
	v.node = node[:]
}

// Push an new value into the end of vector
func (v *vector) Push(value interface{}) {
	if v.cnt == v.cap {
		v.Resize(v.cap * 2)
	}
	v.node[v.tail] = value
	v.tail = (v.tail + 1) % v.cap
	v.cnt++
}

// Pop removet the first value of the vector
// <nil, false> means the vector is empty.
func (v *vector) Pop() (interface{}, bool) {
	if v.cnt == 0 {
		//clog.Info("Vector : %v", v)
		return nil, false
	}

	node := v.node[v.head]
	v.node[v.head] = nil
	v.head = (v.head + 1) % v.cap
	v.cnt--

	if v.cnt < v.cap/2 && v.cnt > v.initCap {
		v.Resize(v.cap / 2)
	}
	return node, true
}

// At access the value of vector at given index
func (v *vector) At(index int) interface{} {
	index = (v.head + index) % v.cap
	return v.node[index]
}

// Clear vector content
func (v *vector) Clear() {
	for idx := range v.node {
		v.node[idx] = nil
	}
	//v.node = v.node[:0]
	v.head, v.tail = 0, 0
	v.cnt = 0
}
