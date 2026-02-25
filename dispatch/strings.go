package dispatch

import (
	"log"
	"net"

	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

func get(args []string, store *store.MemoryStore, conn net.Conn) error {
	result, ok := store.Get(args[0])
	if !ok {
		return nil
	}
	_, err := conn.Write(protocol.EncodeMessage(result.String()))
	if err != nil {
		log.Printf("Server write error: %v", err)
		return err
	}
	return nil
}

func set(args []string, s *store.MemoryStore, conn net.Conn) error {
	s.Set(args[0], &store.StringValue{
		Value: args[1],
	})

	_, err := conn.Write(protocol.EncodeMessage(SuccessMsg))
	if err != nil {
		log.Printf("Server write error: %v", err)
		return err
	}
	return nil
}

func del(args []string, store *store.MemoryStore, conn net.Conn) error {
	var err error
	if ok := store.Delete(args[0]); ok {
		_, err = conn.Write(protocol.EncodeMessage(""))
	} else {
		_, err = conn.Write(protocol.EncodeMessage(ErrNotFound.Error()))
	}
	if err != nil {
		log.Printf("Server write error: %v", err)
		return err
	}
	return err
}
