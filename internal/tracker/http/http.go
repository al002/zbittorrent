package httpTracker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	trackerTypes "github.com/al002/zbittorrent/internal/tracker/types"
	"github.com/al002/zbittorrent/internal/version"
	"github.com/al002/zbittorrent/pkg/bencode"
	"github.com/al002/zbittorrent/pkg/utils"
)

type AnnounceResponse struct {
	Interval int32
	Seeders  int32
	Leechers int32
	Peers    []Peer
}

type AnnounceOpt struct {
	UserAgent  string
	HostHeader string
	ClientIpV4 net.IP
	ClientIpV6 net.IP
}

func (cl Client) Announce(ctx context.Context, req trackerTypes.AnnounceRequest, opt AnnounceOpt) (ret AnnounceResponse, err error) {
	_url := utils.CopyURL(cl.url_)

	setAnnounceParams(_url, &req)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, _url.String(), nil)
	if err != nil {
		return
	}

	userAgent := opt.UserAgent

	if userAgent == "" {
		userAgent = version.DefaultHttpUserAgent
	}

	if userAgent != "" {
		httpReq.Header.Set("User-Agent", userAgent)
	}

	httpReq.Host = opt.HostHeader

	resp, err := cl.hc.Do(httpReq)

	if err != nil {
		err = fmt.Errorf("HTTP request failed: %w", err)
		return
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)

	if resp.StatusCode != 200 {
		err = fmt.Errorf("response not ok: %s: %q", resp.Status, buf.Bytes())
		return
	}

	var httpResponse HttpTrackerResponse
	err = bencode.Unmarshal(buf.Bytes(), &httpResponse)
	if _, ok := err.(bencode.ErrUnusedTrailingBytes); ok {
		err = nil
	} else if err != nil {
		err = fmt.Errorf("error decoding %q: %s", buf.Bytes(), err)
		return
	}

	if httpResponse.FailureReason != "" {
		err = fmt.Errorf("tracker gave failure reason: %q", httpResponse.FailureReason)
		return
	}

	ret.Interval = httpResponse.Interval
	ret.Leechers = httpResponse.Incomplete
	ret.Seeders = httpResponse.Complete
	ret.Peers = httpResponse.Peers.List

	return
}

func setAnnounceParams(_url *url.URL, req *trackerTypes.AnnounceRequest) {
	q := url.Values{}

	q.Set("key", strconv.FormatInt(int64(req.Key), 10))
	q.Set("info_hash", string(req.InfoHash[:]))
	q.Set("peer_id", string(req.PeerID[:]))
	q.Set("port", fmt.Sprintf("%d", req.Port))
	q.Set("uploaded", strconv.FormatInt(req.Uploaded, 10))
	q.Set("downloaded", strconv.FormatInt(req.Downloaded, 10))

	left := req.Left

	if left < 0 {
		left = math.MaxInt64
	}
	q.Set("left", strconv.FormatInt(left, 10))

	if req.Event != trackerTypes.AnnounceEventEmpty {
		q.Set("event", string(req.Event))
	}

	q.Set("compact", "1")
	q.Set("supportcrypto", "1")
	q.Set("ip", req.IP.String())

	qstr := strings.ReplaceAll(q.Encode(), "+", "%20")

	if _url.RawQuery != "" {
		_url.RawQuery += "&" + qstr
	} else {
		_url.RawQuery = qstr
	}
}
