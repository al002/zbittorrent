package torrent

import (
	"sync"
	"time"

	"github.com/al002/zbittorrent/internal/blocklist"
)

type Session struct {
	Config Config

	mTorrents sync.RWMutex
	torrents  map[string]*Torrent

	mPorts         sync.RWMutex
	availablePorts map[int]struct{}

	mBlocklist         sync.RWMutex
	blocklist          *blocklist.Blocklist
	blocklistTimestamp time.Time

	closeC chan struct{}
}
