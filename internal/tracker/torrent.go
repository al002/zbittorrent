package tracker

// Torrent related info to sent in announce request
type Torrent struct {
	BytesUploaded   int64
	BytesDownlowded int64
	BytesLeft       int64
	InfoHash        [20]byte
	PeerID          [20]byte
	Port            int
}
