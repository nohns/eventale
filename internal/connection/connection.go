package connection

import (
	"context"
	"crypto/aes"
	"errors"
	"log/slog"
	"net"
	"sync"

	"github.com/nohns/eventale/internal/frame"
)

var ErrConnectionTimeout = errors.New("connection timeout")

type FrameHandler interface {
	Handle(conn *Conn, frm *frame.Frame) error
}

type Conn struct {
	ID      int
	NetConn net.Conn
	Logger  *slog.Logger

	enckey []byte
	mu     sync.RWMutex
	dec    *frame.FrameDecoder
	enc    *frame.FrameEncoder
}

func (tc *Conn) Send(ctx context.Context, frm *frame.Frame) error {
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

func (tc *Conn) Unary(ctx context.Context, reqfrm *frame.Frame) (resfrm *frame.Frame, err error) {
	if err := tc.Send(ctx, reqfrm); err != nil {
		return nil, err
	}
	resfrm, err = tc.Recv(ctx)
	if err != nil {
		return nil, err
	}
	return resfrm, nil
}

func (tc *Conn) Recv(ctx context.Context) (*frame.Frame, error) {
	// Run in goroutine, so we can return on timeout, or decode result
	frmc := make(chan *frame.Frame)
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

// Upgrade enabling encryption on communication
func (tc *Conn) Upgrade(key []byte) {
	b, err := aes.NewCipher(key)
}

func (tc *Conn) Close() error {
	if err := tc.NetConn.Close(); err != nil {
		return err
	}
	return nil
}

func (tc *Conn) decoder() *frame.FrameDecoder {
	if tc.dec == nil {
		tc.dec = frame.NewDecoder(tc.NetConn)
	}
	return tc.dec
}

func (tc *Conn) encoder() *frame.FrameEncoder {
	if tc.enc == nil {
		tc.enc = frame.NewEncoder(tc.NetConn)
	}
	return tc.enc
}
