package bencode

import (
	"reflect"
	"strings"
)

func getTag(st reflect.StructTag) tag {
	return parseTag(st.Get("bencode"))
}

type tag []string

func parseTag(tagStr string) tag {
	return strings.Split(tagStr, ",")
}

func (t tag) Ignore() bool {
	return t[0] == "-"
}

func (t tag) Key() string {
	return t[0]
}

func (t tag) HasOpt(opt string) bool {
	if len(t) < 1 {
		return false
	}
	for _, s := range t[1:] {
		if s == opt {
			return true
		}
	}
	return false
}

func (t tag) OmitEmpty() bool {
	return t.HasOpt("omitempty")
}

func (t tag) IgnoreUnmarshalTypeError() bool {
	return t.HasOpt("ignore_unmarshal_type_error")
}
