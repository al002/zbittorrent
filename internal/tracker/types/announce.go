package trackerTypes

import (
	"fmt"
)

type AnnounceEvent int32

func (me *AnnounceEvent) UnmarshalText(text []byte) error {
	for key, str := range announceEventStrings {
		if string(text) == str {
			*me = AnnounceEvent(key)
			return nil
		}
	}
	return fmt.Errorf("unknown event")
}

var announceEventStrings = []string{"", "completed", "started", "stopped"}

func (e AnnounceEvent) String() string {
	// See BEP 3, "event", and
	// https://github.com/anacrolix/torrent/issues/416#issuecomment-751427001. Return a safe default
	// in case event values are not sanitized.
	if e < 0 || int(e) >= len(announceEventStrings) {
		return ""
	}
	return announceEventStrings[e]
}

const (
	AnnounceEventEmpty AnnounceEvent = iota
	AnnounceEventCompleted
	AnnounceEventStarted
	AnnounceEventStopped
)

type AnnounceRequest struct {
	InfoHash   [20]byte
	PeerID     [20]byte
	Downloaded int64
	Left       int64
	Uploaded   int64
	Event      AnnounceEvent
	IPAddress  uint32
	Key        uint32
	NumWant    int32
	Port       uint16
} // 82 bytes
