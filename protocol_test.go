package main

import (
	"bufio"
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
			resp, err := decodeMessage(rd)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.ouput, resp)
			}
		})
	}
}
