package messaging

import (
	"bytes"
	"errors"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

// MsgbusType message bus type
type MsgbusType int

const (
	// Mock mock null broker
	Mock MsgbusType = iota
	// RabbitMQ rabbitMQ broker
	RabbitMQ MsgbusType = iota
	// Nats NATS broker
	Nats MsgbusType = iota
)

// Register the messaging bus
func Register(busType MsgbusType, uri string) (Msgbus, error) {
	mb := RegisterMsgbus(busType, uri)
	return mb, nil
}

// Unregister the messaging bus
func Unregister(mb Msgbus) error {
	err := UnregisterMsgbus(mb)
	return err
}

// RegisterMsgHandler register a msg handler
func RegisterMsgHandler(target string, msgHandler func([]byte, bool) ([]byte, error)) error {
	if defaultMsgbus == nil {
		return ErrNotInited
	}
	return defaultMsgbus.registerMsgHandler(target, msgHandler)
}

// UnregisterMsgHandler unregister a msg handler
func UnregisterMsgHandler(target string) error {
	if defaultMsgbus == nil {
		return ErrNotInited
	}
	return defaultMsgbus.unregisterMsgHandler(target)
}

// Send some bytes to a target
func Send(data []byte, target string) error {
	if defaultMsgbus == nil {
		return ErrNotInited
	}

	msg := newMsg(data)
	msg.generateHash()
	d, e := proto.Marshal(msg)
	if e != nil {
		panic("unable to marshal msg")
	}
	e = defaultMsgbus.send(d, target)
	return e
}

// SendAndWaitResponse send a byte payload to a named receiver and wait for response
func SendAndWaitResponse(data []byte, target string, timeout time.Duration) ([]byte, uuid.UUID, error) {
	if defaultMsgbus == nil {
		return nil, uuid.Nil, errors.New("bus is not initialized")
	}
	msg := newMsg(data)
	msg.Flags |= msgFlagsRespExpected
	msg.generateHash()
	d, err := proto.Marshal(msg)
	if err != nil {
		panic("unable to marshal msg")
	}
	u, err := uuid.FromBytes(msg.GetHandle())
	if err != nil {
		panic("invalid handle")
	}
	r, err := defaultMsgbus.sendAndWaitResponse(d, target, u.String(), timeout)
	if err != nil {
		return nil, u, err
	}
	m, err := unmarshal(r)
	if err != nil {
		panic("unable to unmarshal msg")
	}
	m.verifyHash()
	if !bytes.Equal(m.GetHandle(), msg.GetHandle()) {
		panic("request response mismatch")
	}
	return m.GetPayload(), u, err
}
