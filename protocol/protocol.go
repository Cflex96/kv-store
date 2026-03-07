package protocol

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// MaxMessageSize 64kb
const MaxMessageSize int = 16000

var (
	ErrMissingHeader = errors.New("missing header")
	ErrInvalidHeader = errors.New("invalid header")
	ErrMessageTooBig = errors.New("message too big")
	ErrCorruptMsg    = errors.New("corrupt message")
)

func DecodeMessage(rd *bufio.Reader) (string, error) {
	messageSize, err := GetMessageSize(rd)
	if err != nil {
		return "", err
	}

	// +2 to account for "\r\n"
	buffer := make([]byte, messageSize+2)
	rd.Read(buffer)

	if string(buffer[messageSize:]) != "\r\n" {
		return "", fmt.Errorf("%w: missing \\r\\n delimiter", ErrCorruptMsg)
	}

	return string(buffer[:len(buffer)-2]), nil
}

func EncodeString(msg string) []byte {
	ln := len(msg)

	formattedMsg := fmt.Sprintf("%d:%s\r\n", ln, msg)
	return []byte(formattedMsg)
}

func EncodeList(items []string) []byte {
	var buf bytes.Buffer
	buf.WriteByte('*')
	fmt.Fprintf(&buf, "%d:", len(items))
	for _, item := range items {
		buf.Write(EncodeString(item))
	}
	return buf.Bytes()
}

func EncodeMap(msg map[string]string) []byte {
	var buf bytes.Buffer
	buf.WriteByte('%')
	fmt.Fprintf(&buf, "%d:", len(msg))
	for k, v := range msg {
		buf.Write(EncodeString(k))
		buf.Write(EncodeString(v))
	}
	return buf.Bytes()
}

func GetMessageSize(rd *bufio.Reader) (int, error) {
	header, err := rd.ReadString(':')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return 0, io.EOF
		} else {
			return 0, fmt.Errorf("%w: %w", ErrMissingHeader, err)
		}
	}

	messageSize, err := strconv.Atoi(header[:len(header)-1])
	if err != nil {
		return 0, fmt.Errorf(
			"%w: could not parse %q as int",
			ErrInvalidHeader,
			header[:len(header)-1],
		)
	}
	if messageSize > MaxMessageSize {
		return 0, fmt.Errorf(
			"%w: %d bytes exceeds %d limit",
			ErrMessageTooBig,
			messageSize,
			MaxMessageSize,
		)
	}
	return messageSize, nil
}
