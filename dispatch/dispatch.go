package dispatch

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

type Command string

const (
	GET    Command = "GET"
	SET    Command = "SET"
	DEL    Command = "DEL"
	PING   Command = "PING"
	LPUSH  Command = "LPUSH"
	RPUSH  Command = "RPUSH"
	LPOP   Command = "LPOP"
	RPOP   Command = "RPOP"
	LRANGE Command = "LRANGE"
	LLEN   Command = "LLEN"
)

var dispatch = map[Command]handler{
	GET:    {1, 1, get},
	SET:    {2, 2, set},
	DEL:    {1, 1, del},
	PING:   {0, 0, ping},
	LPUSH:  {2, -1, lpush},
	RPUSH:  {2, -1, rpush},
	LPOP:   {1, 1, lpop},
	RPOP:   {1, 1, rpop},
	LRANGE: {3, 3, lrange},
	LLEN:   {1, 1, llen},
}

var (
	ErrInvalidCmd  = errors.New("ERR_INVALID_CMD")
	ErrWrongArgs   = errors.New("ERR_WRONG_ARGS")
	ErrNotFound    = errors.New("ERR_NOT_FOUND")
	ErrInvalidType = errors.New("ERR_INVALID_TYPE")
)

const SuccessMsg string = "OK"

type handler struct {
	minArgs int
	maxArgs int
	fn      func(args []string, store *store.MemoryStore, writer io.Writer) error
}

type Dispatcher struct {
	dispatchTable map[Command]handler
	store         *store.MemoryStore
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
		conn.Write(protocol.EncodeString(ErrInvalidCmd.Error()))
		return ErrInvalidCmd
	}
	if len(args) < handler.minArgs || (len(args) > handler.maxArgs && handler.maxArgs != -1) {
		conn.Write(
			protocol.EncodeString(
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

func writeStringResponse(msg string, writer io.Writer) error {
	if _, err := writer.Write(protocol.EncodeString(msg)); err != nil {
		return err
	}
	return nil
}

func writeListResponse(msg []string, writer io.Writer) error {
	if _, err := writer.Write(protocol.EncodeList(msg)); err != nil {
		return err
	}
	return nil
}

func writeMapResponse(msg map[string]string, writer io.Writer) error {
	if _, err := writer.Write(protocol.EncodeMap(msg)); err != nil {
		return err
	}
	return nil
}

func ping(args []string, store *store.MemoryStore, writer io.Writer) error {
	_, err := writer.Write(protocol.EncodeString("PONG"))
	if err != nil {
		log.Printf("Server write error: %v", err)
		return err
	}
	return err
}
