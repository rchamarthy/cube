package msgbus

import (
	"bytes"
	"crypto/sha1"

	"encoding/binary"
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

var (
	// ErrNewMsg failed to create a new msg
	ErrNewMsg = errors.New("msg: unable to create a msg")
	// ErrMarshal marshaling error
	ErrMarshal = errors.New("msg: unable to marshal")
	// ErrUnmarshal unmarshaling error
	ErrUnmarshal = errors.New("msg: unable to unmarshal")
	// ErrHash hashing error
	ErrHash = errors.New("msg: failed to verify hash")
)

type msgFlags int32

const (
	// MsgFlagsEmpty no flags
	MsgFlagsEmpty msgFlags = 0
	// MsgFlagsRespExpected indicates a response is needed for this msg
	MsgFlagsRespExpected = 1 << iota
	// MsgFlagsMax is a sentinel
	MsgFlagsMax
)

func newMsg(payload []byte) *Msg {
	msg := new(Msg)
	msg.Payload = payload
	msg.Handle = uuid.NewV4().Bytes()
	return msg
}

// MakeReply a reply msg using called msg
func (msg *Msg) MakeReply(payload []byte) *Msg {
	if msg == nil {
		panic("invalid msg")
	}
	r := new(Msg)
	*r = *msg
	r.Payload = payload
	r.GenerateHash()
	return r
}

// Marshal a msg into a byte array
func Marshal(msg *Msg) ([]byte, error) {
	d, err := proto.Marshal(msg)
	return d, err
}

// Unmarshal a byte array into a msg
func Unmarshal(d []byte) (*Msg, error) {
	m := &Msg{}
	err := proto.Unmarshal(d, m)
	return m, err
}

// GenerateHash creates an hash of the msg
func (msg *Msg) GenerateHash() {
	h := sha1.New()
	h.Write(msg.Payload)
	h.Write(msg.Handle)
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, msg.Flags)
	if err != nil {
		panic(err)
	}
	h.Write(buf.Bytes())
	hash := h.Sum(nil)
	msg.Hash = hash
	//msg.dump()
}

// VerifyHash verifies the msg hash
func (msg *Msg) VerifyHash() {
	//msg.dump()
	h := sha1.New()
	h.Write(msg.Payload)
	h.Write(msg.Handle)
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, msg.Flags)
	if err != nil {
		panic(err)
	}
	h.Write(buf.Bytes())
	hash := h.Sum(nil)
	if !bytes.Equal(hash, msg.GetHash()) {
		panic("invalid hash")
	}
}
