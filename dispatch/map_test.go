package dispatch

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

func TestMset(t *testing.T) {
	tests := []struct {
		description        string
		input              []string
		startingStoreKey   string
		startingStoreState map[string]string
		expectedStoreState map[string]string
		expectedResponse   string
	}{
		{
			description:        "test new entry",
			input:              []string{"storeKey", "mapKey", "mapVal"},
			startingStoreKey:   "storeKey",
			startingStoreState: nil,
			expectedStoreState: map[string]string{
				"mapKey": "mapVal",
			},
			expectedResponse: "1",
		},
		{
			description:      "test overriding entry",
			input:            []string{"storeKey", "mapKey", "mapValNew"},
			startingStoreKey: "storeKey",
			startingStoreState: map[string]string{
				"mapKey": "mapVal",
			},
			expectedStoreState: map[string]string{
				"mapKey": "mapValNew",
			},
			expectedResponse: "1",
		},
		{
			description:        "test new store entry with multiple kv pairs",
			input:              []string{"storeKey", "mapKey", "mapVal", "mapKey2", "mapVal2"},
			startingStoreKey:   "storeKey",
			startingStoreState: nil,
			expectedStoreState: map[string]string{
				"mapKey":  "mapVal",
				"mapKey2": "mapVal2",
			},
			expectedResponse: "2",
		},
		{
			description:      "test existing store entry add map entry",
			input:            []string{"storeKey", "mapKey2", "mapVal2"},
			startingStoreKey: "storeKey",
			startingStoreState: map[string]string{
				"mapKey": "mapVal",
			},
			expectedStoreState: map[string]string{
				"mapKey":  "mapVal",
				"mapKey2": "mapVal2",
			},
			expectedResponse: "2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			s := store.New()
			if tt.startingStoreState != nil && tt.startingStoreKey != "" {
				s.Set(tt.startingStoreKey, &store.MapValue{Value: tt.startingStoreState})
			}

			var buf bytes.Buffer

			err := mset(tt.input, s, &buf)
			assert.NoError(t, err)

			msg, err := protocol.DecodeMessage(bufio.NewReader(&buf))
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResponse, msg)
		})
	}
}
