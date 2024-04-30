package frame

import (
	"fmt"

	"github.com/nohns/eventale/internal"
	"google.golang.org/protobuf/proto"
)

const (
	_uint32Len = 4
)

type FrameKind uint32

const (
	FrameKindHeartbeat FrameKind = iota + 1
	FrameKindSecretPublish
	FrameKindClientHello
	FrameKindServerHello
	_FrameKindLast
)

func (fk FrameKind) validate() error {
	if fk < FrameKindHeartbeat || fk >= _FrameKindLast {
		return fmt.Errorf("invalid frame kind %d", fk)
	}
	return nil
}

type Frame struct {
	Kind       FrameKind
	ID         internal.ID
	RespondsTo internal.ID
	Payload    []byte
}

type makeOpt func(*Frame) error

func WithProto(msg proto.Message) makeOpt {
	return makeOpt(func(frm *Frame) error {
		b, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		frm.Payload = b
		return nil
	})
}

// WithID takes in an IDer on which GenID is called to generate an ID for the frame
func WithID(ider internal.IDer) makeOpt {
	return makeOpt(func(f *Frame) error {
		id, err := ider.GenID()
		if err != nil {
			return err
		}
		f.ID = id
		return nil
	})
}

func WithRespondTo(id internal.ID) makeOpt {
	return makeOpt(func(f *Frame) error {
		f.RespondsTo = id
		return nil
	})
}

func Make(kind FrameKind, mkopts ...makeOpt) (*Frame, error) {
	frm := &Frame{Kind: kind}
	for _, f := range mkopts {
		if err := f(frm); err != nil {
			return nil, err
		}
	}
	return frm, nil
}

type Framer struct {
	ider internal.IDer
}
