package eventale

import (
	"context"
	"fmt"
	"net"
	"time"

	eventalepb "github.com/nohns/eventale/gen/v1"
	"github.com/nohns/eventale/internal/transport"
	"google.golang.org/protobuf/proto"
)

var _networkTimeout = 30 * time.Second

type frameDecoder interface {
	Decode(ctx context.Context) (*transport.Frame, error)
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

	// Send hello and receive server hello
	b, err := proto.Marshal(&eventalepb.WireClientHello{})
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(b)
	if err != nil {
		return nil, err
	}
	frm, err := transport.NewDecoder(conn).Decode(opts.ctx)
	if err != nil {
		return nil, fmt.Errorf("client dial: %v", err)
	}
	if frm.Kind != transport.FrameKindServerHello {
		return nil, fmt.Errorf("unexpected frame kind %d after client hello", frm.Kind)
	}
	var srvhello eventalepb.WireServerHello
	if err := proto.Unmarshal(frm.Payload, &srvhello); err != nil {
		return nil, fmt.Errorf("client dial: %v", err)
	}

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
