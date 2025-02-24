package bencode

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeParseValue(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    interface{}
		expectedErr error
	}{
		{
			name:     "Integer",
			input:    "i123e",
			expected: int64(123),
		},
		{
			name:     "Negative Integer",
			input:    "i-456e",
			expected: int64(-456),
		},
		{
			name:     "String",
			input:    "5:hello",
			expected: "hello",
		},
		{
			name:     "Empty String",
			input:    "0:",
			expected: "",
		},
		{
			name:     "List",
			input:    "l4:spam4:eggse",
			expected: []interface{}{"spam", "eggs"},
		},
		{
			name:     "Empty List",
			input:    "le",
			expected: []interface{}{},
		},
		{
			name:  "Dictionary",
			input: "d3:foo3:bar5:helloi123ee",
			expected: map[string]interface{}{
				"foo":   "bar",
				"hello": int64(123),
			},
		},
		{
			name:     "Empty Dictionary",
			input:    "de",
			expected: map[string]interface{}{},
		},
		{
			name:        "Invalid Input - Unexpected EOF",
			input:       "i12",
			expectedErr: &SyntaxError{},
		},
		{
			name:        "Invalid Input - Non-string Dictionary Key",
			input:       "di123e3:bare",
			expectedErr: &SyntaxError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := &Decoder{r: strings.NewReader(tc.input)}
			var result interface{}
			err := decoder.Decode(&result)

			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error, but got nil")
				}
				if !errorAs(err, &tc.expectedErr) {
					t.Fatalf("Expected error type %T, but got %T", tc.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected result:\n%#v\ngot:\n%#v", tc.expected, result)
			}
		})
	}
}

func errorAs(err error, target interface{}) bool {
	return errors.As(err, target)
}

func TestDecodeInvalidArg(t *testing.T) {
	decoder := &Decoder{r: bytes.NewReader([]byte("i1e"))}

	err := decoder.Decode(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	target := new(UnmarshalInvalidArgError)
	if !errors.As(err, &target) {
		t.Fatalf("expected UnmarshalInvalidArgError, got %T", err)
	}
}
