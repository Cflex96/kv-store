package dispatch

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Cflex96/kv-store/store"
)

func mset(args []string, s *store.MemoryStore, writer io.Writer) error {
	key := args[0]
	alternatingKeyValue := args[1:]

	if (len(alternatingKeyValue) % 2) != 0 {
		return writeStringResponse(
			fmt.Errorf(
				"%w: %s",
				ErrWrongArgs,
				"method always needs a key value pair, recieved key without value",
			).Error(), writer)
	}

	var mapValue *store.MapValue
	existing, ok := s.Get(key)
	if ok {
		mapValue, ok = existing.(*store.MapValue)
		if !ok {
			return writeStringResponse(ErrInvalidType.Error(), writer)
		}
	} else {
		mapValue = &store.MapValue{Value: make(map[string]string, len(alternatingKeyValue)/2)}
		s.Set(key, mapValue)
	}

	for i := 0; i < len(alternatingKeyValue); i += 2 {
		mapValue.Value[alternatingKeyValue[i]] = alternatingKeyValue[i+1]
	}
	return writeStringResponse(strconv.Itoa(len(mapValue.Value)), writer)
}
