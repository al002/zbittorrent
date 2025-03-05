package torrent

import "time"

type TrackerStatus int

const (
	// NotContactedYet indicates that no announce request has been made to the tracker.
	NotContactedYet TrackerStatus = iota
	// Contacting the tracker. Sending request or waiting response from the tracker.
	Contacting
	// Working indicates that the tracker has responded as expected.
	Working
	// NotWorking indicates that the tracker didn't respond or returned an error.
	NotWorking
)

type Tracker struct {
	URL      string
	Status   TrackerStatus
	Leechers int
	Seeders  int
	// Error        *AnnouceError
	Warning      string
	LastAnnounce time.Time
	NextAnnounce time.Time
}

type trackersRequest struct {
	Response chan []Tracker
}

func (t *torrent) Start() {
	select {
	case t.startCommandC <- struct{}{}:
	case <-t.closeC:
	}
}
