package metainfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateName(t *testing.T) {
	cases := []struct {
		name    string
		truncated string
		max     int
	}{
		{"foo.bar", "foo.bar", 10},
		{"foo.bar", "foo.bar", 7},
		{"foo.bar", "fo.bar", 6},
		{"foo.bar", ".bar", 4},
		{"foo.bar", "foo", 3},
		{"foobar", "foobar", 10},
		{"foobar", "fo", 2},
		{"ğğğğ", "ğğğğ", 9},
		{"ğğğğ", "ğğğğ", 8},
		{"ğğğğ", "ğğğ", 7},
		{"ğğğğ", "ğğğ", 6},
	}
	for _, c := range cases {
		assert.Equal(t, c.truncated, truncateNameN(c.name, c.max))
	}
}
