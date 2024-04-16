package eventale

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	eventalepb "github.com/nohns/eventale/gen/v1"
	"github.com/nohns/eventale/internal/connection"
	"github.com/nohns/eventale/internal/transport"
)

const (
	// The time period before closing a connection to a client due to timeout.
	_serverConnTimeout = 30 * time.Second
)

var (
	// ErrServerClosed is returned from ListenAndServe() at some point after
	// someone called the Close() method.
	ErrServerClosed      = errors.New("server closed")
	ErrTest              = errors.New("test")
	ErrConnectionTimeout = errors.New("connection timeout")
)

func ListenAndServe(address string) error {
	srv := NewServer(address)
	return srv.ListenAndServe()
}

type serverStatus int

const (
	serverStatusIdle serverStatus = iota
	serverStatusServing
	serverStatusClosed
)

type Server struct {
	// Addr is the address on where the server is served.
	Addr string
	// Logger is the structured logger used when writing to stdout
	Logger *slog.Logger

	lnr    net.Listener
	conns  []*connection.Conn
	state  serverStatus
	mu     sync.Mutex
	nextid int
}

func NewServer(addr string) *Server {
	return &Server{
		Addr:   addr,
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		conns:  make([]*connection.Conn, 0),
		nextid: 1,
	}
}

func (s *Server) ListenAndServe() error {
	s.Logger.Info("Listening for traffic", slog.String("addr", s.Addr))
	lnr, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer s.Close()

	s.mu.Lock()
	s.state = serverStatusServing
	s.mu.Unlock()

	for {
		s.Logger.Debug("Waiting for connection...")
		conn, err := lnr.Accept()
		if errors.Is(err, io.EOF) {
			return ErrServerClosed
		}
		if err != nil {
			if s.readState() == serverStatusClosed {
				return ErrServerClosed
			}
			return err
		}

		s.Logger.Debug("Connecting to client")
		s.mu.Lock()
		c := &connection.Conn{
			ID:      s.nextid,
			NetConn: conn,
			Logger:  s.Logger,
		}
		s.conns = append(s.conns, c)
		s.mu.Unlock()

		go s.listenOnConn(c)
		s.Logger.Debug("client connected", slog.Int("id", c.ID))
	}
}

func (s *Server) Close() error {
	if err := s.lnr.Close(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = serverStatusClosed

	for _, conn := range s.conns {
		if err := conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) listenOnConn(conn *connection.Conn) {
	defer conn.Close()
	for {
		ctx, cancel := context.WithTimeoutCause(context.Background(), _serverConnTimeout, ErrConnectionTimeout)
		frm, err := conn.Recv(ctx)
		cancel()
		if errors.Is(err, ErrConnectionTimeout) {
			s.Logger.Info("Connection timeout")
			return
		}
		if errors.Is(err, io.EOF) {
			s.Logger.Info(fmt.Sprintf("Quit listen on connection %d - EOF", conn.ID))
			return
		}
		if err != nil {
			s.Logger.Error("Failed to read frame", slog.String("error", err.Error()))
			continue
		}
		if err := s.handleFrame(conn, frm); err != nil {
			s.Logger.Error("Failed to handle frame", slog.String("error", err.Error()))
		}
	}
}

func (s *Server) handleFrame(conn *connection.Conn, frm *transport.Frame) error {
	switch frm.Kind {
	case transport.FrameKindClientHello:
		// Respond with server hello
		s.Logger.Info("Client hello - replying with server hello...", slog.Int("connID", conn.ID))
		frm, err := transport.FromProtoMsg(transport.FrameKindServerHello, &eventalepb.WireServerHello{
			ServerVersion: &eventalepb.SemanticVersion{
				Major: 0,
				Minor: 0,
				Patch: 1,
			},
		})
		if err != nil {
			return fmt.Errorf("from proto msg: %v", err)
		}
		if err := conn.Send(context.TODO(), frm); err != nil {
			return fmt.Errorf("conn send: %v", err)
		}

	case transport.FrameKindHeartbeat:
		// Respond with heartbeat again
		frm, err := transport.FromProtoMsg(transport.FrameKindHeartbeat, nil)
		if err != nil {
			return fmt.Errorf("from proto msg: %v", err)
		}
		if err := conn.Send(context.TODO(), frm); err != nil {
			return fmt.Errorf("conn send: %v", err)
		}
	}
	return nil
}

func (s *Server) readState() serverStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}
