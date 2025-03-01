package tracker

import (
	"context"
	"encoding/binary"

	trHttp "github.com/al002/zbittorrent/internal/tracker/http"
	trackerTypes "github.com/al002/zbittorrent/internal/tracker/types"
	udpTracker "github.com/al002/zbittorrent/internal/tracker/udp"
)

type UdpClient struct {
	cl *udpTracker.Client
}

func (c *UdpClient) Close() error {
	return c.cl.Close()
}

func (c *UdpClient) Announce(
	ctx context.Context,
	req trackerTypes.AnnounceRequest,
	opts trHttp.AnnounceOpt,
) (res trHttp.AnnounceResponse, err error) {
	if req.IPAddress == 0 && opts.ClientIpV4 != nil {
		req.IPAddress = binary.BigEndian.Uint32(opts.ClientIpV4.To4())
	}

	h, nas, err := c.cl.Announce(ctx, req)
	if err != nil {
		return
	}

	res.Interval = h.Interval
	res.Leechers = h.Leechers
	res.Seeders = h.Seeders
	for _, cp := range nas.NodeAddrs() {
		res.Peers = append(res.Peers, trHttp.Peer{}.FromNodeAddr(cp))
	}

	return
}
