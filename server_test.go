package main

import (
	"bufio"
	"fmt"
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
			encodeMessageToString("SET"),
			ErrWrongArgs,
			"Invalid command",
		},
		{
			encodeMessageToString("DED"),
			ErrInvalidCmd,
			"Invalid predicate",
		},
		{
			encodeMessageToString("SET chris 26"),
			"OK",
			"Valid Set command",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			server, client := net.Pipe()
			defer server.Close()
			defer client.Close()

			store := New()

			go func() {
				client.Write([]byte(test.input))
			}()

			go handleConnection(server, store)

			rd := bufio.NewReader(client)
			resp, err := decodeMessage(rd)
			require.NoError(t, err)

			assert.Contains(t, resp, test.output, test.description)
		})
	}

	t.Run("Test Commands mutate map", func(t *testing.T) {
		key := "testK"
		val := "testV"
		store := New()

		t.Run("Test Set in storage", func(t *testing.T) {
			msg, err := arrangeAndActOnMapMutationTest(fmt.Sprintf("SET %s %s", key, val), store)

			assert.NoError(t, err)
			assert.Equal(t, SuccessMsg, msg)
			assert.Equal(t, val, store.Get(key))
		})

		t.Run("Test Get from storage", func(t *testing.T) {
			msg2, err := arrangeAndActOnMapMutationTest(fmt.Sprintf("GET %s", key), store)

			assert.NoError(t, err)
			assert.Equal(t, val, msg2)
		})

		t.Run("test delete from storage", func(t *testing.T) {
			msg3, err := arrangeAndActOnMapMutationTest(fmt.Sprintf("DEL %s", key), store)

			assert.NoError(t, err)
			assert.Equal(t, "", msg3)
		})
	})
}

func encodeMessageToString(msg string) string {
	ln := len(msg)

	formattedMsg := fmt.Sprintf("%d:%s\r\n", ln, msg)
	return formattedMsg
}

func arrangeAndActOnMapMutationTest(msg string, store *MemoryStore) (string, error) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rd := bufio.NewReader(client)
	go func() {
		client.Write(encodeMessage(msg))
	}()

	go handleConnection(server, store)
	if msg, err := decodeMessage(rd); err != nil {
		return "", err
	} else {
		return msg, nil
	}
}
