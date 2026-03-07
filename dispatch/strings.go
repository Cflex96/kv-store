package dispatch

import (
	"io"
	"log"

	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

func get(args []string, store *store.MemoryStore, writer io.Writer) error {
	result, ok := store.Get(args[0])
	if !ok {
		return nil
	}
	_, err := writer.Write(protocol.EncodeString(result.String()))
	if err != nil {
		log.Printf("Server write error: %v", err)
		return err
	}
	return nil
}

func set(args []string, s *store.MemoryStore, writer io.Writer) error {
	s.Set(args[0], &store.StringValue{
		Value: args[1],
	})

	_, err := writer.Write(protocol.EncodeString(SuccessMsg))
	if err != nil {
		log.Printf("Server write error: %v", err)
		return err
	}
	return nil
}

func del(args []string, store *store.MemoryStore, writer io.Writer) error {
	var err error
	if ok := store.Delete(args[0]); ok {
		_, err = writer.Write(protocol.EncodeString(""))
	} else {
		_, err = writer.Write(protocol.EncodeString(ErrNotFound.Error()))
	}
	if err != nil {
		log.Printf("Server write error: %v", err)
		return err
	}
	return err
}
