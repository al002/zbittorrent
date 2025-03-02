package metainfo

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTorrent(t *testing.T) {
	f, err := os.Open("testdata/ubuntu-24.04.2-desktop-amd64.iso.torrent")
	if err != nil {
		t.Fatal(err)
	}

	tor, err := New(f)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ubuntu-24.04.2-desktop-amd64.iso", tor.Info.Name)
	assert.Equal(t, int64(6343219200), tor.Info.Length)
	assert.Equal(t, "611f70899d4e1d6a9c39cfc925f103dfef630328", hex.EncodeToString(tor.Info.Hash[:]))
	assert.Equal(t, [][]string{
		{"https://torrent.ubuntu.com/announce"},
		{"https://ipv6.torrent.ubuntu.com/announce"},
	}, tor.AnnounceList)
}
