package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

// 64kB
const maxMessageSize int = 16000

func decodeMessage(conn net.Conn, timeout int) (string, error) {
	conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))

	rd := bufio.NewReader(conn)

	header, err := rd.ReadString(':')
	if err != nil {
		return "", fmt.Errorf("error reading message with error: %s", err)
	}

	messageSize, err := strconv.Atoi(header[:len(header)-1])
	if err != nil {
		return "", fmt.Errorf("invalid Header; could not convert to int: %s", err)
	}
	if messageSize > maxMessageSize {
		return "", errors.New("body may not exceed 64kB")
	}

	// +2 to account for "\r\n"
	buffer := make([]byte, messageSize+2)
	rd.Read(buffer)

	if string(buffer[messageSize:]) != "\r\n" {
		return "", errors.New("message possibly corrupt, could not detect \\r\\n'")
	}

	return string(buffer[:len(buffer)-2]), nil
}

func encodeMessage(msg string) []byte {
	ln := len(msg)

	formattedMsg := fmt.Sprintf("%d:%s\r\n", ln, msg)
	return []byte(formattedMsg)
}
