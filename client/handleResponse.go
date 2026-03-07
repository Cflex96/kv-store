package client

import (
	"bufio"
	"errors"

	"github.com/Cflex96/kv-store/protocol"
	"github.com/Cflex96/kv-store/store"
)

func HandleServerResponse(rd *bufio.Reader) (store.Value, error) {
	firstByte, err := rd.Peek(1)
	if err != nil {
		return nil, err
	}

	switch firstByte[0] {
	case '*':
		if messages, err := HandleListResponse(rd); err != nil {
			return nil, err
		} else {
			return messages, nil
		}
	case '%':
		if messages, err := HandleMapResponse(rd); err != nil {
			return nil, err
		} else {
			return messages, nil
		}
	default:
		if message, err := HandleStringResponse(rd); err != nil {
			return nil, err
		} else {
			return message, nil
		}
	}
}

func HandleListResponse(rd *bufio.Reader) (*store.ListValue, error) {
	var messages []string
	firstByte, err := rd.ReadByte()
	if err != nil {
		return nil, err
	}
	if firstByte != '*' {
		return nil, errors.New("multiMessageResponse without * as first char")
	}

	msgCount, err := protocol.GetMessageSize(rd)
	if err != nil {
		return nil, err
	}
	messages = make([]string, msgCount)

	for i := range msgCount {
		msg, err := protocol.DecodeMessage(rd)
		if err != nil {
			return nil, err
		}
		messages[i] = msg
	}
	return &store.ListValue{
		Value: messages,
	}, nil
}

func HandleMapResponse(rd *bufio.Reader) (*store.MapValue, error) {
	var messages []string
	rd.ReadByte()

	msgCount, err := protocol.GetMessageSize(rd)
	if err != nil {
		return nil, err
	}
	messages = make([]string, msgCount)

	for i := range msgCount {
		msg, err := protocol.DecodeMessage(rd)
		if err != nil {
			return nil, err
		}
		messages[i] = msg
	}

	m := make(map[string]string, len(messages)/2)
	for i := 0; i < len(messages); i += 2 {
		m[messages[i]] = messages[i+1]
	}

	return &store.MapValue{
		Value: m,
	}, nil
}

func HandleStringResponse(rd *bufio.Reader) (*store.StringValue, error) {
	resp, err := protocol.DecodeMessage(rd)
	if err != nil {
		return nil, err
	}
	return &store.StringValue{
		Value: resp,
	}, nil
}
