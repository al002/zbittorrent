package bencode

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	testCases := []struct {
		input          interface{}
		expectedOutput string
		expectedError  string
	}{
		{
			input:          nil,
			expectedOutput: "",
		},
		{
			input:          true,
			expectedOutput: "i1e",
		},
		{
			input:          false,
			expectedOutput: "i0e",
		},
		{
			input:          int(123),
			expectedOutput: "i123e",
		},
		{
			input:          int8(-42),
			expectedOutput: "i-42e",
		},
		{
			input:          int16(1024),
			expectedOutput: "i1024e",
		},
		{
			input:          int32(-65536),
			expectedOutput: "i-65536e",
		},
		{
			input:          int64(9223372036854775807),
			expectedOutput: "i9223372036854775807e",
		},
		{
			input:          uint(456),
			expectedOutput: "i456e",
		},
		{
			input:          uint8(255),
			expectedOutput: "i255e",
		},
		{
			input:          uint16(32768),
			expectedOutput: "i32768e",
		},
		{
			input:          uint32(4294967295),
			expectedOutput: "i4294967295e",
		},
		{
			input:          uint64(18446744073709551615),
			expectedOutput: "i18446744073709551615e",
		},
		{
			input:          "hello world",
			expectedOutput: "11:hello world",
		},
		{
			input:          []byte("bencode"),
			expectedOutput: "7:bencode",
		},
		{
			input:          []interface{}{1, "two", []byte("three")},
			expectedOutput: "li1e3:two5:threee",
		},
		{
			input:          map[string]interface{}{"a": 1, "b": "two"},
			expectedOutput: "d1:ai1e1:b3:twoe",
		},
		{
			input: map[string]interface{}{
				"list":  []int{1, 2, 3},
				"value": "test",
			},
			expectedOutput: "d4:listli1ei2ei3ee5:value4:teste",
		},
		{
			input: map[string]interface{}{
				"nested": map[string]int{"x": 10, "y": 20},
				"string": "example",
			},
			expectedOutput: "d6:nestedd1:xi10e1:yi20ee6:string7:examplee",
		},
		{
			input:          big.NewInt(1234567890),
			expectedOutput: "i1234567890e",
		},
		{
			input: struct {
				Name   string `bencode:"name"`
				Age    int    `bencode:"age"`
				OmitMe string `bencode:"omit_me,omitempty"`
			}{"Alice", 30, ""},
			expectedOutput: "d4:name5:Alice3:agei30ee",
		},
		{
			input: struct {
				Name string `bencode:"name"`
				Age  int    `bencode:"age,omitempty"`
			}{"Bob", 0},
			expectedOutput: "d4:name3:Bobe",
		},
	}

	for i, tc := range testCases {
		testName := fmt.Sprintf("TestCase-%d", i)
		t.Run(testName, func(t *testing.T) {
			buffer := new(bytes.Buffer)
			encoder := Encoder{w: buffer}
			err := encoder.Encode(tc.input)

			if tc.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Fatalf("expected error %q, got %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if buffer.String() != tc.expectedOutput {
					t.Fatalf("expected output %q, got %q", tc.expectedOutput, buffer.String())
				}
			}
		})
	}
}
