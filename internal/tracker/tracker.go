// From https://github.com/cenkalti/rain
package tracker

import (
	"context"
	"errors"
	"net"
	"time"
)

type Tracker interface {
	Announce(ctx context.Context, req AnnounceRequest) (*AnnounceResponse, error)
	URL() string
}

type AnnounceRequest struct {
	Torrent Torrent
	Event   Event
	NumWant int
}

type AnnounceResponse struct {
	Interval       time.Duration
	MinInterval    time.Duration
	Leechers       int32
	Seeders        int32
	WarningMessage string
	Peers          []*net.TCPAddr
}

var ErrDecode = errors.New("cannot decode response")

type Error struct {
	FailureReason string
	RetryIn       time.Duration
}

func (e *Error) Error() string {
	return e.FailureReason
}
