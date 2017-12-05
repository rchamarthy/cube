package msgbus

import (
	"errors"
	"sync"
	"time"

	"github.com/anuvu/cube/component"
	"github.com/anuvu/cube/config"
	"github.com/anuvu/cube/internal/messaging"
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

// configuration defines the configurable parameters of http server
type configuration struct {
	config.BaseConfig
	// Listen port
	MsgbusType messaging.MsgbusType `json:"msgbus_type"`
	MsgbusURI  string               `json:"msgbus_uri"`
}

// Msgbus is a message bus
type Msgbus interface {
	RegisterMsgHandler(target string, msgHandler func([]byte, bool) ([]byte, error)) error
	UnregisterMsgHandler(target string) error
	Send(data []byte, target string) error
	SendAndWaitResponse(data []byte, target string, timeout time.Duration) ([]byte, uuid.UUID, error)
	// NOTE: the following are only needed to manually setup the msgbus
	Register() error
	Unregister() error
}

type msgbus struct {
	config  *configuration
	mb      messaging.Msgbus
	running bool
	lock    *sync.RWMutex
}

// New returns a new msgbus
func New() Msgbus {
	cfg := &configuration{
		BaseConfig: config.BaseConfig{ConfigKey: "msgbus"},
	}

	return &msgbus{config: cfg, running: false, lock: &sync.RWMutex{}}
}

func (m *msgbus) Register() error {
	if m == nil {
		return ErrInvalid
	}

	m.mb = messaging.RegisterMsgbus(m.config.MsgbusType, m.config.MsgbusURI)
	return nil
}

func (m *msgbus) Unregister() error {
	if m == nil {
		return ErrNotInited
	}

	if m.mb == nil {
		return ErrNotInited
	}

	err := messaging.UnregisterMsgbus(m.mb)
	m.mb = nil
	return err
}

func (m *msgbus) RegisterMsgHandler(target string, msgHandler func([]byte, bool) ([]byte, error)) error {
	if m == nil {
		return ErrNotInited
	}
	if m.mb == nil {
		return ErrNotInited
	}
	if target == "" {
		return ErrBadSub
	}
	return messaging.RegisterMsgHandler(target, msgHandler)
}

func (m *msgbus) UnregisterMsgHandler(target string) error {
	if m == nil {
		return ErrNotInited
	}
	if m.mb == nil {
		return ErrNotInited
	}
	if target == "" {
		return ErrBadSub
	}
	return messaging.UnregisterMsgHandler(target)
}

func (m *msgbus) Send(data []byte, target string) error {
	if m == nil {
		return ErrNotInited
	}
	if m.mb == nil {
		return ErrNotInited
	}
	if target == "" {
		return ErrBadSub
	}
	if len(data) == 0 {
		return ErrBadPayload
	}
	return messaging.Send(data, target)
}

func (m *msgbus) SendAndWaitResponse(data []byte, target string, timeout time.Duration) ([]byte, uuid.UUID, error) {
	if m == nil {
		return nil, uuid.Nil, ErrNotInited
	}
	if m.mb == nil {
		return nil, uuid.Nil, ErrNotInited
	}
	if target == "" {
		return nil, uuid.Nil, ErrBadSub
	}
	if len(data) == 0 {
		return nil, uuid.Nil, ErrBadPayload
	}
	return messaging.SendAndWaitResponse(data, target, timeout)
}

func (m *msgbus) Config() config.Config {
	return m.config
}

func (m *msgbus) Configure(ctx component.Context) error {
	return nil
}

func (m *msgbus) Start(ctx component.Context) error {
	err := m.Register()
	if err == nil {
		m.lock.Lock()
		defer m.lock.Unlock()
		m.running = true
	}
	return err
}

func (m *msgbus) Stop(ctx component.Context) error {
	err := m.Unregister()
	m.lock.Lock()
	defer m.lock.Unlock()
	m.running = false
	return err
}

func (m *msgbus) IsHealthy(ctx component.Context) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.running
}
