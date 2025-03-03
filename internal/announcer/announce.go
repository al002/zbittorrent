package announcer

import (
	"context"
	"errors"

	"github.com/al002/zbittorrent/internal/tracker"
)

func announce(
  ctx context.Context,
  t tracker.Tracker,
  e tracker.Event,
  numWant int,
  torrent tracker.Torrent,
  responseC chan *tracker.AnnounceResponse,
  errC chan error,
) {
  req := tracker.AnnounceRequest{
    Torrent: torrent,
    Event: e,
    NumWant: numWant,
  }
  resp, err := t.Announce(ctx, req)
  if errors.Is(err, context.Canceled) {
    return
  }

  if err != nil {
    select {
    case errC <- err:
    case <-ctx.Done():
    }
  }

  select {
  case responseC <- resp:
  case <-ctx.Done():
  }
}
