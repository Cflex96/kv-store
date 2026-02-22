package dispatch

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

var (
	ErrInvalidCmd = errors.New("ERR_INVALID_CMD")
	ErrWrongArgs  = errors.New("ERR_WRONG_ARGS")
	ErrNotFound   = errors.New("ERR_NOT_FOUND")
)

type Command string

const (
	GET  Command = "GET"
	SET  Command = "SET"
	DEL  Command = "DEL"
	PING Command = "PING"
)

const SuccessMsg string = "OK"

type handler struct {
	minArgs int
	maxArgs int
	fn      func(args []string, store *store.MemoryStore, conn net.Conn) error
}

type Dispatcher struct {
	dispatchTable map[Command]handler
	store         *store.MemoryStore
}

var dispatch = map[Command]handler{
	GET:  {1, 1, handleGet},
	SET:  {2, 2, handleSet},
	DEL:  {1, 1, handleDel},
	PING: {0, 0, handlePing},
}

func NewDispatcher(store *store.MemoryStore) Dispatcher {
	return Dispatcher{
		dispatchTable: dispatch,
		store:         store,
	}
}

func (d Dispatcher) Dispatch(args []string, conn net.Conn, cmd Command) error {
	handler, ok := d.dispatchTable[cmd]
	if !ok {
		conn.Write(protocol.EncodeMessage(ErrInvalidCmd.Error()))
		return ErrInvalidCmd
	}
	if len(args) < handler.minArgs || len(args) > handler.maxArgs {
		conn.Write(
			protocol.EncodeMessage(
				fmt.Errorf("%w: %s takes %d to %d args, got %d", ErrWrongArgs, string(cmd), handler.minArgs, handler.maxArgs, len(args)).
					Error(),
			),
		)
		return ErrWrongArgs
	}
	err := handler.fn(args, d.store, conn)
	if err != nil {
		return err
	}
	return nil
}

func handleGet(args []string, store *store.MemoryStore, conn net.Conn) error {
	if len(args) != 1 {
		conn.Write(
			protocol.EncodeMessage(
				fmt.Errorf("%w: GET requires 1 arg, got %d", ErrWrongArgs, len(args)).Error(),
			),
		)
		return ErrWrongArgs
	}
	result := store.Get(args[0])
	_, err := conn.Write(protocol.EncodeMessage(result.String()))
	if err != nil {
		log.Printf("Server write error: %v", err)
		return err
	}
	return nil
}

func handleSet(args []string, s *store.MemoryStore, conn net.Conn) error {
	if len(args) != 2 {
		conn.Write(
			protocol.EncodeMessage(
				fmt.Errorf("%w: SET requires 2 args, got %d", ErrWrongArgs, len(args)).Error(),
			),
		)
		return ErrWrongArgs
	}

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

func handleDel(args []string, store *store.MemoryStore, conn net.Conn) error {
	if len(args) != 1 {
		conn.Write(
			protocol.EncodeMessage(
				fmt.Errorf("%w: DEL requires 1 arg, got %d", ErrWrongArgs, len(args)).Error(),
			),
		)
		return ErrWrongArgs
	}
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

func handlePing(args []string, store *store.MemoryStore, conn net.Conn) error {
	_, err := conn.Write(protocol.EncodeMessage("PONG"))
	if err != nil {
		log.Printf("Server write error: %v", err)
		return err
	}
	return err
}
