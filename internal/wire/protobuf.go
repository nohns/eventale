package wire

import (
	"context"

	"github.com/nohns/eventale/internal/frame"
	"google.golang.org/protobuf/proto"
)

type unaryConn interface {
	Send(context.Context, *frame.Frame) error
	Recv(context.Context) (*frame.Frame, error)
}

/*func CallUnary[Req proto.Message, Res proto.Message](conn unaryConn, req Req, res Res) error {
	b, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	eventalepb.File_v1_tcp_proto.Messages()
	req.ProtoReflect().Descriptor().Index()
	frm, err := frame.Make(frame.FrameKind(req.ProtoReflect().Descriptor().Index()))
}*/

func FrameUnmarshal[T proto.Message](data []byte, msg T) {
	proto.Unmarshal(data, msg)
}
