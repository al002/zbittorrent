package httpTracker

import "net"

type Peer struct {
	IP   net.IP `bencode:"ip"`
	Port int    `bencode:"port"`
	ID   []byte `bencode:"peer_id"`
}

// Non-compact BEP 3
func (p *Peer) FromDictInterface(d map[string]interface{}) {
  p.IP = net.ParseIP(d["ip"].(string))
	if _, ok := d["peer id"]; ok {
		p.ID = []byte(d["peer id"].(string))
	}

	p.Port = int(d["port"].(int64))
}
