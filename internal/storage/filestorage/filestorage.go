package filestorage

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/al002/zbittorrent/internal/storage"
)

type FileStorage struct {
	dest string
	perm fs.FileMode
}

func New(dest string, perm fs.FileMode) (*FileStorage, error) {
	var err error
	dest, err = filepath.Abs(dest)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		dest: dest,
		perm: perm,
	}, nil
}

var _ storage.Storage = (*FileStorage)(nil)

func (s *FileStorage) Open(name string, size int64) (f storage.File, exists bool, err error) {
	name = filepath.Clean(name)

	name = filepath.Join(s.dest, name)

	err = os.MkdirAll(filepath.Dir(name), os.ModeDir|s.perm)
	if err != nil {
		return
	}

	var of *os.File
	// Make sure OS file closed
	defer func() {
		if err == nil && of != nil {
			err = disableReadAhead(of)
		}
		if err != nil && of != nil {
			_ = of.Close()
		} else {
			f = of
		}
	}()

	var mode = s.perm &^ 0111
	openFlags := os.O_RDWR | os.O_SYNC
	openFlags = applyNoAtimeFlag(openFlags)
	of, err = os.OpenFile(name, openFlags, mode)

	if os.IsNotExist(err) {
		openFlags |= os.O_CREATE
		of, err = os.OpenFile(name, openFlags, mode)
		if err != nil {
			return
		}
		err = of.Truncate(size)
		return
	}

	if err != nil {
		return
	}

	exists = true
	fi, err := of.Stat()
	if err != nil {
		return
	}

	if fi.Size() != size {
		err = of.Truncate(size)
	}

	return
}

func (s *FileStorage) RootDir() string {
	return s.dest
}
