package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

var desc = strings.TrimSpace(`
	Alice is the command-line interface for interacting with the Eventale server (taled). 
	Use it for querying events, checking stats or other magical things.
`)

func main() {
	app := &cli.App{
		Name:        "alice",
		Usage:       "ahah test",
		Description: desc,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
