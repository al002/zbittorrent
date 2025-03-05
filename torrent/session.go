package torrent

import (
	"errors"
	"sync"
	"time"

	"github.com/al002/zbittorrent/internal/blocklist"
	"github.com/al002/zbittorrent/internal/trackermanager"
	"github.com/mitchellh/go-homedir"
)

type Session struct {
	config Config

	trackerManager *trackermanager.TrackerManager

	peerID    [20]byte
	mTorrents sync.RWMutex
	torrents  map[string]*Torrent

	mPorts         sync.RWMutex
	availablePorts map[int]struct{}

	mBlocklist         sync.RWMutex
	blocklist          *blocklist.Blocklist
	blocklistTimestamp time.Time

	closeC chan struct{}
}

func NewSession(cfg Config) (*Session, error) {
	if cfg.PortBegin >= cfg.PortEnd {
		return nil, errors.New("Invalid port range")
	}

	// if cfg.MaxOpenFiles > 0 {
	// }

	var err error
	cfg.DataDir, err = homedir.Expand(cfg.DataDir)
	if err != nil {
		return nil, err
	}

	ports := make(map[int]struct{})
	for p := cfg.PortBegin; p < cfg.PortEnd; p++ {
		ports[int(p)] = struct{}{}
	}

	bl := blocklist.New()
	var blTracker *blocklist.Blocklist
	if cfg.BlocklistEnabledForTrackers {
		blTracker = bl
	}

	c := &Session{
		config:         cfg,
		blocklist:      bl,
		trackerManager: trackermanager.New(blTracker, cfg.DNSResolveTimeout, !cfg.TrackerHTTPVerifyTLS),
		torrents:       make(map[string]*Torrent),
		availablePorts: ports,
		closeC:         make(chan struct{}),
	}

	return c, nil
}

func (s *Session) getTrackerUserAgent(private bool) string {
	if private {
		return s.config.TrackerHTTPPrivateUserAgent
	}

	return trackerHTTPPublicUserAgent
}
