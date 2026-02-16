package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleConection(t *testing.T) {
	tests := []struct {
		input       string
		output      string
		description string
	}{
		{
			"SET",
			"OK",
			"Valid Predicate",
		},
		{
			"DED",
			"ERROR_INVALID_CMD",
			"Invalid predicate",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			server, client := net.Pipe()
			defer server.Close()
			defer client.Close()

			go handleConnection(server)

			msg := encodeMessage(test.input)
			client.Write(msg)

			resp, err := decodeMessage(client, 200)
			require.NoError(t, err)

			assert.Equal(t, test.output, resp, test.description)
		})
	}
}
