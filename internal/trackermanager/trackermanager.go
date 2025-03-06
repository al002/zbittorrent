package trackermanager

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/al002/zbittorrent/internal/blocklist"
	"github.com/al002/zbittorrent/internal/log"
	"github.com/al002/zbittorrent/internal/resolver"
	"github.com/al002/zbittorrent/internal/tracker"
	"github.com/al002/zbittorrent/internal/tracker/httptracker"
	"github.com/al002/zbittorrent/internal/tracker/udptracker"
)

type TrackerManager struct {
	httpTransport *http.Transport
	udpTransport  *udptracker.Transport
  log log.Logger
}

func New(bl *blocklist.Blocklist, dnsTimeout time.Duration, tlsSkipVerify bool, logger log.Logger) *TrackerManager {
	m := &TrackerManager{
		httpTransport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify},
		},
		udpTransport: udptracker.NewTransport(bl, dnsTimeout, logger),
    log: logger,
	}

	go m.udpTransport.Run()

	m.httpTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		ip, port, err := resolver.Resolve(ctx, addr, dnsTimeout, bl)
		if err != nil {
			return nil, err
		}

		var d net.Dialer
		taddr := &net.TCPAddr{
			IP:   ip,
			Port: port,
		}

		return d.DialContext(ctx, network, taddr.String())
	}

	return m
}

func (m *TrackerManager) Close() {
	m.httpTransport.CloseIdleConnections()
	m.udpTransport.Close()
}

func (m *TrackerManager) Get(s string, httpTimeout time.Duration, httpUserAgent string, httpMaxResponseLength int64) (tracker.Tracker, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		tr := httptracker.New(s, u, httpTimeout, m.httpTransport, httpUserAgent, httpMaxResponseLength, m.log)
		return tr, nil
	case "udp":
		tr := udptracker.New(s, u, m.udpTransport, m.log)
		return tr, nil
	default:
		return nil, fmt.Errorf("unsupported tracker scheme: %s", u.Scheme)
	}
}
