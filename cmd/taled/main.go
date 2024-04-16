package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/nohns/eventale"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	srv := eventale.NewServer("127.0.0.1:9999")
	srv.Logger = logger

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start taled: %v", err)
	}
}
