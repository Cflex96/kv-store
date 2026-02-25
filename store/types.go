package store

import "fmt"

const (
	ListType   = "list"
	MapType    = "map"
	StringType = "string"
)

type Value interface {
	Type() string
	String() string
}

type StringValue struct {
	Value string
}

func (v *StringValue) Type() string {
	return StringType
}

func (v *StringValue) String() string {
	return v.Value
}

type ListValue struct {
	Value []string
}

func (v *ListValue) Type() string {
	return ListType
}

func (v *ListValue) String() string {
	return fmt.Sprintf("%v", v.Value)
}

type MapValue struct {
	Value map[string]string
}

func (v *MapValue) Type() string {
	return MapType
}

func (v *MapValue) String() string {
	return fmt.Sprintf("%v", v.Value)
}
