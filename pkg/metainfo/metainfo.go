// from https://github.dev/anacrolix/torrent/blob/master/metainfo/metainfo.go ，logic simplified
package metainfo

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/al002/zbittorrent/pkg/bencode"
)

type MetaInfo struct {
	Announce     string        `bencode:"announce,omitempty"`
	InfoBytes    bencode.Bytes `bencode:"info,omitempty"`                              // BEP 3
	AnnounceList AnnounceList  `bencode:"announce-list,omitempty"`                     // BEP 12
	Nodes        []Node        `bencode:"nodes,omitempty,ignore_unmarshal_type_error"` // BEP 5
	CreationDate int64         `bencode:"creation date,omitempty,ignore_unmarshal_type_error"`
	Comment      string        `bencode:"comment,omitempty"`
	CreatedBy    string        `bencode:"created by,omitempty"`
	Encoding     string        `bencode:"encoding,omitempty"`
	UrlList      UrlList       `bencode:"url-list,omitempty"` // BEP 19
}

func (mi *MetaInfo) HashInfo() (Hash, error) {
	b, err := bencode.Marshal(mi.InfoBytes)
	if err != nil {
		return Hash{}, err
	}

	return HashBytes(b), nil
}

func (mi *MetaInfo) UnmarshalInfo() (info Info, err error) {
	err = bencode.Unmarshal(mi.InfoBytes, &info)
	return
}

func (mi *MetaInfo) HashInfoBytes() (infoHash Hash) {
	return HashBytes(mi.InfoBytes)
}

func (mi *MetaInfo) Marshal(w io.Writer) error {
	return bencode.NewEncoder(w).Encode(mi)
}

func (mi *MetaInfo) ConvertToAnnounceList() AnnounceList {
	if mi.AnnounceList.shouldOverrideAnnounce(mi.Announce) {
		return mi.AnnounceList
	}

	if mi.Announce != "" {
		return [][]string{{mi.Announce}}
	}

	return nil
}

func Load(r io.Reader) (*MetaInfo, error) {
	var mi MetaInfo
	d := bencode.NewDecoder(r)
	err := d.Decode(&mi)
	if err != nil {
		return nil, err
	}

	err = d.ReadEOF()
	if err != nil {
		err = fmt.Errorf("error after decoding metainfo: %w", err)
	}

	return &mi, err
}

func LoadFromFile(filename string) (*MetaInfo, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	var buf bufio.Reader
	buf.Reset(f)
	return Load(&buf)
}

type AnnounceList [][]string

func (al AnnounceList) DistinctAnnounceList() (ret []string) {
	exists := make(map[string]struct{})

	for _, tier := range al {
		for _, v := range tier {
			if _, ok := exists[v]; !ok {
				exists[v] = struct{}{}
				ret = append(ret, v)
			}
		}
	}

	return
}

func (al AnnounceList) shouldOverrideAnnounce(announce string) bool {
	for _, tier := range al {
		for _, url := range tier {
			if url != "" || announce == "" {
				return true
			}
		}
	}
	return false
}

type Node string

var _ bencode.Unmarshaler = (*Node)(nil)

func (n *Node) UnmarshalBencode(b []byte) (err error) {
	var iface interface{}
	err = bencode.Unmarshal(b, &iface)

	if err != nil {
		return
	}

	switch v := iface.(type) {
	case string:
		*n = Node(v)
	case []interface{}:
		func() {
			defer func() {
				r := recover()
				if r != nil {
					err = r.(error)
				}
			}()
			*n = Node(
				net.JoinHostPort(v[0].(string), strconv.FormatInt(v[1].(int64), 10)),
			)
		}()
	default:
		err = fmt.Errorf("unsupported type: %T", iface)
	}

	return
}

type UrlList []string

var _ bencode.Unmarshaler = (*UrlList)(nil)

func (me *UrlList) UnmarshalBencode(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	if b[0] == 'l' {
		var l []string
		err := bencode.Unmarshal(b, &l)
		*me = l
		return err
	}

	var s string
	err := bencode.Unmarshal(b, &s)
	*me = []string{s}
	return err
}
