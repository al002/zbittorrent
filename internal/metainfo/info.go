package metainfo

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/al002/zbittorrent/pkg/bencode"
)

var (
	errInvalidPieceData = errors.New("invalid piece data")
	errZeroPieceLength  = errors.New("torrent has zero piece length")
	errZeroPieces       = errors.New("torrent has zero pieces")
	errPieceLength      = errors.New("piece length must be multiple of 16K")
)

type Info struct {
	PieceLength uint32
	Name        string
	Hash        [20]byte
	Length      int64
	NumPieces   uint32
	Private     bool
	Files       []File
	Raw         []byte
	pieces      []byte
}

type File struct {
	Length int64
	Path   string
	// https://www.bittorrent.org/beps/bep_0047.html
	Padding bool
}

type file struct {
	Length   int64    `bencode:"length"`
	Path     []string `bencode:"path"`
	PathUTF8 []string `bencode:"path.utf-8,omitempty"`
	Attr     string   `bencode:"attr"`
}

func (f *file) isPadding() bool {
	// BEP 0047
	if strings.ContainsRune(f.Attr, 'p') {
		return true
	}

	// BitComet convention that do not conform BEP 0047
	if len(f.Path) > 0 && strings.HasPrefix(f.Path[len(f.Path)-1], "_____padding_file") {
		return true
	}

	return false
}

func (i *infoType) overrideUTF8Keys() {
	if len(i.NameUTF8) > 0 {
		i.Name = i.NameUTF8
	}

	for j := range i.Files {
		if len(i.Files[j].PathUTF8) > 0 {
			i.Files[j].Path = i.Files[j].PathUTF8
		}
	}
}

type infoType struct {
	PieceLength uint32 `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces,omitempty"`
	Name        string `bencode:"name"`
	NameUTF8    string `bencode:"name.utf-8,omitempty"`
	Private     bool  `bencode:"private,omitempty"`
	Length      int64  `bencode:"length,omitempty"` // Single File Mode
	Files       []file `bencode:"files,omitempty"`  // Multiple File mode
}

func NewInfo(b []byte, utf8 bool, pad bool) (*Info, error) {
	var it infoType
	if err := bencode.Unmarshal(b, &it); err != nil {
		return nil, err
	}

	if it.PieceLength == 0 {
		return nil, errZeroPieceLength
	}

	if len(it.Pieces)%sha1.Size != 0 {
		return nil, errInvalidPieceData
	}

	numPieces := len(it.Pieces) / sha1.Size

	if numPieces == 0 {
		return nil, errZeroPieces
	}

	if utf8 {
		it.overrideUTF8Keys()
	}

	for _, file := range it.Files {
		for _, path := range file.Path {
			if strings.TrimSpace(path) == ".." {
				return nil, fmt.Errorf("invalid filename: %q", filepath.Join(file.Path...))
			}
		}
	}

	i := Info{
		PieceLength: it.PieceLength,
		NumPieces:   uint32(numPieces),
		Name:        it.Name,
		Private:     it.Private,
		pieces:      it.Pieces,
	}

	i.setLength(it)

	err := i.checkPieceDataLength()
	if err != nil {
		return nil, err
	}

	i.Raw = b

	// info hash
	i.setHash(b)

	// fill name filed
	i.setName(it)

	// set files
  err = i.setFiles(it, pad)
  if err != nil{
    return nil, err
  }

	return &i, nil
}

func (i *Info) setLength(it infoType) {
	multiFile := len(it.Files) > 0
	if multiFile {
		for _, f := range it.Files {
			i.Length += f.Length
		}
	} else {
		i.Length = it.Length
	}
}

func (i *Info) checkPieceDataLength() error {
	totalPieceDataLength := int64(i.PieceLength) * int64(i.NumPieces)
	delta := totalPieceDataLength - i.Length

	if delta >= int64(i.PieceLength) || delta < 0 {
		return errInvalidPieceData
	}

	return nil
}

func (i *Info) setHash(b []byte) {
	hash := sha1.New()
	_, _ = hash.Write(b)
	copy(i.Hash[:], hash.Sum(nil))
}

func (i *Info) setName(it infoType) {
	if it.Name != "" {
		i.Name = it.Name
	} else {
		i.Name = hex.EncodeToString(i.Hash[:])
	}
}

func (i *Info) setFiles(it infoType, pad bool) error {
	multiFile := len(it.Files) > 0
	if multiFile {
		i.Files = make([]File, len(it.Files))
		uniquePaths := make(map[string]interface{}, len(it.Files))

		for j, f := range it.Files {
			parts := make([]string, 0, len(f.Path)+1)
      // torrent name as base path
      parts = append(parts, truncateName(i.Name))

      for _, p := range f.Path {
        // append filename to parts
        parts = append(parts, truncateName(p))
      }

      // get filename path
      joinedPath := filepath.Join(parts...)

      if _, ok := uniquePaths[joinedPath]; ok {
        return fmt.Errorf("duplicate filename: %q", joinedPath)
      } else {
        uniquePaths[joinedPath] = nil
      }

      i.Files[j] = File{
        Path: joinedPath,
        Length: f.Length,
      }

      if pad {
        i.Files[j].Padding = f.isPadding()
      }
		}
	} else {
		i.Files = []File{{Path: truncateName(i.Name), Length: i.Length}}
	}

  return nil
}

func truncateName(s string) string {
	return truncateNameN(s, 255)
}

func truncateNameN(s string, max int) string {
	s = strings.ToValidUTF8(s, string(unicode.ReplacementChar))
	s = trimName(s, max)
	s = strings.ToValidUTF8(s, "")

	return replaceSeparator(s)
}

func trimName(s string, max int) string {
	if len(s) <= max {
		return s
	}

	ext := path.Ext(s)
	if len(ext) > max {
		return s[:max]
	}

	return s[:max-len(ext)] + ext
}

func replaceSeparator(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '/' {
			return '_'
		}
		return r
	}, s)
}
