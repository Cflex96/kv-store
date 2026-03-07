package protocol

import (
	"bufio"
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeMessage(t *testing.T) {
	tests := []struct {
		input       string
		ouput       string
		description string
		isError     bool
	}{
		{
			"SET\r\n",
			"",
			"Missing header",
			true,
		},
		{
			"3:SET",
			"",
			"Missing newline",
			true,
		},
		{
			"12:SET chris 26\r\n",
			"SET chris 26",
			"Valid Set Command",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			server, client := net.Pipe()
			defer server.Close()
			defer client.Close()

			go func() {
				client.Write([]byte(test.input))
				client.Close()
			}()

			rd := bufio.NewReader(server)
			resp, err := DecodeMessage(rd)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.ouput, resp)
			}
		})
	}
}

func TestEncodeString(t *testing.T) {
	tests := []struct {
		description string
		input       string
		expected    []byte
	}{
		{
			description: "simple string",
			input:       "hello",
			expected:    []byte("5:hello\r\n"),
		},
		{
			description: "empty string",
			input:       "",
			expected:    []byte("0:\r\n"),
		},
		{
			description: "string with spaces",
			input:       "SET key value",
			expected:    []byte("13:SET key value\r\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			assert.Equal(t, tt.expected, EncodeString(tt.input))
		})
	}
}

func TestEncodeList(t *testing.T) {
	tests := []struct {
		description string
		input       []string
	}{
		{
			description: "multiple items",
			input:       []string{"val1", "val2", "val3"},
		},
		{
			description: "single item",
			input:       []string{"only"},
		},
		{
			description: "empty list",
			input:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			encoded := EncodeList(tt.input)
			assert.Equal(t, byte('*'), encoded[0])

			rd := bufio.NewReader(bytes.NewReader(encoded[1:]))
			count, err := GetMessageSize(rd)
			assert.NoError(t, err)
			assert.Equal(t, len(tt.input), count)

			for _, expected := range tt.input {
				msg, err := DecodeMessage(rd)
				assert.NoError(t, err)
				assert.Equal(t, expected, msg)
			}
		})
	}
}

func TestEncodeMap(t *testing.T) {
	tests := []struct {
		description string
		input       map[string]string
	}{
		{
			description: "single pair",
			input:       map[string]string{"key": "value"},
		},
		{
			description: "multiple pairs",
			input:       map[string]string{"k1": "v1", "k2": "v2"},
		},
		{
			description: "empty map",
			input:       map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			encoded := EncodeMap(tt.input)
			assert.Equal(t, byte('%'), encoded[0])

			rd := bufio.NewReader(bytes.NewReader(encoded[1:]))
			count, err := GetMessageSize(rd)
			assert.NoError(t, err)
			assert.Equal(t, len(tt.input), count)

			decoded := make(map[string]string, count)
			for range count {
				k, err := DecodeMessage(rd)
				assert.NoError(t, err)
				v, err := DecodeMessage(rd)
				assert.NoError(t, err)
				decoded[k] = v
			}
			assert.Equal(t, tt.input, decoded)
		})
	}
}
