package torrent

import (
	"fmt"
	"sync"

	"github.com/al002/zbittorrent/internal/logger"
	"github.com/al002/zbittorrent/pkg/metainfo"
)

type HubConfig struct {
	BasePath            string
	GlobalUploadLimit   int
	GlobalDownloadLimit int
}

type Hub struct {
	logger   logger.Logger
	torrents map[metainfo.Hash]*Torrent
	config   HubConfig
	mu       sync.RWMutex
}

func (h *Hub) NewHub(cfg HubConfig, log logger.Logger) *Hub {
	return &Hub{
		torrents: make(map[metainfo.Hash]*Torrent),
		config:   cfg,
		logger:   log,
	}
}

func (h *Hub) Start() {
	h.logger.Info("Starting torrent hub")
}

func (h *Hub) Stop() {
	h.logger.Info("Stopping torrent hub")
}

func (h *Hub) AddTorrentFromFile(filename string) (*Torrent, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	torrent, err := NewTorrentFromFile(filename)

	if err != nil {
		return nil, err
	}

	if _, exists := h.torrents[torrent.InfoHash]; exists {
		return nil, fmt.Errorf("torrent already exists")
	}

	h.torrents[torrent.InfoHash] = torrent

	return torrent, nil
}

func (h *Hub) RemoveTorrent(infoHash metainfo.Hash, deleteFile bool) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, exists := h.torrents[infoHash]

	if !exists {
		return fmt.Errorf("torrent with info hash %s not found", infoHash.HexString())
	}

	if deleteFile {

	}

	delete(h.torrents, infoHash)

	return nil
}

func (h *Hub) GetTorrent(infoHash metainfo.Hash) (*Torrent, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	torrent, exists := h.torrents[infoHash]
	return torrent, exists
}
