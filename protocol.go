package main

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
)

// 64kB
const maxMessageSize int = 16000

var (
	ErrMissingHeader = errors.New("missing header")
	ErrInvalidHeader = errors.New("invalid header")
	ErrMessageTooBig = errors.New("message too big")
	ErrCorruptMsg    = errors.New("corrupt message")
)

func decodeMessage(rd *bufio.Reader) (string, error) {
	header, err := rd.ReadString(':')
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrMissingHeader, err)
	}

	messageSize, err := strconv.Atoi(header[:len(header)-1])
	if err != nil {
		return "", fmt.Errorf("%w: could not parse %q as int", ErrInvalidHeader, header[:len(header)-1])
	}
	if messageSize > maxMessageSize {
		return "", fmt.Errorf("%w: %d bytes exceeds %d limit", ErrMessageTooBig, messageSize, maxMessageSize)
	}

	// +2 to account for "\r\n"
	buffer := make([]byte, messageSize+2)
	rd.Read(buffer)

	if string(buffer[messageSize:]) != "\r\n" {
		return "", fmt.Errorf("%w: missing \\r\\n delimiter", ErrCorruptMsg)
	}

	return string(buffer[:len(buffer)-2]), nil
}

func encodeMessage(msg string) []byte {
	ln := len(msg)

	formattedMsg := fmt.Sprintf("%d:%s\r\n", ln, msg)
	return []byte(formattedMsg)
}
