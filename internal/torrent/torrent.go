package torrent

import (
	"fmt"
	"time"

	"github.com/al002/zbittorrent/pkg/metainfo"
	"github.com/al002/zbittorrent/pkg/types"
)

type State int

const (
	StateChecking State = iota
	StateDownloading
	StateSeeding
	StatePaused
	StateStopped
	StateError
)

type Stats struct {
	DownloadedBytes int64
	UploadedBytes   int64
	DownloadSpeed   float64
	UploadSpeed     float64
	Progress        float64
	Peers           int
	Seeds           int
	ETA             time.Duration
}

type Torrent struct {
  Trackers [][]string // The tiered tracker URIs
	Name        string
	InfoHash    metainfo.Hash
	MetaInfo    *metainfo.MetaInfo
	Info        *metainfo.Info
	AddedAt     time.Time
	CompletedAt *time.Time
	State       State

	peerId types.PeerID
	stats  Stats
}

func NewTorrentFromFile(filename string) (*Torrent, error) {
	mi, err := metainfo.LoadFromFile(filename)

	if err != nil {
		return nil, fmt.Errorf("failed to load torrent file: %w", err)
	}

	info, err := mi.UnmarshalInfo()

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal info: %w", err)
	}

	return &Torrent{
    Trackers: mi.ConvertToAnnounceList(),
		Name:     info.FinalName(),
		InfoHash: mi.HashInfoBytes(),
		MetaInfo: mi,
		Info:     &info,
		AddedAt:  time.Now(),
		State:    StateStopped,
	}, nil
}

func (t *Torrent) Start() error {
	return nil
}

func (t *Torrent) Stop() error {
	return nil
}

func (t *Torrent) Pause() error {
	return nil
}

func (t *Torrent) Resume() error {
	return nil
}

func (t *Torrent) GetState() State {
	return t.State
}
