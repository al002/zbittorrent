// Interface for reading and writing files in a torrent.
package storage

import "io"

type Storage interface {
	Open(name string, size int64) (f File, exists bool, err error)
	RootDir() string
}

type File interface {
	io.ReaderAt
	io.WriterAt
	io.Closer
}

type Provider interface {
	GetStorage(torrentID string) (Storage, error)
}
