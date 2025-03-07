package announcer

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/al002/zbittorrent/internal/resolver"
	"github.com/al002/zbittorrent/internal/tracker"
	"github.com/al002/zbittorrent/internal/tracker/httptracker"
	"github.com/cenkalti/backoff/v5"
)

type Status int

const (
	NotContactedYet Status = iota
	Contacting
	Working
	NotWorking
)

// Announces the torrent to the tracker periodically
type PeriodicalAnnouncer struct {
	Tracker        tracker.Tracker
	numWant        int
	interval       time.Duration
	minInterval    time.Duration
	seeders        int
	leechers       int
	warningMsg     string
	completedC     chan struct{}
	newPeersC      chan []*net.TCPAddr
	backoff        backoff.BackOff
	getTorrent     func() tracker.Torrent
	lastAnnounce   time.Time
	nextAnnounce   time.Time
	HasAnnounced   bool
	responseC      chan *tracker.AnnounceResponse
	errC           chan error
	closeC         chan struct{}
	doneC          chan struct{}
	needMorePeers  bool
	mNeedMorePeers sync.RWMutex
	needMorePeersC chan struct{}
	lastError      *AnnounceError

	status        Status
	statsCommandC chan statsRequest
}

func NewPeriodicalAnnouncer(t tracker.Tracker, numWant int, minInterval time.Duration, getTorrent func() tracker.Torrent, completedC chan struct{}, newPeersC chan []*net.TCPAddr) *PeriodicalAnnouncer {
	return &PeriodicalAnnouncer{
		Tracker:        t,
		status:         NotContactedYet,
		numWant:        numWant,
		minInterval:    minInterval,
		completedC:     completedC,
		newPeersC:      newPeersC,
		getTorrent:     getTorrent,
		needMorePeersC: make(chan struct{}, 1),
		responseC:      make(chan *tracker.AnnounceResponse),
		errC:           make(chan error),
		closeC:         make(chan struct{}),
		doneC:          make(chan struct{}),
		statsCommandC:  make(chan statsRequest),
		backoff: &backoff.ExponentialBackOff{
			InitialInterval:     5 * time.Second,
			RandomizationFactor: 0.5,
			Multiplier:          2,
			MaxInterval:         30 * time.Minute,
		},
	}
}

func (a *PeriodicalAnnouncer) Close() {
	close(a.closeC)
	<-a.doneC
}

