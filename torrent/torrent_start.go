package torrent

import (
	"net"

	"github.com/al002/zbittorrent/internal/acceptor"
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

  t.startAcceptor()
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

func (t *torrent) startAcceptor() {
	if t.acceptor != nil {
		return
	}

	ip := net.ParseIP(t.session.config.Host)
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{
		IP:   ip,
		Port: t.port,
	})

	if err != nil {
		t.log.Warn(
			"cannot listen port",
			"port", t.port,
			"err", err.Error(),
		)
	} else {
		t.log.Info(
			"Listening peers on tcp://"+listener.Addr().String(),
			"addr", listener.Addr().String(),
		)
    t.port = listener.Addr().(*net.TCPAddr).Port
    t.acceptor = acceptor.New(listener, t.incomingConnC, t.log)
    go t.acceptor.Run()
	}
}
