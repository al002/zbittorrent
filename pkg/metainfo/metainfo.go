// from https://github.dev/anacrolix/torrent/blob/master/metainfo/metainfo.go ，logic simplified
package metainfo

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/al002/zbittorrent/pkg/bencode"
)

type AnnounceList [][]string

type Node string

type UrlList []string

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
	if shouldOverrideAnnounce(mi.Announce, mi.AnnounceList) {
		return mi.AnnounceList
	}

	if mi.Announce != "" {
		return [][]string{{mi.Announce}}
	}

	return nil
}

func (mi *MetaInfo) DistinctAnnounceList() (ret []string) {
	exists := make(map[string]struct{})

	for _, tier := range mi.AnnounceList {
		for _, v := range tier {
			if _, ok := exists[v]; !ok {
				exists[v] = struct{}{}
				ret = append(ret, v)
			}
		}
	}

  return
}

func shouldOverrideAnnounce(announce string, list AnnounceList) bool {
	for _, tier := range list {
		for _, url := range tier {
			if url != "" || announce == "" {
				return true
			}
		}
	}
	return false
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
