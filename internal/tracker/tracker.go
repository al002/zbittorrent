package tracker

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/al002/zbittorrent/internal/logger"
	trHttp "github.com/al002/zbittorrent/internal/tracker/http"
	trackerTypes "github.com/al002/zbittorrent/internal/tracker/types"
	"github.com/anacrolix/dht/v2/krpc"
)

type Tracker struct {
	Url          string
	Request      trackerTypes.AnnounceRequest
	HostHeader   string
	HttpProxy    func(*http.Request) (*url.URL, error)
	DialContext  func(ctx context.Context, network, addr string) (net.Conn, error)
	ListenPacket func(network, addr string) (net.PacketConn, error)
	ServeName    string
	UserAgent    string
	UdpNetwork   string
	ClientIpV4   krpc.NodeAddr
	ClientIpV6   krpc.NodeAddr
	Context      context.Context
	Logger       logger.Logger
}

const DefaultTrackerAnnounceTimeout = 15 * time.Second

func (t Tracker) Announce() (res trHttp.AnnounceResponse, err error) {
	cl, err := NewTrackerClient(t.Url, NewTrackerClientOpts{
		Http: trHttp.NewClientOpts{
			Proxy:       t.HttpProxy,
			DialContext: t.DialContext,
			ServerName:  t.ServeName,
		},
		// TODO: support context
		Logger:       t.Logger,
		ListenPacket: t.ListenPacket,
	})

	if err != nil {
		return
	}

	defer cl.Close()

	if t.Context == nil {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultTrackerAnnounceTimeout)
		defer cancel()
		t.Context = ctx
	}

	return cl.Announce(t.Context, t.Request, trHttp.AnnounceOpt{
		UserAgent:  t.UserAgent,
		HostHeader: t.HostHeader,
		ClientIpV4: t.ClientIpV4.IP,
		ClientIpV6: t.ClientIpV6.IP,
	})
}
