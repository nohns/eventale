package eventale

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	eventalepb "github.com/nohns/eventale/gen/v1"
	"github.com/nohns/eventale/internal/connection"
	"github.com/nohns/eventale/internal/frame"
	"github.com/nohns/eventale/internal/wire"
	"google.golang.org/protobuf/proto"
)

var _networkTimeout = 30 * time.Second

type frameDecoder interface {
	Decode(ctx context.Context) (*frame.Frame, error)
}

type Client struct {
	conn  net.Conn
	frmer frameDecoder
}

func WithContext(ctx context.Context) dialOpt {
	return dialOptFunc(func(opts *dialOpts) {
		if ctx == nil {
			return
		}
		opts.ctx = ctx
	})
}

func Dial(address string, options ...dialOpt) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), _networkTimeout)
	defer cancel()

	opts := dialOpts{
		ctx: ctx,
	}
	for _, opt := range options {
		opt.apply(&opts)
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	c := &connection.Conn{
		NetConn: conn,
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	// Send hello and receive server hello
	fmt.Print("send client hello\n")
	frm, err := frame.Make(frame.FrameKindClientHello, frame.WithProto(&eventalepb.WireClientHello{
		ClientVersion: &eventalepb.SemanticVersion{
			Major: 0,
			Minor: 0,
			Patch: 1,
		},
	}))
	if err != nil {
		return nil, err
	}
	if err := c.Send(context.Background(), frm); err != nil {
		return nil, err
	}

	frm, err = frame.NewDecoder(conn).Decode()
	if err != nil {
		return nil, fmt.Errorf("client dial: %v", err)
	}
	if frm.Kind != frame.FrameKindServerHello {
		return nil, fmt.Errorf("unexpected frame kind %d after client hello", frm.Kind)
	}
	var srvhello eventalepb.WireServerHello
	if err := proto.Unmarshal(frm.Payload, &srvhello); err != nil {
		return nil, fmt.Errorf("client dial: %v", err)
	}
	fmt.Printf("recv server hello - %s\n", wire.SemVerStr(srvhello.ServerVersion))

	return &Client{
		conn: conn,
	}, nil
}

type dialOpts struct {
	ctx context.Context
}

type dialOpt interface {
	apply(opts *dialOpts)
}

type dialOptFunc func(opts *dialOpts)

func (f dialOptFunc) apply(opts *dialOpts) {
	f(opts)
}
