package messaging

import (
	"fmt"
	"testing"

	"math/rand"
	"time"

	"github.com/stretchr/testify/assert"
)

var mockURI = "mock://uri.is.ignored"

func BenchmarkNewMsg(b *testing.B) {
	newMsg([]byte("hello world"))
}

const request = "all well?"
const reply = "ca va!"

func testAllHarnesses(t *testing.T, testFunc func(*testing.T)) {
	mb, err := Register(Mock, mockURI)
	assert.Nil(t, err)

	testFunc(t)

	assert.Nil(t, Unregister(mb))
}

func testMsgHandler(data []byte, respExpected bool) ([]byte, error) {
	fmt.Printf("received [len=%d,data=%s]\n", len(data), data)
	if respExpected {
		fmt.Printf("response is expected\n")
		var i int
		var x string
		fmt.Sscanf(fmt.Sprintf("%s", data), "[%d]%s", &i, &x)
		// return the response byte payload
		r := []byte(fmt.Sprintf("[%d]%s", i, reply))
		fmt.Printf("sending response [len=%d, data=%s]\n", len(r), r)
		return r, nil
	}
	fmt.Printf("response is not expected\n")
	return nil, nil
}

func testSubUnsub(t *testing.T) {
	assert.Nil(t, RegisterMsgHandler("match", testMsgHandler))
	assert.Nil(t, UnregisterMsgHandler("match"))
}

func TestSubUnsub(t *testing.T) {
	testAllHarnesses(t, testSubUnsub)
}

func doTestSend(t *testing.T, i int) {
	data := []byte(fmt.Sprintf("[%d]%s", i, request))
	assert.Nil(t, Send(data, "match"))
}

func testSend(t *testing.T) {
	assert.Nil(t, RegisterMsgHandler("match", testMsgHandler))
	doTestSend(t, 1)
	assert.Nil(t, UnregisterMsgHandler("match"))
}

func TestSend(t *testing.T) {
	testAllHarnesses(t, testSend)
}

func doTestSendAndWait(i int, timeout time.Duration) error {
	start := time.Now()
	defer func() { fmt.Printf("[rtt]%s\n", time.Now().Sub(start)) }()
	data := []byte(fmt.Sprintf("[%d]%s", i, request))
	r, _, err := SendAndWaitResponse(data, "match", timeout)
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}
	fmt.Printf("received reply [len=%d, data=%s]\n", len(r), r)
	return nil
}

func testSendAndWait(t *testing.T) {
	assert.NotNil(t, doTestSendAndWait(1, time.Duration(100*time.Microsecond)))
	assert.Nil(t, RegisterMsgHandler("match", testMsgHandler))
	assert.NotNil(t, RegisterMsgHandler("match", testMsgHandler))
	assert.Nil(t, doTestSendAndWait(1, 0))
	assert.NotNil(t, doTestSendAndWait(1, time.Duration(100*time.Microsecond)))
	assert.Nil(t, UnregisterMsgHandler("match"))
	assert.NotNil(t, UnregisterMsgHandler("match"))
}

func TestSendAndWait(t *testing.T) {
	testAllHarnesses(t, testSendAndWait)
}

func testSendAndWaitTimeout(t *testing.T) {
	err := RegisterMsgHandler("match", testMsgHandler)
	if err != nil {
		panic(err)
	}
	err = doTestSendAndWait(1, time.Duration(time.Microsecond*100))
	if err == nil {
		panic("should have timed out")
	}
	err = UnregisterMsgHandler("match")
	if err != nil {
		panic(err)
	}
}

func TestSendAndWaitTimeout(t *testing.T) {
	testAllHarnesses(t, testSendAndWait)
}

func TestMockNegative1(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	_, err := Register(100, "")
	assert.NotNil(t, err)
}

func TestMockNegative2(t *testing.T) {
	assert.NotNil(t, RegisterMsgHandler("match", testMsgHandler))
	assert.NotNil(t, UnregisterMsgHandler("match"))
	assert.NotNil(t, Send([]byte(""), "match"))
	_, _, err := SendAndWaitResponse([]byte(""), "match", 0)
	assert.NotNil(t, err)
	var nmb *mockMsgbus
	assert.NotNil(t, nmb.unregister())
	assert.NotNil(t, nmb.registerMsgHandler("", testMsgHandler))
	assert.NotNil(t, nmb.unregisterMsgHandler(""))
	assert.NotNil(t, nmb.send([]byte(""), "match"))
	_, err = nmb.sendAndWaitResponse([]byte(""), "match", "handle", 0)
	assert.NotNil(t, err)
}

func BenchmarkGenVerifyHash(b *testing.B) {
	rand.Seed(int64(time.Now().Nanosecond()))
	l := rand.Intn(1024)
	buf := make([]byte, l)
	fmt.Printf("Benchmarking iter=%d len=%d\n", b.N, len(buf))
	for i := 0; i < b.N; i++ {
		msg := newMsg([]byte(buf))
		msg.generateHash()
		msg.verifyHash()
	}
}

func benchAllHarnesses(b *testing.B, testFunc func(*testing.B)) {
	mb, err := Register(Mock, mockURI)
	if err != nil {
		panic(err)
	}
	testFunc(b)
	err = Unregister(mb)
	if err != nil {
		panic(err)
	}
}

func benchmarkSendAndWait(b *testing.B) {
	err := RegisterMsgHandler("match", testMsgHandler)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 100; i++ {
		doTestSendAndWait(i, 0)
	}
	err = UnregisterMsgHandler("match")
	if err != nil {
		panic(err)
	}
}

func BenchmarkSendAndWait(b *testing.B) {
	benchAllHarnesses(b, benchmarkSendAndWait)
}
