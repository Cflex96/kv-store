package main

import (
	"context"
	"os"

	"github.com/Cflex96/kv-store/server"
)

func main() {
	ctx := context.Background()
	code := server.RunTCPServer(ctx)
	os.Exit(code)
}
