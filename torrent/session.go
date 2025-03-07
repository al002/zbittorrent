package torrent

import (
	"errors"
	"sync"
	"time"

	"github.com/al002/zbittorrent/internal/blocklist"
	"github.com/al002/zbittorrent/internal/log"
	"github.com/al002/zbittorrent/internal/trackermanager"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/time/rate"
)

type Session struct {
	config Config

	trackerManager  *trackermanager.TrackerManager
	log             log.Logger
	downloadLimiter *rate.Limiter
	uploadLimiter   *rate.Limiter

	mTorrents sync.RWMutex
	torrents  map[string]*Torrent

	mPorts         sync.RWMutex
	availablePorts map[int]struct{}

	mBlocklist         sync.RWMutex
	blocklist          *blocklist.Blocklist
	blocklistTimestamp time.Time

	closeC chan struct{}
}

func NewSession(cfg Config, logger log.Logger) (*Session, error) {
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
		log:            logger,
		blocklist:      bl,
		trackerManager: trackermanager.New(blTracker, cfg.DNSResolveTimeout, !cfg.TrackerHTTPVerifyTLS, logger),
		torrents:       make(map[string]*Torrent),
		availablePorts: ports,
		closeC:         make(chan struct{}),
	}

	dlSpeed := cfg.SpeedLimitDownload * 1024
	if cfg.SpeedLimitDownload > 0 {
		c.downloadLimiter = rate.NewLimiter(rate.Limit(dlSpeed), int(dlSpeed))
	}
	ulSpeed := cfg.SpeedLimitUpload * 1024
	if cfg.SpeedLimitUpload > 0 {
		c.uploadLimiter = rate.NewLimiter(rate.Limit(ulSpeed), int(ulSpeed))
	}

	return c, nil
}

func (s *Session) Close() {
	close(s.closeC)

	var wg sync.WaitGroup
	s.mTorrents.Lock()
	wg.Add(len(s.torrents))
	for _, t := range s.torrents {
		go func(t *Torrent) {
			t.torrent.Close()
			wg.Done()
		}(t)
	}
	wg.Wait()
	s.torrents = nil
	s.mTorrents.Unlock()

	s.trackerManager.Close()
}

func (s *Session) getTrackerUserAgent(private bool) string {
	if private {
		return s.config.TrackerHTTPPrivateUserAgent
	}

	return trackerHTTPPublicUserAgent
}

func (s *Session) getPort() (int, error) {
  s.mPorts.Lock()
  defer s.mPorts.Unlock()
  for p := range s.availablePorts {
    delete(s.availablePorts, p)
    return p, nil
  }

  return 0, errors.New("no free port")
}

func (s *Session) releasePort(port int) {
  s.mPorts.Lock()
  defer s.mPorts.Unlock()
  s.availablePorts[port] = struct{}{}
}
