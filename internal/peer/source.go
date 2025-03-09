package peer

type Source int

const (
	// The peer is found from tracker by announcing the torrent
	Tracker Source = iota
	// The peer is found from DHT node
	DHT
	// The peer is found from another peer with PEX messages
	PEX
	// The peer is added manually by user
	Manual
	// The peer found us
	Incoming
)

func (s Source) String() string {
	switch s {
	case Tracker:
		return "tracker"
	case DHT:
		return "dht"
	case PEX:
		return "pex"
	case Manual:
		return "manual"
	case Incoming:
		return "incoming"
	default:
		panic("unhandled source")
	}
}
