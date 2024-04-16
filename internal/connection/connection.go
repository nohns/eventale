package connection

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"github.com/nohns/eventale/internal/transport"
)

var ErrConnectionTimeout = errors.New("connection timeout")

type FrameHandler interface {
	Handle(conn *Conn, frm *transport.Frame) error
}

type Conn struct {
	ID      int
	NetConn net.Conn
	Logger  *slog.Logger

	dec *transport.FrameDecoder
	enc *transport.FrameEncoder
}

func (tc *Conn) Send(ctx context.Context, frm *transport.Frame) error {
	// Run in goroutine, so we can return on timeout, or encode result
	errc := make(chan error)
	go func() {
		defer close(errc)
		errc <- tc.encoder().Encode(frm)
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (tc *Conn) Recv(ctx context.Context) (*transport.Frame, error) {
	// Run in goroutine, so we can return on timeout, or decode result
	frmc := make(chan *transport.Frame)
	errc := make(chan error)
	go func() {
		defer close(frmc)
		defer close(errc)
		frm, err := tc.decoder().Decode()
		if err != nil {
			errc <- err
			return
		}
		frmc <- frm
	}()

	select {
	case frm := <-frmc:
		return frm, nil
	case err := <-errc:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (tc *Conn) Close() error {
	if err := tc.NetConn.Close(); err != nil {
		return err
	}
	return nil
}

func (tc *Conn) decoder() *transport.FrameDecoder {
	if tc.dec == nil {
		tc.dec = transport.NewDecoder(tc.NetConn)
	}
	return tc.dec
}

func (tc *Conn) encoder() *transport.FrameEncoder {
	if tc.enc == nil {
		tc.enc = transport.NewEncoder(tc.NetConn)
	}
	return tc.enc
}
