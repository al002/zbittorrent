package infohash

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

const Size = 20

// 20-byte SHA1 hash used for info and pieces.
type T [Size]byte

func HashBytes(b []byte) (ret T) {
	hasher := sha1.New()
	hasher.Write(b)
	copy(ret[:], hasher.Sum(nil))
	return
}

func (t T) String() string {
	return t.HexString()
}

func (t T) HexString() string {
	return fmt.Sprintf("%x", t[:])
}

func (t T) FromHexString(s string) (err error) {
	err = fmt.Errorf("hash hex string has bad length: %d", len(s))
	if len(s) != 2*Size {
		return
	}

	n, err := hex.Decode(t[:], []byte(s))

	if err != nil {
		return
	}

	if n != Size {
		return fmt.Errorf("hex.Decode decoded %d bytes, expected %d", n, Size)
	}

	return
}

func FromHexString(s string) (h T, err error) {
	err = h.FromHexString(s)
	if err != nil {
		return T{}, err
	}

	return h, nil
}
