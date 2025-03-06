package httptracker

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/al002/zbittorrent/internal/log"
	"github.com/al002/zbittorrent/internal/tracker"
	"github.com/al002/zbittorrent/pkg/bencode"
)

type HTTPTracker struct {
	rawURL            string
	http              *http.Client
	transport         *http.Transport
	trackerID         string
	userAgent         string
	maxResponseLength int64
	log               log.Logger
}

var _ tracker.Tracker = (*HTTPTracker)(nil)

func New(rawURL string, u *url.URL, timeout time.Duration, t *http.Transport, userAgent string, maxResponseLength int64, log log.Logger) *HTTPTracker {
	return &HTTPTracker{
		rawURL:            rawURL,
		transport:         t,
		userAgent:         userAgent,
		maxResponseLength: maxResponseLength,
		http: &http.Client{
			Timeout:   timeout,
			Transport: t,
		},
    log: log,
	}
}

func (t *HTTPTracker) URL() string {
	return t.rawURL
}

func (t *HTTPTracker) Announce(ctx context.Context, req tracker.AnnounceRequest) (*tracker.AnnounceResponse, error) {
	s := t.buildRequest(req)

	t.log.Debug(
		"announce request",
		"request_str", s,
	)

	httpReq, err := http.NewRequest(http.MethodGet, s, nil)

	if err != nil {
		return nil, err
	}

	httpReq = httpReq.WithContext(ctx)

	httpReq.Header.Set("User-Agent", t.userAgent)

	doReq := func() (int, http.Header, []byte, error) {
		resp, err := t.http.Do(httpReq)
		if err != nil {
			return 0, nil, nil, err
		}

		t.log.Debug(
			"tracker response",
			"resp_code", resp.StatusCode,
			"content_length", resp.ContentLength,
		)

		defer resp.Body.Close()

		if resp.ContentLength > t.maxResponseLength {
			return 0, resp.Header, nil, fmt.Errorf("tracker response too large: %d", resp.ContentLength)
		}

		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)

		return resp.StatusCode, resp.Header, buf.Bytes(), err
	}

	code, header, body, err := doReq()
	if uerr, ok := err.(*url.Error); ok && uerr.Err == context.Canceled {
		return nil, context.Canceled
	}

	if err != nil {
		return nil, err
	}
	t.log.Debug(
		"read bytes from body",
		"body_length", len(body),
	)

	var response announceResponse

	err = bencode.Unmarshal(body, &response)
	if err != nil {
		if code != 200 {
			return nil, &StatusError{
				Code:   code,
				Header: header,
				Body:   string(body[:]),
			}
		}
		return nil, tracker.ErrDecode
	}

	if response.FailureReason != "" {
		retryIn, _ := strconv.Atoi(response.RetryIn)
		return nil, &tracker.Error{
			FailureReason: response.FailureReason,
			RetryIn:       time.Duration(retryIn) * time.Minute,
		}
	}

	if response.TrackerID != "" {
		t.trackerID = response.TrackerID
	}

	var peers []*net.TCPAddr

	if len(response.Peers) > 0 {
		if response.Peers[0] == 'l' {
			// non-compact peers
			peers, err = parsePeersDictionary(response.Peers)
		} else {
			var b []byte
			err = bencode.Unmarshal(response.Peers, &b)
			if err != nil {
				return nil, tracker.ErrDecode
			}
			peers, err = tracker.DecodePeersCompact(b)
		}
	}

	if err != nil {
		return nil, err
	}

	t.log.Debug(
		"got peers",
		"peers_length", len(peers),
	)

	if len(response.ExternalIP) != 0 {
		var filtered int
		for i, p := range peers {
			if !bytes.Equal(p.IP[:], response.ExternalIP) {
				peers[i] = p
				filtered++
			}
		}

		peers = peers[:filtered]
	}

	return &tracker.AnnounceResponse{
		Interval:       time.Duration(response.Interval) * time.Second,
		MinInterval:    time.Duration(response.MinInterval) * time.Second,
		Leechers:       response.Incomplete,
		Seeders:        response.Complete,
		Peers:          peers,
		WarningMessage: response.WarningMessage,
	}, nil
}

func (t *HTTPTracker) buildRequest(req tracker.AnnounceRequest) string {
	var sb strings.Builder
	sb.WriteString(t.rawURL)

	if strings.ContainsRune(t.rawURL, '?') {
		sb.WriteString("&info_hash=")
	} else {
		sb.WriteString("?info_hash=")
	}
	sb.WriteString(percentEscape(req.Torrent.InfoHash))

	sb.WriteString("&peer_id=")
	sb.WriteString(percentEscape(req.Torrent.PeerID))

	sb.WriteString("&port=")
	sb.WriteString(strconv.Itoa(req.Torrent.Port))

	sb.WriteString("&uploaded=")
	sb.WriteString(strconv.FormatInt(req.Torrent.BytesUploaded, 10))

	sb.WriteString("&downloaded=")
	sb.WriteString(strconv.FormatInt(req.Torrent.BytesDownlowded, 10))

	sb.WriteString("&left=")
	sb.WriteString(strconv.FormatInt(req.Torrent.BytesLeft, 10))

	sb.WriteString("&compact=1")
	sb.WriteString("&no_peer_id=1")

	sb.WriteString("&num_want=")
	sb.WriteString(strconv.Itoa(req.NumWant))

	if req.Event != tracker.EventNone {
		sb.WriteString("&event=")
		sb.WriteString(req.Event.String())
	}

	if t.trackerID != "" {
		sb.WriteString("&trackerid=")
		sb.WriteString(t.trackerID)
	}

	sb.WriteString("&key=")
	sb.WriteString(hex.EncodeToString(req.Torrent.PeerID[16:20]))

	return sb.String()
}

func percentEscape(b [20]byte) string {
	var sb strings.Builder
	sb.Grow(60)
	s := hex.EncodeToString(b[:])
	for i := 0; i < 20; i++ {
		sb.WriteRune('%')
		sb.WriteByte(s[i*2])
		sb.WriteByte(s[i*2+1])
	}

	return sb.String()
}

func parsePeersDictionary(b bencode.Bytes) ([]*net.TCPAddr, error) {
	var peers []struct {
		IP   string `bencode:"ip"`
		Port uint16 `bencode:"port"`
	}
	err := bencode.Unmarshal(b, &peers)
	if err != nil {
		return nil, tracker.ErrDecode
	}

	addrs := make([]*net.TCPAddr, len(peers))

	for i, p := range peers {
		pe := &net.TCPAddr{
			IP:   net.ParseIP(p.IP),
			Port: int(p.Port),
		}
		addrs[i] = pe
	}

	return addrs, err
}
