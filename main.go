package main

import (
	"context"
	"os"
)

func main() {
	ctx := context.Background()
	code := RunTCPServer(ctx)
	os.Exit(code)
}
