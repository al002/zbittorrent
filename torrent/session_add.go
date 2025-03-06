package torrent

import (
	"io"
	"time"

	"github.com/al002/zbittorrent/internal/metainfo"
	"github.com/al002/zbittorrent/internal/tracker"
)

type AddTorrentOptions struct {
	ID string
}

func (s *Session) AddTorrent(r io.Reader, opts *AddTorrentOptions) (*Torrent, error) {
	if opts == nil {
		opts = &AddTorrentOptions{}
	}

	t, err := s.addTorrent(r, opts)

	if err != nil {
		return nil, err
	}

	err = t.Start()

	return t, err
}

func (s *Session) addTorrent(r io.Reader, opts *AddTorrentOptions) (*Torrent, error) {
	r = io.LimitReader(r, int64(s.config.MaxTorrentSize))
	mi, err := s.parseMetaInfo(r)

	if err != nil {
		return nil, err
	}

	t, err := newTorrent(
		s,
		opts.ID,
		time.Now(),
		mi.Info.Hash[:],
		&mi.Info,
		mi.Info.Name,
		s.parseTrackers(mi.AnnounceList, mi.Info.Private),
	)

	if err != nil {
		return nil, err
	}

	// defer func() {
	//   if err != nil {
	//     t.Close()
	//   }
	// }()

	t2 := s.insertTorrent(t)

	return t2, nil
}

func (s *Session) parseMetaInfo(r io.Reader) (*metainfo.MetaInfo, error) {
	mi, err := metainfo.New(r)
	if err != nil {
		return nil, err
	}

	return mi, nil
}

func (s *Session) parseTrackers(tiers [][]string, private bool) []tracker.Tracker {
	ret := make([]tracker.Tracker, 0, len(tiers))
	for _, tier := range tiers {
		trackers := make([]tracker.Tracker, 0, len(tier))
		for _, tr := range tier {
			t, err := s.trackerManager.Get(tr, s.config.TrackerHTTPTimeout, s.getTrackerUserAgent(private), int64(s.config.TrackerHTTPMaxResponseSize))
			if err != nil {
				continue
			}

			trackers = append(trackers, t)
		}
		if len(trackers) > 0 {
			tra := tracker.NewTier(trackers)
			ret = append(ret, tra)
		}
	}

	return ret
}

func (s *Session) insertTorrent(t *torrent) *Torrent {
	t2 := &Torrent{
		torrent: t,
	}

	s.mTorrents.Lock()
	defer s.mTorrents.Unlock()
	s.torrents[t.id] = t2

	return t2
}
