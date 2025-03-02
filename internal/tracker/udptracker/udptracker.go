package udptracker

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/al002/zbittorrent/internal/log"
	"github.com/al002/zbittorrent/internal/tracker"
)

type UDPTracker struct {
	rawURL    string
	dest      string
	urlData   string
	transport *Transport
	log       log.Logger
}

var _ tracker.Tracker = (*UDPTracker)(nil)

func New(rawURL string, u *url.URL, t *Transport) *UDPTracker {
	return &UDPTracker{
		rawURL:    rawURL,
		dest:      u.Host,
		urlData:   u.RequestURI(),
		transport: t,
	}
}

func (t *UDPTracker) URL() string {
	return t.rawURL
}

func (t *UDPTracker) Announce(ctx context.Context, req tracker.AnnounceRequest) (*tracker.AnnounceResponse, error) {
	announce := newTransportRequest(ctx, req, t.dest, t.urlData)

	reply, err := t.transport.Do(announce)
	if err != nil {
		return nil, err
	}

	response, peers, err := t.parseAnnounceResponse(reply)
	if err != nil {
		return nil, tracker.ErrDecode
	}

	return &tracker.AnnounceResponse{
		Interval: time.Duration(response.Interval) * time.Second,
		Leechers: response.Leechers,
		Seeders:  response.Seeders,
		Peers:    peers,
	}, nil
}

func (t *UDPTracker) parseAnnounceResponse(data []byte) (*udpAnnounceResponse, []*net.TCPAddr, error) {
	var response udpAnnounceResponse
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &response)
	if err != nil {
		return nil, nil, err
	}
	t.log.Debug(
		"announce response",
		"announce_response", response,
	)
	if response.Action != actionAnnounce {
		return nil, nil, errors.New("invalid action")
	}

	peers, err := tracker.DecodePeersCompact(data[binary.Size(response):])
	if err != nil {
		return nil, nil, err
	}

	return &response, peers, nil
}
