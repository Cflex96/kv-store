package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

func newReader(data []byte) *bufio.Reader {
	return bufio.NewReader(bytes.NewReader(data))
}

func encodeMap(pairs ...string) []byte {
	var buf bytes.Buffer
	buf.WriteByte('%')
	fmt.Fprintf(&buf, "%d:", len(pairs))
	for _, p := range pairs {
		buf.Write(protocol.EncodeString(p))
	}
	return buf.Bytes()
}

func TestHandleStringResponse(t *testing.T) {
	tests := []struct {
		description string
		input       []byte
		expected    string
	}{
		{
			description: "simple string",
			input:       protocol.EncodeString("hello"),
			expected:    "hello",
		},
		{
			description: "empty string",
			input:       protocol.EncodeString(""),
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result, err := HandleStringResponse(newReader(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, &store.StringValue{Value: tt.expected}, result)
		})
	}
}

func TestHandleStringResponseErrors(t *testing.T) {
	tests := []struct {
		description string
		input       []byte
		expectedErr error
	}{
		{
			description: "empty input",
			input:       []byte{},
			expectedErr: io.EOF,
		},
		{
			description: "invalid header",
			input:       []byte("abc:hello\r\n"),
			expectedErr: protocol.ErrInvalidHeader,
		},
		{
			description: "missing CRLF delimiter",
			input:       []byte("5:helloXX"),
			expectedErr: protocol.ErrCorruptMsg,
		},
		{
			description: "message too big",
			input:       []byte("99999:hello\r\n"),
			expectedErr: protocol.ErrMessageTooBig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result, err := HandleStringResponse(newReader(tt.input))
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Nil(t, result)
		})
	}
}

func TestHandleListResponse(t *testing.T) {
	tests := []struct {
		description string
		input       []byte
		expected    []string
	}{
		{
			description: "list with multiple items",
			input:       protocol.EncodeList([]string{"val1", "val2", "val3"}),
			expected:    []string{"val1", "val2", "val3"},
		},
		{
			description: "single item list",
			input:       protocol.EncodeList([]string{"only"}),
			expected:    []string{"only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result, err := HandleListResponse(newReader(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, &store.ListValue{Value: tt.expected}, result)
		})
	}
}

func TestHandleListResponseErrors(t *testing.T) {
	tests := []struct {
		description string
		input       []byte
	}{
		{
			description: "wrong type byte",
			input:       encodeMap("k", "v"),
		},
		{
			description: "truncated: fewer items than count claims",
			input:       append([]byte("*3:"), protocol.EncodeString("only_one")...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result, err := HandleListResponse(newReader(tt.input))
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestHandleMapResponse(t *testing.T) {
	tests := []struct {
		description string
		input       []byte
		expected    map[string]string
	}{
		{
			description: "map with multiple pairs",
			input:       encodeMap("key1", "val1", "key2", "val2"),
			expected:    map[string]string{"key1": "val1", "key2": "val2"},
		},
		{
			description: "single pair",
			input:       encodeMap("k", "v"),
			expected:    map[string]string{"k": "v"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result, err := HandleMapResponse(newReader(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, &store.MapValue{Value: tt.expected}, result)
		})
	}
}

func TestHandleServerResponse(t *testing.T) {
	tests := []struct {
		description string
		input       []byte
		expected    store.Value
	}{
		{
			description: "routes string response",
			input:       protocol.EncodeString("hello"),
			expected:    &store.StringValue{Value: "hello"},
		},
		{
			description: "routes list response",
			input:       protocol.EncodeList([]string{"a", "b"}),
			expected:    &store.ListValue{Value: []string{"a", "b"}},
		},
		{
			description: "routes map response",
			input:       encodeMap("k", "v"),
			expected:    &store.MapValue{Value: map[string]string{"k": "v"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result, err := HandleServerResponse(newReader(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
