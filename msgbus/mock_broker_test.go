package msgbus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockBrokerBadCalls(t *testing.T) {
	testMsg := []byte("Mr. Watson--come here--I want to see you.")
	target1 := "foo"
	var nmb *mockBroker
	bmb := &mockBroker{}

	assert.NotNil(t, nmb.Register())
	assert.NotNil(t, nmb.Unregister())
	assert.NotNil(t, nmb.RegisterMsgHandler(target1, msgListener))
	assert.NotNil(t, nmb.UnregisterMsgHandler(target1))
	assert.NotNil(t, nmb.Send([]byte(testMsg), target1))
	_, e := nmb.SendAndWaitResponse([]byte(testMsg), target1, "", 0)
	assert.NotNil(t, e)
	timeout := time.Duration(10 * time.Microsecond) // no msgbus can do this
	_, e = bmb.SendAndWaitResponse([]byte(testMsg), target1, "", timeout)
	assert.NotNil(t, e)
}
