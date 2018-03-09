package std

import (
	"sync/atomic"
	"time"

	uuid "github.com/satori/go.uuid"
)

var (
	nextUID uint64
)

// GenUniqueID generate unique id
func GenUniqueID() uint64 {
	return atomic.AddUint64(&nextUID, 1)
}

// GenUIDs generate unique id string
func GenUIDs() string {
	id := uuid.Must(uuid.NewV4())
	return id.String()
}

// GetNowMs return UTC time since 1/1/1970 in Millisecond
func GetNowMs() int64 {
	t := time.Now().UTC()
	return t.Unix()*1000 + int64(t.Nanosecond())/int64(time.Millisecond)
}
