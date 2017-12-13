package msgbus

import (
	"bytes"
	"errors"
	"sync"
	"time"

	"github.com/anuvu/cube/component"
	"github.com/anuvu/cube/config"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

var (
	// ErrNotInited not initialized
	ErrNotInited = errors.New("msgbus: not initialized")
	// ErrInvalid invalid
	ErrInvalid = errors.New("msgbus: invalid")
	// ErrBadSub bad subscription
	ErrBadSub = errors.New("msgbus: bad subscription")
	// ErrBadPayload bad data payload
	ErrBadPayload = errors.New("msgbus: bad payload")
)

// Configuration defines the configurable parameters of http server
type Configuration struct {
	config.BaseConfig
	// Listen port
	MsgbusType string `json:"msgbus_type"`
	MsgbusURI  string `json:"msgbus_uri"`
}

// Msgbus is a message bus
type Msgbus interface {
	RegisterMsgHandler(target string, msgHandler func([]byte, bool) ([]byte, error)) error
	UnregisterMsgHandler(target string) error
	Send(data []byte, target string) error
	SendAndWaitResponse(data []byte, target string, timeout time.Duration) ([]byte, uuid.UUID, error)
}

type msgbus struct {
	config  *Configuration
	broker  Broker
	running bool
	lock    *sync.RWMutex
}

// New returns a new msgbus
func New() Msgbus {
	cfg := &Configuration{
		BaseConfig: config.BaseConfig{ConfigKey: "msgbus"},
	}

	return &msgbus{config: cfg, running: false, lock: &sync.RWMutex{}}
}

func (mb *msgbus) RegisterMsgHandler(target string, msgHandler func([]byte, bool) ([]byte, error)) error {
	if mb == nil {
		return ErrNotInited
	}
	if mb.broker == nil {
		return ErrNotInited
	}
	if target == "" {
		return ErrBadSub
	}
	return mb.broker.RegisterMsgHandler(target, msgHandler)
}

func (mb *msgbus) UnregisterMsgHandler(target string) error {
	if mb == nil {
		return ErrNotInited
	}
	if mb.broker == nil {
		return ErrNotInited
	}
	if target == "" {
		return ErrBadSub
	}
	return mb.broker.UnregisterMsgHandler(target)
}

func (mb *msgbus) Send(data []byte, target string) error {
	if mb == nil {
		return ErrNotInited
	}
	if mb.broker == nil {
		return ErrNotInited
	}
	if target == "" {
		return ErrBadSub
	}
	if len(data) == 0 {
		return ErrBadPayload
	}

	msg := newMsg(data)
	msg.GenerateHash()
	d, e := proto.Marshal(msg)
	if e != nil {
		panic("unable to marshal msg")
	}

	return mb.broker.Send(d, target)
}

func (mb *msgbus) SendAndWaitResponse(data []byte, target string, timeout time.Duration) ([]byte, uuid.UUID, error) {
	if mb == nil {
		return nil, uuid.Nil, ErrNotInited
	}
	if mb.broker == nil {
		return nil, uuid.Nil, ErrNotInited
	}
	if target == "" {
		return nil, uuid.Nil, ErrBadSub
	}
	if len(data) == 0 {
		return nil, uuid.Nil, ErrBadPayload
	}

	msg := newMsg(data)
	msg.Flags |= MsgFlagsRespExpected
	msg.GenerateHash()
	d, err := proto.Marshal(msg)
	if err != nil {
		panic("unable to marshal msg")
	}
	u, err := uuid.FromBytes(msg.GetHandle())
	if err != nil {
		panic("invalid handle")
	}
	r, err := mb.broker.SendAndWaitResponse(d, target, u.String(), timeout)
	if err != nil {
		return nil, u, err
	}
	m, err := Unmarshal(r)
	if err != nil {
		panic("unable to unmarshal msg")
	}
	m.VerifyHash()
	if !bytes.Equal(m.GetHandle(), msg.GetHandle()) {
		panic("request response mismatch")
	}
	return m.GetPayload(), u, err
}

func (mb *msgbus) Config() config.Config {
	return mb.config
}

func (mb *msgbus) Configure(ctx component.Context) error {
	return nil
}

func (mb *msgbus) Start(ctx component.Context) error {
	// instantiate the broker based on configuration
	b, err := NewBroker(mb.config)
	if err != nil {
		return err
	}
	mb.broker = b
	err = mb.broker.Register()
	if err == nil {
		mb.lock.Lock()
		defer mb.lock.Unlock()
		mb.running = true
	}
	return err
}

func (mb *msgbus) Stop(ctx component.Context) error {
	err := mb.broker.Unregister()
	mb.lock.Lock()
	defer mb.lock.Unlock()
	mb.running = false
	return err
}

func (mb *msgbus) IsHealthy(ctx component.Context) bool {
	mb.lock.RLock()
	defer mb.lock.RUnlock()
	return mb.running
}
