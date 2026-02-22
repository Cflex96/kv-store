package server

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Cflex96/kv-store/dispatch"
	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

const (
	TCPConnectionIdleTimeout = 200
	serverURL                = "localhost"
	port                     = "6379"
)

func RunTCPServer(ctx context.Context) int {
	log.Println("Starting up database server")
	store := store.New()
	log.Println("Initialized store")
	d := dispatch.NewDispatcher(store)

	serverCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listener, err := net.Listen("tcp", serverURL+":"+port)
	if err != nil {
		log.Fatalf("Failed to Start Server: %s", err.Error())
		return 1
	}
	defer listener.Close()
	log.Printf("Listening on port %s\n", port)

	go func() {
		<-serverCtx.Done()
		log.Print("Recieved shutdown signal, shutting down")
		listener.Close()
	}()

	var wg sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			if serverCtx.Err() != nil {
				break
			}
			log.Println("Error accepting conn:", err)
			continue
		}
		connCtx, cancel := context.WithTimeout(serverCtx, time.Minute*5)

		wg.Add(1)

		go func() {
			defer cancel()
			defer wg.Done()
			handleConnection(connCtx, conn, d)
		}()
	}
	wg.Wait()
	return 0
}

func handleConnection(ctx context.Context, conn net.Conn, d dispatch.Dispatcher) {
	defer conn.Close()
	rd := bufio.NewReader(conn)

	for {
		// idle connection timeout
		conn.SetReadDeadline(
			time.Now().Add(time.Duration(time.Millisecond) * TCPConnectionIdleTimeout),
		)

		if err := handleMessage(conn, rd, d); err != nil {
			return
		}

		// Connection Deadline timeout
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Println("handleConnection timed out")
			}
			return
		}
	}
}

func handleMessage(conn net.Conn, rd *bufio.Reader, d dispatch.Dispatcher) error {
	msg, err := protocol.DecodeMessage(rd)
	if err != nil {
		return err
	}
	fields := strings.Fields(msg)
	cmd := dispatch.Command(fields[0])
	args := fields[1:]

	if err := d.Dispatch(args, conn, cmd); err != nil {
		return err
	}
	return nil
}
