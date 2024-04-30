package eventale

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
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
	"github.com/nohns/eventale/internal/frame"
	"github.com/nohns/eventale/internal/uuid"
	"google.golang.org/protobuf/proto"
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

	lnr        net.Listener
	conns      []*connection.Conn
	authedkeys map[[32]byte]rsa.PublicKey
	state      serverStatus
	mu         sync.RWMutex
	nextid     int
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

func (s *Server) handleFrame(conn *connection.Conn, frm *frame.Frame) error {
	switch frm.Kind {
	case frame.FrameKindClientHello:
		// Respond with server hello
		var msg eventalepb.WireClientHello
		if err := proto.Unmarshal(frm.Payload, &msg); err != nil {
			return fmt.Errorf("decode client hello: %v", err)
		}

		// Gen secret symmetric encryption key, and encrypt using the connect
		// clients associated public key. This way, the client and decrypt it
		// and also use it when communicating.

		if len(msg.Signature) != 32 {
			return fmt.Errorf("incorrect key length")
		}
		s.mu.RLock()
		pubkey, ok := s.authedkeys[[32]byte(msg.Signature)]
		s.mu.RUnlock()
		if !ok {
			return fmt.Errorf("unauthorized")
		}
		plainkey := make([]byte, 32)
		n, err := rand.Reader.Read(plainkey)
		if err != nil {
			return fmt.Errorf("rand read enc key: %v", err)
		}
		if n != len(plainkey) {
			return fmt.Errorf("could not read enough bytes for enc key")
		}
		h := sha256.New()
		cipherkey, err := rsa.EncryptOAEP(h, rand.Reader, &pubkey, plainkey, nil)
		if err != nil {
			return fmt.Errorf("rsa encrypt: %v", err)
		}

		s.Logger.Info("Client hello - replying with server hello...", slog.Int("connID", conn.ID))
		frm, err := frame.Make(frame.FrameKindServerHello, frame.WithID(uuid.IDer), frame.WithRespondTo(frm.ID), frame.WithProto(&eventalepb.WireServerHello{
			ServerVersion: &eventalepb.SemanticVersion{
				Major: 0,
				Minor: 0,
				Patch: 1,
			},
			EncryptionKey: cipherkey,
		}))
		if err != nil {
			return fmt.Errorf("frame make: %v", err)
		}

		if err := conn.Send(context.TODO(), frm); err != nil {
			return fmt.Errorf("conn send: %v", err)
		}

	case frame.FrameKindHeartbeat:
		// Respond with heartbeat again
		frm, err := frame.Make(frame.FrameKindHeartbeat)
		if err != nil {
			return fmt.Errorf("frame make: %v", err)
		}
		if err := conn.Send(context.TODO(), frm); err != nil {
			return fmt.Errorf("conn send: %v", err)
		}
	case frame.FrameKindSecretPublish:

	}
	return nil
}

func (s *Server) readState() serverStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}
