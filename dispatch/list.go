package dispatch

import (
	"net"
	"strconv"

	"github.com/Cflex96/kv-store/store"
)

func lpush(args []string, s *store.MemoryStore, conn net.Conn) error {
	list, err := getOrCreateListValue(args[0], s)
	if err != nil {
		return writeResponse(err.Error(), conn)
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
	return writeResponse(strconv.Itoa(len(newSlice)), conn)
}

func rpush(args []string, s *store.MemoryStore, conn net.Conn) error {
	list, err := getOrCreateListValue(args[0], s)
	if err != nil {
		return writeResponse(err.Error(), conn)
	}

	newValues := args[1:]
	newSlice := make([]string, len(list.Value)+len(newValues))
	copy(newSlice, list.Value)
	copy(newSlice[len(list.Value):], newValues)

	s.Set(args[0], &store.ListValue{
		Value: newSlice,
	})
	return writeResponse(strconv.Itoa(len(newSlice)), conn)
}

func lpop(args []string, s *store.MemoryStore, conn net.Conn) error {
	result, ok := s.Get(args[0])
	if !ok {
		return writeResponse("nil", conn)
	}
	list, ok := result.(*store.ListValue)
	if !ok {
		return writeResponse(ErrInvalidType.Error(), conn)
	}

	deletedValue := list.Value[0]
	list.Value = list.Value[1:]
	s.Set(args[0], list)
	return writeResponse(deletedValue, conn)
}

func rpop(args []string, s *store.MemoryStore, conn net.Conn) error {
	result, ok := s.Get(args[0])
	if !ok {
		return writeResponse("nil", conn)
	}
	list, ok := result.(*store.ListValue)
	if !ok {
		return writeResponse(ErrInvalidType.Error(), conn)
	}

	deletedValue := list.Value[len(list.Value)-1]
	list.Value = list.Value[:len(list.Value)-1]
	s.Set(args[0], list)
	return writeResponse(deletedValue, conn)
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
