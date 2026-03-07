package dispatch

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Cflex96/kv-store/store"
)

func lpush(args []string, s *store.MemoryStore, writer io.Writer) error {
	list, err := getOrCreateListValue(args[0], s)
	if err != nil {
		return writeStringResponse(err.Error(), writer)
	}

	newValues := args[1:]
	newSlice := make([]string, len(list.Value)+len(newValues))
	for i, v := range newValues {
		newSlice[len(newValues)-i-1] = v
	}
	copy(newSlice[len(newValues):], list.Value)

	s.Set(args[0], &store.ListValue{
		Value: newSlice,
	})
	return writeStringResponse(strconv.Itoa(len(newSlice)), writer)
}

func rpush(args []string, s *store.MemoryStore, writer io.Writer) error {
	list, err := getOrCreateListValue(args[0], s)
	if err != nil {
		return writeStringResponse(err.Error(), writer)
	}

	newValues := args[1:]
	newSlice := make([]string, len(list.Value)+len(newValues))
	copy(newSlice, list.Value)
	copy(newSlice[len(list.Value):], newValues)

	s.Set(args[0], &store.ListValue{
		Value: newSlice,
	})
	return writeStringResponse(strconv.Itoa(len(newSlice)), writer)
}

func lpop(args []string, s *store.MemoryStore, writer io.Writer) error {
	result, ok := s.Get(args[0])
	if !ok {
		return writeStringResponse("nil", writer)
	}
	list, ok := result.(*store.ListValue)
	if !ok {
		return writeStringResponse(ErrInvalidType.Error(), writer)
	}

	deletedValue := list.Value[0]
	list.Value = list.Value[1:]
	s.Set(args[0], list)
	return writeStringResponse(deletedValue, writer)
}

func rpop(args []string, s *store.MemoryStore, writer io.Writer) error {
	result, ok := s.Get(args[0])
	if !ok {
		return writeStringResponse("nil", writer)
	}
	list, ok := result.(*store.ListValue)
	if !ok {
		return writeStringResponse(ErrInvalidType.Error(), writer)
	}

	deletedValue := list.Value[len(list.Value)-1]
	list.Value = list.Value[:len(list.Value)-1]
	s.Set(args[0], list)
	return writeStringResponse(deletedValue, writer)
}

func lrange(args []string, s *store.MemoryStore, writer io.Writer) error {
	result, ok := s.Get(args[0])
	if !ok {
		return writeStringResponse("nil", writer)
	}

	list, ok := result.(*store.ListValue)
	if !ok {
		return writeStringResponse(ErrInvalidType.Error(), writer)
	}

	size := len(list.Value)

	leftLimit, err := strconv.Atoi(args[1])
	if err != nil {
		return writeStringResponse(fmt.Errorf("%w:%w", ErrWrongArgs, err).Error(), writer)
	}
	if !checkLimit(leftLimit, size) {
		return writeListResponse([]string{}, writer)
	}

	// go slices do bot include the item at right limit so +1
	rightLimit, err := strconv.Atoi(args[2])
	if err != nil {
		return writeStringResponse(fmt.Errorf("%w:%w", ErrWrongArgs, err).Error(), writer)
	}

	if rightLimit < 0 {
		return writeListResponse(list.Value[leftLimit:], writer)
	}

	rightLimit += 1
	if !checkLimit(rightLimit, size) {
		return writeListResponse([]string{}, writer)
	}

	return writeListResponse(list.Value[leftLimit:rightLimit], writer)
}

func getOrCreateListValue(key string, s *store.MemoryStore) (*store.ListValue, error) {
	existing, ok := s.Get(key)
	if !ok {
		return &store.ListValue{}, nil
	}
	list, ok := existing.(*store.ListValue)
	if !ok {
		return nil, ErrInvalidType
	}
	return list, nil
}

func checkLimit(sliceLimit int, length int) bool {
	return sliceLimit <= length
}
