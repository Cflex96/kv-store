package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Command string

const (
	GET  Command = "GET"
	SET  Command = "SET"
	DEL  Command = "DEL"
	PING Command = "PING"
)

const (
	ErrInvalidCmd = "ERR_INVALID_CMD"
	ErrWrongArgs  = "ERR_WRONG_ARGS"
	ErrNotFound   = "ERR_NOT_FOUND"
)

const SuccessMsg string = "OK"

const (
	TCPConnectionIdleTimeout = 200
	serverURL                = "localhost"
	port                     = "6379"
)

func RunTCPServer(ctx context.Context) int {
	log.Println("Starting up database server")
	store := New()
	log.Println("Initialized store")

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
			handleConnection(connCtx, conn, store)
		}()
	}
	wg.Wait()
	return 0
}

func handleConnection(ctx context.Context, conn net.Conn, store *MemoryStore) {
	defer conn.Close()
	rd := bufio.NewReader(conn)

	for {
		// idle connection timeout
		conn.SetReadDeadline(
			time.Now().Add(time.Duration(time.Millisecond) * TCPConnectionIdleTimeout),
		)

		if err := handleMessage(conn, rd, store); err != nil {
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

func handleMessage(conn net.Conn, rd *bufio.Reader, store *MemoryStore) error {
	msg, err := decodeMessage(rd)
	if err != nil {
		return err
	}
	fields := strings.Fields(msg)
	cmd := Command(fields[0])
	args := fields[1:]

	switch cmd {
	case GET:
		if len(args) != 1 {
			conn.Write(
				encodeMessage(
					fmt.Sprintf("%s: GET requires 1 arg, got %d", ErrWrongArgs, len(args)),
				),
			)
			return err
		}
		result := store.Get(args[0])
		_, err = conn.Write(encodeMessage(result))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return err
		}

	case SET:
		if len(args) != 2 {
			conn.Write(
				encodeMessage(
					fmt.Sprintf("%s: SET requires 2 args, got %d", ErrWrongArgs, len(args)),
				),
			)
			return err
		}
		store.Set(args[0], args[1])
		_, err = conn.Write(encodeMessage(SuccessMsg))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return err
		}

	case DEL:
		if len(args) != 1 {
			conn.Write(
				encodeMessage(
					fmt.Sprintf("%s: DEL requires 1 arg, got %d", ErrWrongArgs, len(args)),
				),
			)
			return err
		}
		if ok := store.Delete(args[0]); ok {
			_, err = conn.Write(encodeMessage(""))
		} else {
			_, err = conn.Write(encodeMessage(ErrNotFound))
		}
		if err != nil {
			log.Printf("Server write error: %v", err)
			return err
		}

	case PING:
		if len(args) != 0 {
			conn.Write(
				encodeMessage(
					fmt.Sprintf("%s: PING takes no args, got %d", ErrWrongArgs, len(args)),
				),
			)
			return err
		}
		_, err = conn.Write(encodeMessage("PONG"))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return err
		}

	default:
		_, err = conn.Write(encodeMessage(fmt.Sprintf("%s: %s", ErrInvalidCmd, cmd)))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return err
		}
	}
	return nil
}
