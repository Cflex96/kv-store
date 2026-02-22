package store

import "fmt"

type Type interface {
	Type() string
	String() string
}

type StringValue struct {
	Value string
}

func (v *StringValue) Type() string {
	return "string"
}

func (v *StringValue) String() string {
	return v.Value
}

type ListValue struct {
	Value []string
}

func (v *ListValue) Type() string {
	return "string"
}

func (v *ListValue) String() string {
	return fmt.Sprintf("%v", v.Value)
}

type MapValue struct {
	Value map[string]string
}

func (v *MapValue) Type() string {
	return "map"
}

func (v *MapValue) String() string {
	return fmt.Sprintf("%v", v.Value)
}
