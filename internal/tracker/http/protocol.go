package httpTracker

import (
	"fmt"

	"github.com/al002/zbittorrent/pkg/bencode"
	"github.com/anacrolix/dht/v2/krpc"
)

type HttpTrackerResponse struct {
	FailureReason string `bencode:"failure reason"`
	Interval      int32  `bencode:"interval"`
	TrackerId     string `bencode:"tracker id"`
	Complete      int32  `bencode:"complete"`
	Incomplete    int32  `bencode:"incomplete"`
	Peers         Peers  `bencode:"peers"`
}

type Peers struct {
	List    []Peer
	Compact bool
}

func (p *Peers) UnmarshalBencode(b []byte) (err error) {
	var _v interface{}
	err = bencode.Unmarshal(b, &_v)

	if err != nil {
		return
	}

	switch v := _v.(type) {
  // BEP 23 compact
	case string:
		var cnas krpc.CompactIPv4NodeAddrs
		err = cnas.UnmarshalBinary([]byte(v))
		if err != nil {
			return
		}

		p.Compact = true
		for _, cp := range cnas {
			p.List = append(p.List, Peer{
				IP:   cp.IP[:],
				Port: cp.Port,
			})
		}
		return
  // BEP 3 non-compact
	case []interface{}:
		p.Compact = false
		for _, i := range v {
			var pp Peer
			pp.FromDictInterface(i.(map[string]interface{}))
			p.List = append(p.List, pp)
		}
		return
	default:
		err = fmt.Errorf("unsupported type: %T", _v)
		return
	}
}

func (p *Peers) MarshalBencode() ([]byte, error) {
	if p.Compact {
		cnas := make([]krpc.NodeAddr, 0, len(p.List))
		for _, peer := range p.List {
			cnas = append(cnas, krpc.NodeAddr{
				IP:   peer.IP,
				Port: peer.Port,
			})
		}
		return krpc.CompactIPv4NodeAddrs(cnas).MarshalBencode()
	} else {
		return bencode.Marshal(p.List)
	}
}
