package msgbus

import (
	"testing"
	"time"

	"bytes"
	"fmt"
	"os"

	"github.com/anuvu/cube/component"
	"github.com/anuvu/cube/internal/messaging"
	"github.com/stretchr/testify/assert"
)

var msgChkCh chan []byte

func msgListener(inMsg []byte, respReqd bool) ([]byte, error) {
	fmt.Printf("msgListener: received msg %s\n", inMsg)
	msgChkCh <- inMsg
	if respReqd {
		fmt.Printf("msgListener: reply sent %s\n", inMsg)
		return inMsg, nil
	}
	return nil, nil
}

func checkRcvdMsgIs(sent []byte) bool {
	timeout := make(chan struct{}, 1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		timeout <- struct{}{}
	}()
	select {
	case rcvd := <-msgChkCh:
		fmt.Printf("sent:%s rcvd:%s\n", sent, rcvd)
		// a read from ch has occurred
		if bytes.Equal(sent, rcvd) {
			return true
		}
	case <-timeout:
		// the read from ch has timed out
	}
	return false
}

// expects msgbus setup-teardown happening from the caller
func testAPISequence(t *testing.T, mb Msgbus) {
	// setup the feedback channel
	msgChkCh = make(chan []byte, 1)

	testMsg := []byte("Mr. Watson--come here--I want to see you.")
	target1 := "foo"
	var nmb *msgbus

	// register a listener
	fmt.Printf("Register a listener\n")
	assert.NotNil(t, nmb.RegisterMsgHandler(target1, msgListener))
	assert.NotNil(t, mb.RegisterMsgHandler("", msgListener))
	assert.Nil(t, mb.RegisterMsgHandler(target1, msgListener))
	assert.NotNil(t, mb.RegisterMsgHandler(target1, msgListener))

	// send a msg to first target
	fmt.Printf("Fire-and-forget to first target\n")
	assert.NotNil(t, nmb.Send([]byte(testMsg), target1))
	assert.NotNil(t, mb.Send([]byte(testMsg), ""))
	assert.NotNil(t, mb.Send(nil, target1))
	assert.Nil(t, mb.Send([]byte(testMsg), target1))
	assert.True(t, checkRcvdMsgIs(testMsg))

	// send a blocking (forever) send-receive to first target
	fmt.Printf("Blocking send and wait forever\n")
	_, _, e := nmb.SendAndWaitResponse([]byte(testMsg), target1, 0)
	assert.NotNil(t, e)
	_, _, e = mb.SendAndWaitResponse([]byte(testMsg), "", 0)
	assert.NotNil(t, e)
	_, _, e = mb.SendAndWaitResponse(nil, target1, 0)
	assert.NotNil(t, e)
	r, _, e := mb.SendAndWaitResponse([]byte(testMsg), target1, 0)
	assert.True(t, checkRcvdMsgIs(testMsg))
	assert.Nil(t, e)
	assert.True(t, bytes.Equal(testMsg, r))

	// send a blocking (with timeout) send-receive to first target
	fmt.Printf("Blocking send and long timeout\n")
	timeout := time.Duration(1 * time.Second)
	r, _, e = mb.SendAndWaitResponse([]byte(testMsg), target1, timeout)
	assert.True(t, checkRcvdMsgIs(testMsg))
	assert.Nil(t, e) // if there is a timeout, this will be non-nil
	assert.True(t, bytes.Equal(testMsg, r))

	// send a blocking (with timeout) send-receive to first target but cause timeout
	fmt.Printf("Blocking send but cause timeout\n")
	timeout = time.Duration(10 * time.Microsecond) // no msgbus can do this
	_, _, e = mb.SendAndWaitResponse([]byte(testMsg), target1, timeout)
	assert.True(t, checkRcvdMsgIs(testMsg))
	assert.Equal(t, e, messaging.ErrTimeout) // if there is a timeout, this will be non-nil

	// send to second target not received
	fmt.Printf("Send to a non-listening target\n")
	target2 := "bar"
	assert.Nil(t, mb.Send([]byte(testMsg), target2))
	assert.False(t, checkRcvdMsgIs(testMsg))

	// unregister the first listener
	fmt.Printf("Unregister first listener\n")
	assert.NotNil(t, nmb.UnregisterMsgHandler(target1))
	assert.NotNil(t, mb.UnregisterMsgHandler(""))
	assert.NotNil(t, mb.UnregisterMsgHandler("unknown"))
	assert.Nil(t, mb.UnregisterMsgHandler(target1))

	// send a msg to first target which should not be received
	fmt.Printf("Send to first target\n")
	assert.Nil(t, mb.Send([]byte(testMsg), target1))
	assert.False(t, checkRcvdMsgIs(testMsg))
}

func TestMsgbusAPI(t *testing.T) {
	// create a new bus instance
	fmt.Printf("Creating a new bus instance\n")
	mb := New()
	assert.NotNil(t, mb)

	// register it to remote broker
	fmt.Printf("Register the msgbus\n")
	var nmb *msgbus
	assert.NotNil(t, nmb.Register())
	assert.Nil(t, mb.Register())

	testAPISequence(t, mb)

	// unregister the bus
	fmt.Printf("Unregister the msgbus\n")
	assert.Nil(t, mb.Unregister())

	testMsg := []byte("Mr. Watson--come here--I want to see you.")
	target1 := "foo"
	assert.NotNil(t, nmb.RegisterMsgHandler(target1, msgListener))
	assert.NotNil(t, mb.RegisterMsgHandler(target1, msgListener))
	assert.NotNil(t, mb.UnregisterMsgHandler(target1))
	assert.NotNil(t, mb.Send(testMsg, target1))
	_, _, err := mb.SendAndWaitResponse(testMsg, target1, 0)
	assert.NotNil(t, err)

	assert.NotNil(t, mb.Unregister())
	assert.NotNil(t, nmb.Unregister())
}

func TestGroupIntegration(t *testing.T) {
	// Replace os.Args for test case
	oldArgs := os.Args
	os.Args = []string{"msgbus_test", "-config.mem", `{"msgbus": {"msgbus_type": 0}}`}
	defer func() { os.Args = oldArgs }()

	// new group
	fmt.Printf("Create a new group\n")
	grp := component.New("msgbus_test")
	assert.Nil(t, grp.Add(New))
	assert.Nil(t, grp.Create())

	// pick up "testing" configuration
	assert.Nil(t, grp.Configure())

	grp.Invoke(func(mb Msgbus) {
		// start it
		fmt.Printf("Starting group\n")
		assert.Nil(t, grp.Start())

		assert.True(t, grp.IsHealthy())

		testAPISequence(t, mb)

		assert.True(t, grp.IsHealthy())

		// stop it
		fmt.Printf("Stopping group\n")
		assert.Nil(t, grp.Stop())

		assert.False(t, grp.IsHealthy())
	})
}
