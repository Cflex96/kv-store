package dispatch

import (
	"bufio"
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Cflex96/kv-store/client"
	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

func TestLPush(t *testing.T) {
	s := store.New()
	tests := []struct {
		input              []string
		store              *store.MemoryStore
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
		{
			input: []string{"key", "value1"},
			store: func() *store.MemoryStore {
				s := store.New()
				s.Set("key", &store.StringValue{Value: "hello"})
				return s
			}(),
			expectedResp:       ErrInvalidType.Error(),
			expectedStoreState: nil,
			description:        "LPUSH on key with wrong type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()
			rd := bufio.NewReader(client)

			activeStore := s
			if tt.store != nil {
				activeStore = tt.store
			}

			go func() {
				err := lpush(tt.input, activeStore, server)
				assert.NoError(t, err)
			}()

			msg, err := protocol.DecodeMessage(rd)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResp, msg)

			if tt.expectedStoreState != nil {
				storeState, ok := activeStore.Get(tt.input[0])
				if !ok {
					assert.Fail(t, "store is nil after write")
				}
				storeValue := storeState.(*store.ListValue).Value
				assert.NotNil(t, storeValue)
				assert.Equal(t, tt.expectedStoreState, storeValue)
			}
		})
	}
}

func TestRpush(t *testing.T) {
	s := store.New()
	tests := []struct {
		input              []string
		store              *store.MemoryStore
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
		{
			input: []string{"key", "value1"},
			store: func() *store.MemoryStore {
				s := store.New()
				s.Set("key", &store.StringValue{Value: "hello"})
				return s
			}(),
			expectedResp:       ErrInvalidType.Error(),
			expectedStoreState: nil,
			description:        "RPUSH on key with wrong type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()
			rd := bufio.NewReader(client)

			activeStore := s
			if tt.store != nil {
				activeStore = tt.store
			}

			go func() {
				err := rpush(tt.input, activeStore, server)
				assert.NoError(t, err)
			}()

			msg, err := protocol.DecodeMessage(rd)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResp, msg)

			if tt.expectedStoreState != nil {
				storeState, ok := activeStore.Get(tt.input[0])
				if !ok {
					assert.Fail(t, "store is nil after write")
				}
				storeValue := storeState.(*store.ListValue).Value
				assert.NotNil(t, storeValue)
				assert.Equal(t, tt.expectedStoreState, storeValue)
			}
		})
	}
}

func TestLPop(t *testing.T) {
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
		{
			input: []string{"key"},
			store: func() *store.MemoryStore {
				s := store.New()
				s.Set("key", &store.ListValue{Value: []string{"value1"}})
				return s
			}(),
			expectedResp:       "value1",
			expectedStoreState: nil,
			description:        "LPOP removes key when last element is popped",
		},
		{
			input: []string{"key"},
			store: func() *store.MemoryStore {
				s := store.New()
				s.Set("key", &store.StringValue{Value: "hello"})
				return s
			}(),
			expectedResp:       ErrInvalidType.Error(),
			expectedStoreState: nil,
			description:        "LPOP on key with wrong type",
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
			if tt.expectedStoreState != nil {
				storeValue := storeState.(*store.ListValue)
				assert.Equal(
					t,
					tt.expectedStoreState.Value,
					storeValue.Value,
					"Expected Store State",
				)
			}
		})
	}
}

func TestRPop(t *testing.T) {
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
		{
			input: []string{"key"},
			store: func() *store.MemoryStore {
				s := store.New()
				s.Set("key", &store.ListValue{Value: []string{"value1"}})
				return s
			}(),
			expectedResp:       "value1",
			expectedStoreState: nil,
			description:        "RPOP removes key when last element is popped",
		},
		{
			input: []string{"key"},
			store: func() *store.MemoryStore {
				s := store.New()
				s.Set("key", &store.StringValue{Value: "hello"})
				return s
			}(),
			expectedResp:       ErrInvalidType.Error(),
			expectedStoreState: nil,
			description:        "RPOP on key with wrong type",
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
			if tt.expectedStoreState != nil {
				storeValue := storeState.(*store.ListValue)
				assert.Equal(
					t,
					tt.expectedStoreState.Value,
					storeValue.Value,
					"Expected Store State",
				)
			}
		})
	}
}

func TestLRange(t *testing.T) {
	tests := []struct {
		input           []string
		key             string
		description     string
		store           *store.MemoryStore
		inputStoreState []string
		expectedResp    []string
		expectedError   string
	}{
		{
			input:           []string{"key", "0", "2"},
			key:             "key",
			description:     "Get List values from 0 to 2, values exist",
			inputStoreState: []string{"val1", "val2", "val3"},
			expectedResp:    []string{"val1", "val2", "val3"},
		},
		{
			input:           []string{"key", "0", "-1"},
			key:             "key",
			description:     "Get all values with 0 to -1",
			inputStoreState: []string{"val1", "val2", "val3"},
			expectedResp:    []string{"val1", "val2", "val3"},
		},
		{
			input:           []string{"key", "5", "10"},
			key:             "key",
			description:     "Non-existing index range",
			inputStoreState: []string{"val1", "val2", "val3"},
			expectedResp:    []string{},
		},
		{
			input:           []string{"key", "10", "1"},
			key:             "key",
			description:     "left index out of bounds",
			inputStoreState: []string{"val1", "val2", "val3"},
			expectedResp:    []string{},
		},
		{
			input:           []string{"key", "0", "10"},
			key:             "key",
			description:     "right index out of bound",
			inputStoreState: []string{"val1", "val2", "val3"},
			expectedResp:    []string{},
		},
		{
			input: []string{"key", "0", "-1"},
			key:   "key",
			store: func() *store.MemoryStore {
				s := store.New()
				s.Set("key", &store.StringValue{Value: "hello"})
				return s
			}(),
			description:   "LRANGE on key with wrong type",
			expectedError: ErrInvalidType.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			s := store.New()
			if tt.store != nil {
				s = tt.store
			} else {
				s.Set(tt.key, &store.ListValue{Value: tt.inputStoreState})
			}

			var buf bytes.Buffer
			err := lrange(tt.input, s, &buf)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				val, err := client.HandleStringResponse(bufio.NewReader(&buf))
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, val.Value)
			} else {
				val, err := client.HandleListResponse(bufio.NewReader(&buf))
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, val.Value)
			}
		})
	}
}

// maybe test in future: create list len 1, pop, and then get length. Will the key persist? will it get cleaned up?
func TestLlen(t *testing.T) {
	tests := []struct {
		input           []string
		description     string
		store           *store.MemoryStore
		inputStoreState []string
		expectedResp    string
	}{
		{
			input:           []string{"key"},
			description:     "get len of valid key",
			inputStoreState: []string{"key", "0"},
			expectedResp:    "1",
		},
		{
			input:           []string{"nokey"},
			description:     "get len of invalid key",
			inputStoreState: []string{"key", "0"},
			expectedResp:    "0",
		},
		{
			input: []string{"key"},
			store: func() *store.MemoryStore {
				s := store.New()
				s.Set("key", &store.StringValue{Value: "hello"})
				return s
			}(),
			description:  "LLEN on key with wrong type",
			expectedResp: ErrInvalidType.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			s := store.New()
			if tt.store != nil {
				s = tt.store
			} else {
				s.Set(tt.inputStoreState[0], &store.ListValue{Value: tt.inputStoreState[1:]})
			}

			var buf bytes.Buffer

			err := llen(tt.input, s, &buf)
			assert.NoError(t, err)

			val, err := protocol.DecodeMessage(bufio.NewReader(&buf))
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResp, val)
		})
	}
}
