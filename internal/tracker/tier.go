package tracker

import (
	"context"
	"math/rand/v2"
	"sync/atomic"
)

// try announce to the working Tracker
type Tier struct {
  Trackers []Tracker
  index int32
}

var _ Tracker = (*Tier)(nil)

func NewTier(trackers []Tracker) *Tier {
  rand.Shuffle(len(trackers), func(i, j int) {
    trackers[i], trackers[j] = trackers[j], trackers[i]
  })

  return &Tier{
    Trackers: trackers,
  }
}

func (t *Tier) Announce(ctx context.Context, req AnnounceRequest) (*AnnounceResponse, error) {
  index := t.loadIndex()
  resp, err := t.Trackers[index].Announce(ctx, req)
  if err != nil {
    // new index
    atomic.CompareAndSwapInt32(&t.index, index, index+1)
  }

  return resp, err
}

func (t *Tier) URL() string {
  return t.Trackers[t.loadIndex()].URL()
}

func (t *Tier) loadIndex() int32 {
  index := atomic.LoadInt32(&t.index)
  if index >= int32(len(t.Trackers)) {
    index = 0
  }

  return index
}
