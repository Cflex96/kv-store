package main

import (
	"log"
	"net"
)

type Command string

const (
	GET Command = "GET"
	SET Command = "SET"
	DEL Command = "DEL"
	PNG Command = "PNG"
)

const (
	ErrInvalidCMD string = "ERROR_INVALID_CMD"
	SuccessMsg    string = "OK"
)

func runTCPServer() {
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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	msg, err := decodeMessage(conn, 200)
	if err != nil {
		log.Println(err)
		return
	}

	cmd := Command(msg[:3])

	switch cmd {
	case GET:
		_, err = conn.Write(encodeMessage(SuccessMsg))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return
		}

	case SET, DEL, PNG:
		_, err = conn.Write(encodeMessage(SuccessMsg))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return
		}

	default:
		_, err = conn.Write(encodeMessage(ErrInvalidCMD))
		if err != nil {
			log.Printf("Server write error: %v", err)
			return
		}
	}
}
