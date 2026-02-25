package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Cflex96/kv-store/dispatch"
	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

func TestRunTCPServer(t *testing.T) {
	t.Run("Test SIGTERM", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		c := make(chan int)
		go func(c chan int) {
			code := RunTCPServer(ctx)
			c <- code
		}(c)

		// give server time to startup
		time.Sleep(time.Second)
		cancel()

		select {
		case code := <-c:
			assert.Equal(t, 0, code)
		case <-time.After(5 * time.Second):
			t.Fatal("server did not shut down in time")
		}
	})
}

func TestHandleConnection(t *testing.T) {
	tests := []struct {
		input       string
		output      string
		description string
	}{
		{
			encodeMessageToString("SET"),
			dispatch.ErrWrongArgs.Error(),
			"Missing args",
		},
		{
			encodeMessageToString("DED"),
			dispatch.ErrInvalidCmd.Error(),
			"Invalid command",
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

			store := store.New()
			d := dispatch.NewDispatcher(store)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			go handleConnection(ctx, server, d)

			client.Write([]byte(test.input))

			rd := bufio.NewReader(client)
			resp, err := protocol.DecodeMessage(rd)
			require.NoError(t, err)

			assert.Contains(t, resp, test.output, test.description)

			client.Close()
		})
	}
}

func TestHandleMessage(t *testing.T) {
	t.Run("same connection handles multiple messages", func(t *testing.T) {
		store := store.New()
		d := dispatch.NewDispatcher(store)
		server, client := net.Pipe()
		defer server.Close()
		defer client.Close()

		serverReader := bufio.NewReader(server)
		clientReader := bufio.NewReader(client)

		results := make([]string, 5)

		for i := range 5 {
			go client.Write(protocol.EncodeMessage("PING"))

			go handleMessage(server, serverReader, d)
			res, err := protocol.DecodeMessage(clientReader)
			assert.NoError(t, err)
			results[i] = res
		}

		for _, v := range slices.All(results) {
			assert.Equal(t, v, "PONG")
		}
	})

	t.Run("Test Commands mutate map", func(t *testing.T) {
		key := "testK"
		val := "testV"
		store := store.New()
		d := dispatch.NewDispatcher(store)

		t.Run("Test Set in storage", func(t *testing.T) {
			msg, err := arrangeAndActOnMapMutationTest(fmt.Sprintf("SET %s %s", key, val), d)

			assert.NoError(t, err)
			assert.Equal(t, dispatch.SuccessMsg, msg)
			storedVal, ok := store.Get(key)
			if !ok {
				assert.Fail(t, "store empty")
			}
			assert.Equal(t, val, storedVal.String())
		})

		t.Run("Test Get from storage", func(t *testing.T) {
			msg2, err := arrangeAndActOnMapMutationTest(fmt.Sprintf("GET %s", key), d)

			assert.NoError(t, err)
			assert.Equal(t, val, msg2)
		})

		t.Run("test delete from storage", func(t *testing.T) {
			msg3, err := arrangeAndActOnMapMutationTest(fmt.Sprintf("DEL %s", key), d)

			assert.NoError(t, err)
			assert.Equal(t, "", msg3)
		})
	})

	t.Run("Handles EOF Client closes Reader", func(t *testing.T) {
		server, client := net.Pipe()
		defer server.Close()

		store := store.New()
		d := dispatch.NewDispatcher(store)

		rd := bufio.NewReader(server)

		client.Close()
		err := handleMessage(server, rd, d)
		assert.ErrorIs(t, err, io.EOF)
	})
}

func encodeMessageToString(msg string) string {
	ln := len(msg)

	formattedMsg := fmt.Sprintf("%d:%s\r\n", ln, msg)
	return formattedMsg
}

func arrangeAndActOnMapMutationTest(msg string, d dispatch.Dispatcher) (string, error) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	serverRd := bufio.NewReader(server)
	clientRd := bufio.NewReader(client)

	go client.Write(protocol.EncodeMessage(msg))

	go handleMessage(server, serverRd, d)

	if msg, err := protocol.DecodeMessage(clientRd); err != nil {
		return "", err
	} else {
		return msg, nil
	}
}
