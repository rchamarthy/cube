package msgbus

import (
	"errors"
	"fmt"
	"time"
)

// Broker is a message bus broker interface
type Broker interface {
	// Register the bus
	Register() error
	// Unregister the bus
	Unregister() error
	// RegisterMsgHandler registers a listening msg handler
	RegisterMsgHandler(string, func([]byte, bool) ([]byte, error)) error
	// UnregisterMsgHandler unregister listening a msg handler
	UnregisterMsgHandler(string) error
	// Send a byte payload to a named receiver
	Send([]byte, string) error
	// SendAndWaitResponse sends a byte payload to a named receiver and wait for
	// response within a timeout
	SendAndWaitResponse([]byte, string, string, time.Duration) ([]byte, error)
}

var (
	// ErrBadConfig bad configuration
	ErrBadConfig = errors.New("broker: bad configuration")
	// ErrUnSupp unsupported bus
	ErrUnSupp = errors.New("broker: unsupported type")
	// ErrBadBroker invalid broker
	ErrBadBroker = errors.New("broker: invalid broker")
	// ErrTimeout timeout
	ErrTimeout = errors.New("broker: timeout")
	// ErrDupBroker duplicate broker
	ErrDupBroker = errors.New("broker: duplicate broker registration")
	// ErrDupConn duplicate registration
	ErrDupConn = errors.New("broker: duplicate registration")
	// ErrBadAppResp bad app response
	ErrBadAppResp = errors.New("broker: bad app response")
	// ErrSubNotFound subscription not found
	ErrSubNotFound = errors.New("broker: subscription not found")
	// ErrDupSub duplication subscription
	ErrDupSub = errors.New("broker: duplicate subscription")
	// ErrSub subscription error
	ErrSub = errors.New("broker: subscription error")
	// ErrConn connection failed
	ErrConn = errors.New("broker: connection failed")
	// ErrSend send error
	ErrSend = errors.New("msgbug: send error")
)

const defaultTimeout = time.Millisecond * 10

// BrokerFactory is a broker factory
type BrokerFactory func(config *Configuration) (Broker, error)

var brokerFactories = make(map[string]BrokerFactory)

// RegisterFactory registers a broker factory
func RegisterFactory(name string, factory BrokerFactory) {

	if name == "" {
		panic(ErrBadBroker)
	}

	if factory == nil {
		panic(ErrBadBroker)
	}

	_, found := brokerFactories[name]
	if found {
		panic(ErrDupConn)
	}

	brokerFactories[name] = factory
	fmt.Printf("msgbus: registered factory - %s\n", name)
}

// NewBroker creates a broker instance
func NewBroker(config *Configuration) (Broker, error) {
	if config == nil {
		panic(ErrBadConfig)
	}
	factory, ok := brokerFactories[config.MsgbusType]
	if !ok {
		return nil, ErrBadBroker
	}

	return factory(config)
}
