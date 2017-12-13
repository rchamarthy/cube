package msgbus

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMsgBadCalls(t *testing.T) {
	var m *Msg

	assert.Panics(t, func() { m.MakeReply([]byte("")) }, "should panic")
}

func BenchmarkNewMsg(b *testing.B) {
	msg := newMsg([]byte("hello world"))
	if msg == nil {
		panic(ErrNewMsg)
	}
}

func BenchmarkGenVerifyHash(b *testing.B) {
	rand.Seed(int64(time.Now().Nanosecond()))
	l := rand.Intn(1024)
	buf := make([]byte, l)
	fmt.Printf("Benchmarking iter=%d len=%d\n", b.N, len(buf))
	for i := 0; i < b.N; i++ {
		msg := newMsg([]byte(buf))
		if msg == nil {
			panic(ErrNewMsg)
		}
		msg.GenerateHash()
		msg.VerifyHash()
	}
}
