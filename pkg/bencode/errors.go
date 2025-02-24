package bencode

import (
	"fmt"
	"reflect"
)

// float32/float64 has no bencode representation
type MarshalTypeError struct {
	Type reflect.Type
}

func (e *MarshalTypeError) Error() string {
	return fmt.Sprintf("bencode: unsupported type: %s", e.Type.String())
}

// Argument must be a non-nil value of some pointer type
type UnmarshalInvalidArgError struct {
	Type reflect.Type
}

func (e *UnmarshalInvalidArgError) Error() string {
	if e.Type == nil {
		return "bencode: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return fmt.Sprintf("bencode: Unmarshal(non-pointer %s)", e.Type.String())
	}

	return fmt.Sprintf("bencode: Unmarshal(nil %s)", e.Type.String())
}

// A value that was not a appropriate Go value
type UnmarshalTypeError struct {
	BencodeTypeName     string
	UnmarshalTargetType reflect.Type
}

func (e *UnmarshalTypeError) Error() string {
	return fmt.Sprintf(
		"bencode: cannot unmarshal a bencode %v into a %v",
		e.BencodeTypeName,
		e.UnmarshalTargetType,
	)
}

type UnmarshalFieldError struct {
	Key   string
	Type  reflect.Type
	Field reflect.StructField
}

func (e *UnmarshalFieldError) Error() string {
	return fmt.Sprintf("bencode: key \"%s\" led to an unexported field \"%s\" in type: \"%s\"", e.Key, e.Field.Name, e.Type.String())
}

// Malformed bencode input
type SyntaxError struct {
	Offset int64
	Err    error
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("bencode: syntax error offset: %d: %s", e.Offset, e.Err)
}

type ErrUnusedTrailingBytes struct {
  UnusedBytes int
}

func (e ErrUnusedTrailingBytes) Error() string {
  return fmt.Sprintf("%d unused trailing bytes", e.UnusedBytes)
}
