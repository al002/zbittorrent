package torrent

import (
	"crypto/rand"
	"errors"
	"net"
	"time"

	"github.com/al002/zbittorrent/internal/announcer"
	"github.com/al002/zbittorrent/internal/metainfo"
	"github.com/al002/zbittorrent/internal/tracker"
)

type torrent struct {
	session  *Session
	id       string
	addedAt  time.Time
	infoHash [20]byte
	info     *metainfo.Info
	peerID   [20]byte
	// List of addresses to announce
	trackers   []tracker.Tracker
	rawTracker []string
	name       string
	closeC     chan struct{} // When Stop() is called, it will close this channel to singal run() function to stop
	doneC      chan struct{} // Close() blocks untile doneC is closed
	errC       chan error

	completeC chan struct{}

	port int

	// last error sent to errC
	lastError error

	// Channels for sending a message to run() loop
	trackersCommandC    chan trackersRequest   // Trackers()
	startCommandC       chan struct{}          // Start()
	stopCommandC        chan struct{}          // Stop()
	announceCommandC    chan struct{}          // Announce()
	addTrackersCommandC chan []tracker.Tracker // AddTrackers()

	// Trackers send announce responses to this channel
	announcePeersC chan []*net.TCPAddr

	// Announces the status of torrent to trackers to get peer addresses periodically.
	announcers            []*announcer.PeriodicalAnnouncer
	stoppedEventAnnouncer *announcer.StopAnnouncer

	// A signal sent to run() loop when announcers are stopped
	announcersStoppedC chan struct{}
}

func newTorrent(
	session *Session,
	id string,
	addedAt time.Time,
	infoHash []byte,
	info *metainfo.Info,
	name string,
	trackers []tracker.Tracker,
) (*torrent, error) {
	if len(infoHash) != 20 {
		return nil, errors.New("invalid infoHash (must be 20 bytes)")
	}

	var ih [20]byte
	copy(ih[:], infoHash)

	t := &torrent{
		session:             session,
		id:                  id,
		addedAt:             addedAt,
		infoHash:            ih,
		info:                info,
		trackers:            trackers,
		name:                name,
		completeC:           make(chan struct{}),
		closeC:              make(chan struct{}),
		doneC:               make(chan struct{}),
		startCommandC:       make(chan struct{}),
		stopCommandC:        make(chan struct{}),
		trackersCommandC:    make(chan trackersRequest),
		addTrackersCommandC: make(chan []tracker.Tracker),
		announceCommandC:    make(chan struct{}),
		announcersStoppedC:  make(chan struct{}),
	}

	n := t.copyPeerIDPrefix()
	_, err := rand.Read(t.peerID[n:])
	if err != nil {
		return nil, err
	}

	go t.run()
	return t, nil
}

func (t *torrent) copyPeerIDPrefix() int {
	if t.info != nil && t.info.Private {
		return copy(t.peerID[:], t.session.config.PrivatePeerIDPrefix)
	}

	return copy(t.peerID[:], publicPeerIDPrefix)
}

func (t *torrent) run() {
	for {
		select {
		case <-t.closeC:
			t.close()
			close(t.doneC)
			return
		case <-t.startCommandC:
			t.start()
			// case <-t.stopCommandC:
			// case <-t.announceCommandC:
			// case <-t.announcersStoppedC:
			// case req := <-t.trackersCommandC:
			// case trackers := <-t.addTrackersCommandC:
		}
	}
}

func (t *torrent) Name() string {
	return t.name
}

func (t *torrent) InfoHash() []byte {
	b := make([]byte, 20)
	copy(b, t.infoHash[:])

	return b
}

func (t *torrent) Files() ([]File, error) {
	if t.info == nil {
		return nil, errors.New("torrent metadata not ready")
	}

	files := make([]File, 0, len(t.info.Files))

	for _, f := range t.info.Files {
		if !f.Padding {
			files = append(files, File{
				path:   f.Path,
				length: f.Length,
			})
		}
	}

	return files, nil
}

var errClosed = errors.New("torrent is closed")

func (t *torrent) close() {
	// t.stop(errClosed)
}

type File struct {
	path   string
	length int64
}

func (f File) Path() string {
	return f.path
}

func (f File) Length() int64 {
	return f.length
}
