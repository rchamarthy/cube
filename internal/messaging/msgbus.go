package messaging

import (
	"errors"
	"time"
)

// Msgbus message bus
type Msgbus interface {
	// register the bus
	register(string) (Msgbus, error)
	// unregister the bus
	unregister() error
	// register a msg handler
	registerMsgHandler(string, func([]byte, bool) ([]byte, error)) error
	// unregister a msg handler
	unregisterMsgHandler(string) error
	// send a byte payload to a named receiver
	send([]byte, string) error
	// send a byte payload to a named receiver and wait for response within a timeout
	sendAndWaitResponse([]byte, string, string, time.Duration) ([]byte, error)
}

var (
	// ErrUnSupp unsupported bus
	ErrUnSupp = errors.New("msgbus: unsupported type")
	// ErrBadBus invalid msgbus
	ErrBadBus = errors.New("msgbus: invalid msgbus")
	// ErrTimeout timeout
	ErrTimeout = errors.New("msgbus: timeout")
	// ErrNotInited bus is not initialized
	ErrNotInited = errors.New("msgbus: bus is not initialized")
	// ErrDupConn duplicate registration
	ErrDupConn = errors.New("msgus: duplicate registration")
	// ErrBadAppResp bad app response
	ErrBadAppResp = errors.New("msgbus: bad app response")
	// ErrSubNotFound subscription not found
	ErrSubNotFound = errors.New("msgbus: subscription not found")
	// ErrDupSub duplication subscription
	ErrDupSub = errors.New("msgbus: duplicate subscription")
	// ErrSub subscription error
	ErrSub = errors.New("msgbus: subscription error")
	// ErrConn connection failed
	ErrConn = errors.New("msgbus: connection failed")
	// ErrSend send error
	ErrSend = errors.New("msgbug: send error")
)

const defaultTimeout = time.Millisecond * 10

var defaultMsgbus Msgbus

// RegisterMsgbus initialize and register bus
func RegisterMsgbus(busType MsgbusType, uri string) Msgbus {
	// NOTE: idempotent
	if defaultMsgbus != nil {
		return defaultMsgbus
	}

	var mb Msgbus
	var err error

	switch busType {
	case Mock:
		var mmb mockMsgbus
		mb, err = mmb.register(uri)
		if err != nil {
			panic(ErrBadBus)
		}
	default:
		panic(ErrUnSupp)
	}
	if mb == nil {
		panic(ErrBadBus)
	}
	defaultMsgbus = mb
	return defaultMsgbus
}

// UnregisterMsgbus unregister bus
func UnregisterMsgbus(mb Msgbus) error {
	if defaultMsgbus == nil {
		return ErrNotInited
	}
	defaultMsgbus.unregister()
	defaultMsgbus = nil
	return nil
}
