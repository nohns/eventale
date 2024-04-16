package main

import (
	"log"

	"github.com/nohns/eventale"
)

func main() {
	_, err := eventale.Dial("127.0.0.1:9999")
	if err != nil {
		log.Fatalf("Failed to dial client: %v", err)
	}
}
