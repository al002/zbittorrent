package trackerTypes

import (
	"net"

	"github.com/al002/zbittorrent/pkg/metainfo"
	"github.com/al002/zbittorrent/pkg/types"
)

type AnnounceEvent string

const (
	AnnounceEventStarted   AnnounceEvent = "started"
	AnnounceEventCompleted AnnounceEvent = "completed"
	AnnounceEventStopped   AnnounceEvent = "stopped"
	AnnounceEventEmpty     AnnounceEvent = ""
)

type AnnounceRequest struct {
	InfoHash   metainfo.Hash
	PeerID     types.PeerID
	IP         net.IP
	Port       uint16
	Uploaded   int64
	Downloaded int64
	Left       int64
	Event      AnnounceEvent
	Key        uint32
	NumWant    int32
}
