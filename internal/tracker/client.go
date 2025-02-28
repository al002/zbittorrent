package tracker

import (
	"context"
	"errors"
	"net"
	"net/url"

	"github.com/al002/zbittorrent/internal/logger"
	trHttp "github.com/al002/zbittorrent/internal/tracker/http"
	trackerTypes "github.com/al002/zbittorrent/internal/tracker/types"
	udpTracker "github.com/al002/zbittorrent/internal/tracker/udp"
)

var ErrBadScheme = errors.New("unknown scheme")

type AnnounceOpt = trHttp.AnnounceOpt

type TrackerClient interface {
	Announce(context.Context, trackerTypes.AnnounceRequest, AnnounceOpt) (trHttp.AnnounceResponse, error)
	Close() error
}

type NewTrackerClientOpts struct {
	Http         trHttp.NewClientOpts
	Logger       logger.Logger
	ListenPacket func(network, addr string) (net.PacketConn, error)
}

func NewTrackerClient(urlStr string, opts NewTrackerClientOpts) (TrackerClient, error) {
	_url, err := url.Parse(urlStr)

	if err != nil {
		return nil, err
	}

	switch _url.Scheme {
	case "http", "https":
		return trHttp.NewClient(_url, opts.Http), nil
	case "udp", "udp4", "udp6":
		network := _url.Scheme
		c, err := udpTracker.NewClient(udpTracker.ClientOpts{
			Network:       network,
			Addr:          _url.Host,
			Logger:        opts.Logger,
			ListenePacket: opts.ListenPacket,
		})

		if err != nil {
			return nil, err
		}

		return &UdpClient{
			cl: c,
		}, nil
	default:
		return nil, ErrBadScheme
	}
}
