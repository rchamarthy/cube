package msgbus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrokerBadCalls(t *testing.T) {
	assert.Panics(t, func() { NewBroker(nil) }, "should panic")
	var config = &Configuration{MsgbusType: "bad"}
	_, err := NewBroker(config)
	assert.NotNil(t, err)
	assert.Panics(t, func() { RegisterFactory("", func(config *Configuration) (Broker, error) { return nil, nil }) }, "should panic")
	assert.Panics(t, func() { RegisterFactory("bad", nil) }, "should panic")
	RegisterFactory("bad", func(config *Configuration) (Broker, error) { return nil, nil })
	assert.Panics(t, func() { RegisterFactory("bad", func(config *Configuration) (Broker, error) { return nil, nil }) }, "should panic")
}
