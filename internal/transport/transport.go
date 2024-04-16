package transport

import (
	"fmt"

	eventalepb "github.com/nohns/eventale/gen/v1"
	"google.golang.org/protobuf/proto"
)

const (
	_uint32Len = 4
)

type FrameKind uint32

const (
	FrameKindHeartbeat FrameKind = iota + 1
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
	Kind    FrameKind
	Payload []byte
}

// From proto msg takes a kind and a proto message and converts it into a
// frame.
func FromProtoMsg(kind FrameKind, msg proto.Message) (*Frame, error) {
	frm := &Frame{Kind: kind, Payload: nil}
	if msg == nil {
		return frm, nil
	}
	b, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	frm.Payload = b
	return frm, nil
}

func SemVer(semverpb *eventalepb.SemanticVersion) string {
	return fmt.Sprintf("v%d.%d.%d", semverpb.Major, semverpb.Minor, semverpb.Patch)
}
