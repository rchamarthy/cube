package messaging

import (
	"fmt"
	"time"

	"github.com/satori/go.uuid"
)

type mockMsgbus struct {
	subMap map[string]func([]byte, bool) ([]byte, error)
}

func newMockMsgBus(uri string) Msgbus {
	fmt.Printf("mock: connected\n")
	mmb := &mockMsgbus{subMap: make(map[string]func([]byte, bool) ([]byte, error))}
	return mmb
}

func (mmb *mockMsgbus) register(uri string) (Msgbus, error) {
	mb := newMockMsgBus(uri)
	return mb, nil
}

func (mmb *mockMsgbus) unregister() error {
	if mmb == nil {
		return ErrNotInited
	}
	return nil
}

func (mmb *mockMsgbus) registerMsgHandler(target string, msgHandler func([]byte, bool) ([]byte, error)) error {
	if mmb == nil {
		return ErrNotInited
	}

	// FIXME: catch some errors!
	if mmb.subMap[target] != nil {
		return ErrDupSub
	}

	mmb.subMap[target] = msgHandler
	fmt.Printf("mock: subscribed for [%s]\n", target)

	return nil
}

func (mmb *mockMsgbus) unregisterMsgHandler(target string) error {
	if mmb == nil {
		return ErrNotInited
	}
	ch := mmb.subMap[target]
	if ch == nil {
		return ErrSubNotFound
	}
	delete(mmb.subMap, target)
	fmt.Printf("mock: unsubscribed for [%s]\n", target)
	return nil
}

func (mmb *mockMsgbus) send(data []byte, target string) error {
	if mmb == nil {
		return ErrNotInited
	}

	msgHandler := mmb.subMap[target]
	if msgHandler != nil {
		msg, err := unmarshal(data)
		if err != nil {
			panic(err)
		}
		msg.verifyHash()
		// call the handler
		msgHandler(msg.GetPayload(), false)
	}

	fmt.Printf("mock: sent %d bytes to target [%s]\n", len(data), target)
	return nil
}

func (mmb *mockMsgbus) sendAndWaitResponse(data []byte, target string, handle string, timeout time.Duration) ([]byte, error) {
	if mmb == nil {
		return nil, ErrNotInited
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
		msg, err := unmarshal(data)
		if err != nil {
			panic(err)
		}
		msg.verifyHash()
		// call the registered testMsgHandler
		// FIXME: make this more robust by timing out on stuck msgHandler()
		fmt.Printf("mock: received a message of %d bytes\n", len(msg.GetPayload()))
		fmt.Printf("mock: received flags %d\n", msg.GetFlags())
		respExpected := msg.GetFlags()&msgFlagsRespExpected == msgFlagsRespExpected
		r, err := msgHandler(msg.GetPayload(), respExpected)
		if err != nil {
			panic(err)
		}
		// if a response was expected for this msg, a response was expected of the callback!
		if respExpected {
			if len(r) == 0 {
				panic("invalid response")
			}
			rsp := msg.makeReply(r)
			u1, err := uuid.FromBytes(msg.GetHandle())
			if err != nil {
				panic("invalid msg handle")
			}
			d, err := marshal(rsp)
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
