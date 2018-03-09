package chat

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecodeMessage(t *testing.T) {
	var str = `{"content": "1234"}`
	var msg Message

	err := json.NewDecoder(strings.NewReader(str)).Decode(&msg)
	assert.NoError(t, err)

	t.Logf("%v: %#v", str, msg)

	ch := make(chan int)
	go func() {
		ch <- 123
		for {
			i, ok := <-ch
			t.Logf("%v, %v <- ch\n", i, ok)
			if !ok {
				break
			}
		}
	}()

	time.Sleep(5 * time.Second)
	t.Logf("closing channel ...\n")
	close(ch)
}
