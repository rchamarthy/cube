package msgbus

import (
	"fmt"
	"time"

	"github.com/satori/go.uuid"
)

type mockBroker struct {
	subMap map[string]func([]byte, bool) ([]byte, error)
}

func newMockBroker(config *Configuration) (Broker, error) {
	mmb := &mockBroker{subMap: make(map[string]func([]byte, bool) ([]byte, error))}
	return mmb, nil
}

func (mmb *mockBroker) Register() error {
	if mmb == nil {
		return ErrBadBroker
	}
	fmt.Printf("mock: connected\n")
	return nil
}

func (mmb *mockBroker) Unregister() error {
	if mmb == nil {
		return ErrBadBroker
	}
	fmt.Printf("mock: disconnected\n")
	return nil
}

func (mmb *mockBroker) RegisterMsgHandler(target string, msgHandler func([]byte, bool) ([]byte, error)) error {
	if mmb == nil {
		return ErrBadBroker
	}

	// FIXME: catch some errors!
	if mmb.subMap[target] != nil {
		return ErrDupSub
	}

	mmb.subMap[target] = msgHandler
	fmt.Printf("mock: subscribed for [%s]\n", target)

	return nil
}

func (mmb *mockBroker) UnregisterMsgHandler(target string) error {
	if mmb == nil {
		return ErrBadBroker
	}
	ch := mmb.subMap[target]
	if ch == nil {
		return ErrDupSub
	}
	delete(mmb.subMap, target)
	fmt.Printf("mock: unsubscribed for [%s]\n", target)
	return nil
}

func (mmb *mockBroker) Send(data []byte, target string) error {
	if mmb == nil {
		return ErrBadBroker
	}

	msgHandler := mmb.subMap[target]
	if msgHandler != nil {
		msg, err := Unmarshal(data)
		if err != nil {
			panic(err)
		}
		msg.VerifyHash()
		// call the handler
		msgHandler(msg.GetPayload(), false)
	}

	fmt.Printf("mock: sent %d bytes to target [%s]\n", len(data), target)
	return nil
}

func (mmb *mockBroker) SendAndWaitResponse(data []byte, target string, handle string, timeout time.Duration) ([]byte, error) {
	if mmb == nil {
		return nil, ErrBadBroker
	}

	msgHandler := mmb.subMap[target]
	if msgHandler == nil {
		if timeout > 0 {
			// just simulate a timeout
			time.Sleep(timeout)
		}
		return nil, ErrTimeout
	}

	// timeout chan, because timeout can be zero/disabled
	tch := make(chan struct{}, 1)
	if timeout > 0 {
		go func(timeout time.Duration) {
			time.Sleep(timeout)
			tch <- struct{}{}
			close(tch)
		}(timeout)
	}

	// FIXME: simulate a delay
	time.Sleep(1 * time.Millisecond)

	// this is a request-response msg, so subscribe a target on this particular msg handle
	rch := make(chan []byte, 1)
	go func() {
		msg, err := Unmarshal(data)
		if err != nil {
			panic(err)
		}
		msg.VerifyHash()
		// call the registered testMsgHandler
		// FIXME: make this more robust by timing out on stuck msgHandler()
		fmt.Printf("mock: received a message of %d bytes\n", len(msg.GetPayload()))
		fmt.Printf("mock: received flags %d\n", msg.GetFlags())
		respExpected := msg.GetFlags()&MsgFlagsRespExpected == MsgFlagsRespExpected
		r, err := msgHandler(msg.GetPayload(), respExpected)
		if err != nil {
			panic(err)
		}
		// if a response was expected for this msg, a response was expected of the callback!
		if respExpected {
			if len(r) == 0 {
				panic("invalid response")
			}
			rsp := msg.MakeReply(r)
			u1, err := uuid.FromBytes(msg.GetHandle())
			if err != nil {
				panic("invalid msg handle")
			}
			d, err := Marshal(rsp)
			if err != nil {
				panic(err)
			}

			fmt.Printf("mock: sending reply to [%s]\n", u1.String())
			rch <- d
			close(rch)
		}
	}()

	// now, block and wait for response
	var r []byte
	var err error
	select {
	case r = <-rch:
		break
	case <-tch:
		r = nil
		err = ErrTimeout
		break
	}

	return r, err
}

func init() {
	RegisterFactory("mock", newMockBroker)
}
