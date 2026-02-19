package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
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

const TCPConnectionIdleTimeout = 200

func RunTCPServer() {
	store := New()
	listener, err := net.Listen("tcp", "localhost:6379")
	if err != nil {
		log.Fatalf("Failed to Start Server: %s", err.Error())
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)

		go func() {
			defer cancel()
			go handleConnection(ctx, conn, store)
		}()
	}
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
