package metainfo

import "github.com/al002/zbittorrent/pkg/infohash"

const HashSize = infohash.Size

type Hash = infohash.T

var (
	NewHashFromHex = infohash.FromHexString
	HashBytes      = infohash.HashBytes
)
