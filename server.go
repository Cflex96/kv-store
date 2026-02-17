package main

import (
	"bufio"
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

const TCPConnectionTimeout = 200

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
		go handleConnection(conn, store)
	}
}

func handleConnection(conn net.Conn, store *MemoryStore) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(time.Duration(time.Millisecond) * TCPConnectionTimeout))

	rd := bufio.NewReader(conn)
	msg, err := decodeMessage(rd)
	if err != nil {
		log.Println(err)
		return
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
			return
		}
		result := store.Get(args[0])
		_, err = conn.Write(encodeMessage(result))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return
		}

	case SET:
		if len(args) != 2 {
			conn.Write(
				encodeMessage(
					fmt.Sprintf("%s: SET requires 2 args, got %d", ErrWrongArgs, len(args)),
				),
			)
			return
		}
		store.Set(args[0], args[1])
		_, err = conn.Write(encodeMessage(SuccessMsg))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return
		}

	case DEL:
		if len(args) != 1 {
			conn.Write(
				encodeMessage(
					fmt.Sprintf("%s: DEL requires 1 arg, got %d", ErrWrongArgs, len(args)),
				),
			)
			return
		}
		if ok := store.Delete(args[0]); ok {
			_, err = conn.Write(encodeMessage(""))
		} else {
			_, err = conn.Write(encodeMessage(ErrNotFound))
		}
		if err != nil {
			log.Printf("Server write error: %v", err)
			return
		}

	case PING:
		if len(args) != 0 {
			conn.Write(
				encodeMessage(
					fmt.Sprintf("%s: PING takes no args, got %d", ErrWrongArgs, len(args)),
				),
			)
			return
		}
		_, err = conn.Write(encodeMessage("PONG"))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return
		}

	default:
		_, err = conn.Write(encodeMessage(fmt.Sprintf("%s: %s", ErrInvalidCmd, cmd)))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return
		}
	}
}