func (a *PeriodicalAnnouncer) Run() {
	defer close(a.doneC)

	a.backoff.Reset()

	timer := time.NewTimer(math.MaxInt64)
	defer timer.Stop()

	resetTimer := func(interval time.Duration) {
		timer.Reset(interval)
		if interval < 0 {
			a.nextAnnounce = time.Now()
		} else {
			a.nextAnnounce = time.Now().Add(interval)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// BEP 0003: No completed is sent if the file was complete when started.
	select {
	case <-a.completedC:
		a.completedC = nil
	default:
	}

	a.doAnnounce(ctx, tracker.EventStarted, a.numWant)
	for {
		select {
		case <-timer.C:
			if a.status == Contacting {
				break
			}
			a.doAnnounce(ctx, tracker.EventNone, a.numWant)
		case resp := <-a.responseC:
			fmt.Printf("announce responsed %+v\n", resp)
			a.status = Working
			a.seeders = int(resp.Seeders)
			a.leechers = int(resp.Leechers)
			a.warningMsg = resp.WarningMessage
			if a.warningMsg != "" {

			}
			a.interval = resp.Interval
			if resp.MinInterval > 0 {
				a.minInterval = resp.MinInterval
			}
			a.HasAnnounced = true
			a.lastError = nil
			a.backoff.Reset()
			interval := a.getNextInterval()
			resetTimer(interval)
			go func() {
				select {
				case a.newPeersC <- resp.Peers:
				case <-a.closeC:
				}
			}()
		case err := <-a.errC:
			a.status = NotWorking
			a.lastError = a.newAnnounceError(err)
			if a.lastError.Unknown {

			} else {

			}
			interval := a.getNextIntervalFromError(a.lastError)
			resetTimer(interval)
		case <-a.needMorePeersC:
			if a.status == Contacting || a.status == NotWorking {
				break
			}
			interval := time.Until(a.lastAnnounce.Add(a.getNextInterval()))
			resetTimer(interval)
		case <-a.completedC:
			if a.status == Contacting {
				cancel()
				ctx, cancel = context.WithCancel(context.Background())
			}
			a.doAnnounce(ctx, tracker.EventCompleted, 0)
			a.completedC = nil
		case req := <-a.statsCommandC:
			req.Response <- a.stats()
		case <-a.closeC:
			cancel()
			return
		}
	}
}

func (a *PeriodicalAnnouncer) NeedMorePeers(val bool) {
	a.mNeedMorePeers.Lock()
	a.needMorePeers = val
	a.mNeedMorePeers.Unlock()

	select {
	case a.needMorePeersC <- struct{}{}:
	case <-a.doneC:
	default:
	}
}

func (a *PeriodicalAnnouncer) doAnnounce(ctx context.Context, event tracker.Event, numWant int) {
	go a.announce(ctx, event, numWant)
	a.status = Contacting
	a.lastAnnounce = time.Now()
}

func (a *PeriodicalAnnouncer) announce(ctx context.Context, event tracker.Event, numWant int) {
	announce(ctx, a.Tracker, event, numWant, a.getTorrent(), a.responseC, a.errC)
}

func (a *PeriodicalAnnouncer) getNextInterval() time.Duration {
	a.mNeedMorePeers.RLock()
	need := a.needMorePeers
	a.mNeedMorePeers.RUnlock()

	if need {
		return a.minInterval
	}

	return a.interval
}

func (a *PeriodicalAnnouncer) getNextIntervalFromError(err *AnnounceError) time.Duration {
	if terr, ok := err.Err.(*tracker.Error); ok && terr.RetryIn > 0 {
		return terr.RetryIn
	}

	return a.backoff.NextBackOff()
}

type statsRequest struct {
	Response chan Stats
}

func (a *PeriodicalAnnouncer) Stats() Stats {
	var stats Stats
	req := statsRequest{
		Response: make(chan Stats, 1),
	}

	select {
	case a.statsCommandC <- req:
	case <-a.closeC:
	}

	select {
	case stats = <-req.Response:
	case <-a.closeC:
	}

	return stats
}

type Stats struct {
	Status       Status
	Error        *AnnounceError
	Warning      string
	Seeders      int
	Leechers     int
	LastAnnounce time.Time
	NextAnnounce time.Time
}

func (a *PeriodicalAnnouncer) stats() Stats {
	return Stats{
		Status:       a.status,
		Error:        a.lastError,
		Warning:      a.warningMsg,
		Seeders:      a.seeders,
		Leechers:     a.leechers,
		LastAnnounce: a.lastAnnounce,
		NextAnnounce: a.nextAnnounce,
	}
}

type AnnounceError struct {
	Err     error
	Message string
	Unknown bool
}

func (a *PeriodicalAnnouncer) newAnnounceError(err error) (e *AnnounceError) {
	e = &AnnounceError{Err: err}
	switch err {
	case resolver.ErrNotIpv4Address:
		parsed, _ := url.Parse(a.Tracker.URL())
		e.Message = fmt.Sprintf("tracker has no IPv4 address: %s", parsed.Hostname())
	case resolver.ErrBlocked:
		e.Message = "tracker IP is blocked"
		return
	case resolver.ErrInvalidPort:
		parsed, _ := url.Parse(a.Tracker.URL())
		e.Message = fmt.Sprintf("invalid port number in tracker address: %s", parsed.Host)
		return
	case tracker.ErrDecode:
		e.Message = "invalid response from tracker"
		return
	}

	if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
		e.Message = "contactingg tracker timeout"
		return
	}

	switch err := err.(type) {
	case *net.DNSError:
		s := err.Error()
		if strings.HasSuffix(s, "no such host") {
			e.Message = "host not found: " + err.Name
			return
		}
		if strings.HasSuffix(s, "server misbehaving") {
			e.Message = "host not found: " + err.Name
			return
		}
		if strings.HasSuffix(s, "Temporary failure in name resolution") {
			e.Message = "temporary failure in name resolution: " + err.Name
			return
		}
		if strings.HasSuffix(s, "No address associated with hostname") {
			e.Message = "no address associated with hostname: " + err.Name
			return
		}
	case *net.AddrError:
		s := err.Error()
		if strings.HasSuffix(s, "missing port in address") {
			e.Message = "missing port in tracker address"
			return
		}
	case *url.Error:
		s := err.Error()
		if strings.HasSuffix(s, "connection refused") {
			e.Message = "tracker refused the connection"
			return
		}
		if strings.HasSuffix(s, "no such host") {
			parsed, _ := url.Parse(a.Tracker.URL())
			e.Message = "no such host: " + parsed.Hostname()
			return
		}
		if strings.HasSuffix(s, "server misbehaving") {
			parsed, _ := url.Parse(a.Tracker.URL())
			e.Message = "server misbehaving: " + parsed.Hostname()
			return
		}
		if strings.HasSuffix(s, "tls: handshake failure") {
			e.Message = "TLS handshake has failed"
			return
		}
		if strings.HasSuffix(s, "no route to host") {
			parsed, _ := url.Parse(a.Tracker.URL())
			e.Message = "no route to host: " + parsed.Hostname()
			return
		}
		if strings.HasSuffix(s, "No address associated with hostname") {
			parsed, _ := url.Parse(a.Tracker.URL())
			e.Message = "no address associated with hostname: " + parsed.Hostname()
			return
		}
		if strings.HasSuffix(s, resolver.ErrNotIpv4Address.Error()) {
			parsed, _ := url.Parse(a.Tracker.URL())
			e.Message = "tracker has no IPv4 address: " + parsed.Hostname()
			return
		}
		if strings.HasSuffix(s, "connection reset by peer") {
			e.Message = "tracker closed the connection"
			return
		}
		if strings.HasSuffix(s, "EOF") {
			e.Message = "tracker closed the connection"
			return
		}
		if strings.HasSuffix(s, "server gave HTTP response to HTTPS client") {
			e.Message = "invalid server response"
			return
		}
		if strings.Contains(s, "malformed HTTP status code") {
			e.Message = "invalid server response"
			return
		}
		if strings.Contains(s, "network is unreachable") {
			parsed, _ := url.Parse(a.Tracker.URL())
			e.Message = "network is unreachable: " + parsed.Hostname()
			return
		}
	case *httptracker.StatusError:
		if err.Code >= 400 {
			e.Message = fmt.Sprintf("tracker returned HTTP status: %s", strconv.Itoa(err.Code))
			if err.Header.Get("content-type") == "text/plain" {
				msg := err.Body
				if len(msg) > 100 {
					msg = msg[:97] + "..."
				}
				e.Message += " message: " + msg
			}
			return
		}
	case *tracker.Error:
		e.Message = "announce error: " + err.FailureReason
		return
	}

	e.Message = "Unknown error in announce"
	e.Unknown = true
	return
}
