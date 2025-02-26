package metainfo

type FileInfo struct {
	Length int64    `bencode:"length"`
	Path   []string `bencode:"path"` // BEP 3
}

type Info struct {
	PieceLength int64      `bencode:"piece length"` // BEP 3
	Pieces      []byte     `bencode:"pieces,omitempty"`
	Name        string     `bencode:"name"`
	NameUtf8    string     `bencode:"name.utf-8,omitempty"`
	Length      int64      `bencode:"length,omitempty"`
	Private     *bool      `bencode:"private,omitempty"` // BEP27
	Files       []FileInfo `bencode:"files,omitempty"`
}

func (info *Info) FinalName() string {
	if info.NameUtf8 != "" {
		return info.NameUtf8
	}
	return info.Name
}

func (info *Info) isDir() bool {
	return info.Length == 0
}

func (info *Info) NumOfPieces() int {
	return len(info.Pieces) / 20
}
