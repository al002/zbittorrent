package torrent

import (
	"io/fs"
	"time"
)

var (
	publicPeerIDPrefix         = "-ZB" + Version + "-"
	trackerHTTPPublicUserAgent = "ZB/" + Version
)

type Config struct {
	// Database file to save resume data.
	Database string `mapstructure:"database"`
	// DataDir is where files are downloaded.
	DataDir string `mapstructure:"data_dir"`
	// Host to listen for TCP Acceptor. Port is computed automatically
	Host         string `mapstructure:"host"`
	PortBegin    uint16 `mapstructure:"port_begin"`
	PortEnd      uint16 `mapstructure:"port_end"`
	MaxOpenFiles uint64 `mapstructure:"max_open_files"`
	// Enable peer exchange protocol.
	PEXEnabled bool `mapstructure:"pex_enabled"`
	// Resume data (bitfield & stats) are saved to disk at interval to keep IO lower.
	ResumeWriteInterval time.Duration `mapstructure:"resume_write_interval"`
	// Peer id is prefixed with this string. See BEP 20. Remaining bytes of peer id will be randomized.
	PrivatePeerIDPrefix                    string `mapstructure:"private_peer_id_prefix"`
	PrivateExtensionHandshakeClientVersion string `mapstructure:"private_extension_handshake_client_version"`
	// URL to the blocklist file in CIDR format.
	BlocklistURL            string        `mapstructure:"blocklist_url"`
	BlocklistUpdateInterval time.Duration `mapstructure:"blocklist_update_interval"`
	BlocklistUpdateTimeout  time.Duration `mapstructure:"blocklist_update_timeout"`
	// Do not contact tracker if it's IP is blocked
	BlocklistEnabledForTrackers bool `mapstructure:"blocklist-enabled-for-trackers"`
	// Do not connect to peer if it's IP is blocked
	BlocklistEnabledForOutgoingConnections bool `mapstructure:"blocklist_enabled_for_outgoing_connections"`
	// Do not accept connections from peer if it's IP is blocked
	BlocklistEnabledForIncomingConnections bool `mapstructure:"blocklist_enabled_for_incoming_connections"`
	// Do not accept response larger than this size
	BlocklistMaxResponseSize int64 `mapstructure:"blocklist_max_response_size"`
	//  // Time to wait when adding torrent with AddURI().
	//  TorrentAddHTTPTimeout time.Duration `mapstructure:"torrent_add_http_timeout"`
	// // Maximum allowed size to be received by metadata extension.
	// MaxMetadataSize uint `mapstructure:"max_metadata_size"`
	// Maximum allowed size to be read when adding torrent.
	MaxTorrentSize uint `mapstructure:"max_torrent_size"`
	// Time to wait when resolving host names for trackers and peers.
	DNSResolveTimeout time.Duration `mapstructure:"dns_resolve_timeout"`
	// Global download speed limit in KB/s.
	SpeedLimitDownload int64 `mapstructure:"speed_limit_download"`
	// Global upload speed limit in KB/s.
	SpeedLimitUpload int64 `mapstructure:"speed_limit_upload"`
	// Start torrent automatically if it was running when previous session was closed.
	ResumeOnStartup bool `mapstructure:"resume_on_startup"`
	// Check each torrent loop for aliveness. Helps to detect bugs earlier.
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
	// If torrent loop is stuck for more than this duration. Program crashes with stacktrace.
	HealthCheckTimeout time.Duration `mapstructure:"health_check_timeout"`
	// The unix permission of created files, execute bit is removed for files.
	// Effective only when default storage provider is used.
	FilePermissions fs.FileMode `mapstructure:"file_permissions"`

	// Number of peer addresses to request in announce request.
	TrackerNumWant int `mapstructure:"tracker_num_want"`
	// Time to wait for announcing stopped event.
	// Stopped event is sent to the tracker when torrent is stopped.
	TrackerStopTimeout time.Duration `mapstructure:"tracker_stop_timeout"`
	// When the client needs new peer addresses to connect, it ask to the tracker.
	// To prevent spamming the tracker an interval is set to wait before the next announce.
	TrackerMinAnnounceInterval time.Duration `mapstructure:"tracker_min_announce_interval"`
	// Total time to wait for response to be read.
	// This includes ConnectTimeout and TLSHandshakeTimeout.
	TrackerHTTPTimeout time.Duration `mapstructure:"tracker_http_timeout"`
	// User agent sent when communicating with HTTP trackers.
	// Only applies to private torrents.
	TrackerHTTPPrivateUserAgent string `mapstructure:"tracker_http_private_user_agent"`
	// Max number of bytes in a tracker response.
	TrackerHTTPMaxResponseSize uint `mapstructure:"tracker_http_max_response_size"`
	// Check and validate TLS ceritificates.
	TrackerHTTPVerifyTLS bool `mapstructure:"tracker_http_verify_tls"`
}

