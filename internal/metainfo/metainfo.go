package metainfo

import (
	"errors"
	"io"
	"strings"

	"github.com/al002/zbittorrent/pkg/bencode"
)

type MetaInfo struct {
	Info         Info
	AnnounceList [][]string
	URLList      []string
}

func New(r io.Reader) (*MetaInfo, error) {
	var ret MetaInfo
	var t struct {
		Info         bencode.Bytes `bencode:"info"`
		Announce     bencode.Bytes `bencode:"announce"`
		AnnounceList bencode.Bytes `bencode:"announce-list"`
		URLList      bencode.Bytes `bencode:"url-list"`
	}

	err := bencode.NewDecoder(r).Decode(&t)
	if err != nil {
		return nil, err
	}

	if len(t.Info) == 0 {
		return nil, errors.New("no info dict in torrent file")
	}

	info, err := NewInfo(t.Info, true, true)
	if err != nil {
		return nil, err
	}

	ret.Info = *info

	if len(t.AnnounceList) > 0 {
    var ll [][]string
    err = bencode.Unmarshal(t.AnnounceList, &ll)
    if err == nil {
      for _, tier := range ll {
        var ti []string
        for _, t := range tier {
          if isTrackerSupported(t) {
            ti = append(ti, t)
          }
        }
        if len(ti) > 0 {
          ret.AnnounceList = append(ret.AnnounceList, ti)
        }
      }
    }
	} else {
		var s string
		err = bencode.Unmarshal(t.Announce, &s)
		if err == nil && isTrackerSupported(s) {
			ret.AnnounceList = append(ret.AnnounceList, []string{s})
		}
	}

	if len(t.URLList) > 0 {
		if t.URLList[0] == 'l' {
			var l []string
			err = bencode.Unmarshal(t.URLList, &l)
			if err == nil {
				for _, s := range l {
					if isWebseedSupported(s) {
						ret.URLList = append(ret.URLList, s)
					}
				}
			}
		} else {
			var s string
			err = bencode.Unmarshal(t.URLList, &s)
			if err != nil && isWebseedSupported(s) {
				ret.URLList = append(ret.URLList, s)
			}
		}
	}

	return &ret, nil
}

func isTrackerSupported(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "udp://")
}

func isWebseedSupported(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
