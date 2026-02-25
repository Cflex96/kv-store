package dispatch

import (
	"bufio"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

func TestLpush(t *testing.T) {
	s := store.New()
	tests := []struct {
		input              []string
		expectedResp       string
		expectedStoreState []string
		description        string
	}{
		{
			input:              []string{"key", "value1"},
			expectedResp:       "1",
			expectedStoreState: []string{"value1"},
			description:        "LPUSH on empty key",
		},
		{
			input:              []string{"key", "value2", "value3"},
			expectedResp:       "3",
			expectedStoreState: []string{"value3", "value2", "value1"},
			description:        "LPUSH on existing key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()
			rd := bufio.NewReader(client)

			go func() {
				err := lpush(tt.input, s, server)
				assert.NoError(t, err)
			}()

			msg, err := protocol.DecodeMessage(rd)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResp, msg)
			storeState, ok := s.Get(tt.input[0])
			if !ok {
				assert.Fail(t, "store is nil after write")
			}
			storeValue := storeState.(*store.ListValue).Value
			assert.NotNil(t, storeValue)
			assert.Equal(t, tt.expectedStoreState, storeValue)
		})
	}
}

func TestRpush(t *testing.T) {
	s := store.New()
	tests := []struct {
		input              []string
		expectedResp       string
		expectedStoreState []string
		description        string
	}{
		{
			input:              []string{"key", "value1"},
			expectedResp:       "1",
			expectedStoreState: []string{"value1"},
			description:        "RPUSH on empty key",
		},
		{
			input:              []string{"key", "value2", "value3"},
			expectedResp:       "3",
			expectedStoreState: []string{"value1", "value2", "value3"},
			description:        "RPUSH on existing key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()
			rd := bufio.NewReader(client)

			go func() {
				err := rpush(tt.input, s, server)
				assert.NoError(t, err)
			}()

			msg, err := protocol.DecodeMessage(rd)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResp, msg)
			storeState, ok := s.Get(tt.input[0])
			if !ok {
				assert.Fail(t, "store is nil after write")
			}
			storeValue := storeState.(*store.ListValue).Value
			assert.NotNil(t, storeValue)
			assert.Equal(t, tt.expectedStoreState, storeValue)
		})
	}
}

func TestLpop(t *testing.T) {
	tests := []struct {
		input              []string
		store              *store.MemoryStore
		expectedResp       string
		expectedStoreState *store.ListValue
		description        string
	}{
		{
			input: []string{"key"},
			store: func() *store.MemoryStore {
				s := store.New()
				return s
			}(),
			expectedResp:       "nil",
			expectedStoreState: nil,
			description:        "LPOP on empty key",
		},
		{
			input: []string{"key"},
			store: func() *store.MemoryStore {
				s := store.New()
				s.Set("key", &store.ListValue{Value: []string{"value1", "value2"}})
				return s
			}(),
			expectedResp: "value1",
			expectedStoreState: &store.ListValue{
				Value: []string{"value2"},
			},
			description: "LPOP on existing key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()
			rd := bufio.NewReader(client)

			go func() {
				err := lpop(tt.input, tt.store, server)
				assert.NoError(t, err)
			}()

			msg, err := protocol.DecodeMessage(rd)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResp, msg, "Expected Message")

			storeState, ok := tt.store.Get("key")
			if !ok {
				assert.Nil(t, tt.expectedStoreState)
				return
			}
			storeValue := storeState.(*store.ListValue)
			assert.Equal(t, tt.expectedStoreState.Value, storeValue.Value, "Expected Store State")
		})
	}
}

func TestRpop(t *testing.T) {
	tests := []struct {
		input              []string
		store              *store.MemoryStore
		expectedResp       string
		expectedStoreState *store.ListValue
		description        string
	}{
		{
			input: []string{"key"},
			store: func() *store.MemoryStore {
				s := store.New()
				return s
			}(),
			expectedResp:       "nil",
			expectedStoreState: nil,
			description:        "RPOP on empty key",
		},
		{
			input: []string{"key"},
			store: func() *store.MemoryStore {
				s := store.New()
				s.Set("key", &store.ListValue{Value: []string{"value1", "value2"}})
				return s
			}(),
			expectedResp: "value2",
			expectedStoreState: &store.ListValue{
				Value: []string{"value1"},
			},
			description: "RPOP on existing key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()
			rd := bufio.NewReader(client)

			go func() {
				err := rpop(tt.input, tt.store, server)
				assert.NoError(t, err)
			}()

			msg, err := protocol.DecodeMessage(rd)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResp, msg, "Expected Message")

			storeState, ok := tt.store.Get("key")
			if !ok {
				assert.Nil(t, tt.expectedStoreState)
				return
			}
			storeValue := storeState.(*store.ListValue)
			assert.Equal(t, tt.expectedStoreState.Value, storeValue.Value, "Expected Store State")
		})
	}
}