var DefaultConfig = Config{
	// Session
	Database:                               "~/zbittorrent/session.db",
	DataDir:                                "~/zbittorrent/data",
	Host:                                   "0.0.0.0",
	PortBegin:                              20000,
	PortEnd:                                30000,
	MaxOpenFiles:                           10240,
	PEXEnabled:                             true,
	ResumeWriteInterval:                    30 * time.Second,
	PrivatePeerIDPrefix:                    "-ZB" + Version + "-",
	PrivateExtensionHandshakeClientVersion: "zbittorrent " + Version,
	BlocklistUpdateInterval:                24 * time.Hour,
	BlocklistUpdateTimeout:                 10 * time.Minute,
	BlocklistEnabledForTrackers:            true,
	BlocklistEnabledForOutgoingConnections: true,
	BlocklistEnabledForIncomingConnections: true,
	BlocklistMaxResponseSize:               100 << 20,
	// TorrentAddHTTPTimeout:                  30 * time.Second,
	// MaxMetadataSize:                        30 << 20,
	MaxTorrentSize: 10 << 20,
	// MaxPieces:                              64 << 10,
	DNSResolveTimeout:   5 * time.Second,
	ResumeOnStartup:     true,
	HealthCheckInterval: 10 * time.Second,
	HealthCheckTimeout:  60 * time.Second,
	FilePermissions:     0o750,

	// RPC Server
	// RPCEnabled:         true,
	// RPCHost:            "127.0.0.1",
	// RPCPort:            7246,
	// RPCShutdownTimeout: 5 * time.Second,

	// Tracker
	TrackerNumWant:              200,
	TrackerStopTimeout:          5 * time.Second,
	TrackerMinAnnounceInterval:  time.Minute,
	TrackerHTTPTimeout:          10 * time.Second,
	TrackerHTTPPrivateUserAgent: "ZB/" + Version,
	TrackerHTTPMaxResponseSize:  2 << 20,
	TrackerHTTPVerifyTLS:        true,

	// DHT node
	// DHTEnabled:             true,
	// DHTHost:                "0.0.0.0",
	// DHTPort:                7246,
	// DHTAnnounceInterval:    30 * time.Minute,
	// DHTMinAnnounceInterval: time.Minute,
	// DHTBootstrapNodes: []string{
	// 	"router.bittorrent.com:6881",
	// 	"dht.transmissionbt.com:6881",
	// 	"router.utorrent.com:6881",
	// 	"dht.libtorrent.org:25401",
	// 	"dht.aelitis.com:6881",
	// },

	// Peer
	// UnchokedPeers:                3,
	// OptimisticUnchokedPeers:      1,
	// MaxRequestsIn:                250,
	// MaxRequestsOut:               250,
	// DefaultRequestsOut:           50,
	// RequestTimeout:               20 * time.Second,
	// EndgameMaxDuplicateDownloads: 20,
	// MaxPeerDial:                  80,
	// MaxPeerAccept:                20,
	// ParallelMetadataDownloads:    2,
	// PeerConnectTimeout:           5 * time.Second,
	// PeerHandshakeTimeout:         10 * time.Second,
	// PieceReadTimeout:             30 * time.Second,
	// MaxPeerAddresses:             2000,
	// AllowedFastSet:               10,

	// IO
	// ReadCacheBlockSize: 128 << 10,
	// ReadCacheSize:      256 << 20,
	// ReadCacheTTL:       1 * time.Minute,
	// ParallelReads:      1,
	// ParallelWrites:     1,
	// WriteCacheSize:     1 << 30,

	// Webseed settings
	// WebseedDialTimeout:             10 * time.Second,
	// WebseedTLSHandshakeTimeout:     10 * time.Second,
	// WebseedResponseHeaderTimeout:   10 * time.Second,
	// WebseedResponseBodyReadTimeout: 10 * time.Second,
	// WebseedRetryInterval:           time.Minute,
	// WebseedVerifyTLS:               true,
	// WebseedMaxSources:              10,
	// WebseedMaxDownloads:            4,
}
