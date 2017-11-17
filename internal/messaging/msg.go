package messaging

import (
	"bytes"
	"crypto/sha1"

	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

var (
	// ErrMarshal marshaling error
	ErrMarshal = errors.New("msg: unable to marshal")
	// ErrUnmarshal unmarshaling error
	ErrUnmarshal = errors.New("msg: unable to unmarshal")
	// ErrHash hashing error
	ErrHash = errors.New("msg: failed to verify hash")
)

type msgFlags int32

const (
	msgFlagsEmpty        msgFlags = 0
	msgFlagsRespExpected          = 1 << iota
	msgFlagsMax
)

func newMsg(payload []byte) *Msg {
	msg := new(Msg)
	msg.Payload = payload
	msg.Handle = uuid.NewV4().Bytes()
	return msg
}

// make a reply msg using called msg
func (msg *Msg) makeReply(payload []byte) *Msg {
	if msg == nil {
		panic("invalid msg")
	}
	r := new(Msg)
	*r = *msg
	r.Payload = payload
	r.generateHash()
	return r
}

func marshal(msg *Msg) ([]byte, error) {
	d, err := proto.Marshal(msg)
	return d, err
}

func unmarshal(d []byte) (*Msg, error) {
	m := &Msg{}
	err := proto.Unmarshal(d, m)
	return m, err
}

// msg integrity routines
func (msg *Msg) generateHash() {
	h := sha1.New()
	h.Write(msg.Payload)
	h.Write(msg.Handle)
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, msg.Flags)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		panic(err)
	}
	h.Write(buf.Bytes())
	hash := h.Sum(nil)
	msg.Hash = hash
	//msg.dump()
}

func (msg *Msg) verifyHash() {
	//msg.dump()
	h := sha1.New()
	h.Write(msg.Payload)
	h.Write(msg.Handle)
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, msg.Flags)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	h.Write(buf.Bytes())
	hash := h.Sum(nil)
	if !bytes.Equal(hash, msg.GetHash()) {
		panic("invalid hash")
	}
}

func (msg *Msg) dump() {
	fmt.Printf("Msg[%s|%s|%d|%s]\n", msg.Payload, hex.EncodeToString(msg.Handle), msg.Flags, hex.EncodeToString(msg.Hash))
}
