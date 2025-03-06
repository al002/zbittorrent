package torrent

import (
	"github.com/al002/zbittorrent/internal/announcer"
	"github.com/al002/zbittorrent/internal/tracker"
)

func (t *torrent) start() {
	if t.errC != nil {
		return
	}

	if t.stoppedEventAnnouncer != nil {
		t.stoppedEventAnnouncer.Close()
		t.stoppedEventAnnouncer = nil
	}

	t.errC = make(chan error, 1)
	t.lastError = nil

	t.startAnnouncers()
	// if t.info != nil {
	//
	// } else {
	//   t.startAnnouncers()
	// }
}

func (t *torrent) startAnnouncers() {
	if len(t.announcers) == 0 {
		for _, tr := range t.trackers {
			t.startNewAnnouncer(tr)
		}
	}
}

func (t *torrent) startNewAnnouncer(tr tracker.Tracker) {
	a := announcer.NewPeriodicalAnnouncer(
		tr,
		t.session.config.TrackerNumWant,
		t.session.config.TrackerMinAnnounceInterval,
		t.announceGetTorrent,
		t.completeC,
		t.announcePeersC,
	)

	t.announcers = append(t.announcers, a)

	go a.Run()
}
