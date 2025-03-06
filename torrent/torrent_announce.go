package torrent

import "github.com/al002/zbittorrent/internal/tracker"

func (t *torrent) announceGetTorrent() tracker.Torrent {
  tr := tracker.Torrent{
    InfoHash: t.infoHash,
    PeerID: t.peerID,
    Port: t.port,
    BytesDownlowded: 0,
    BytesUploaded: 0,
    // BytesDownloaded: t.bytesDownloaded.Count(),
    // BytesUploaded: t.bytesUploaded.Count(),
  }

  return tr
}
