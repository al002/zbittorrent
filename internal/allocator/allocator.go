package allocator

import (
	"github.com/al002/zbittorrent/internal/metainfo"
	"github.com/al002/zbittorrent/internal/storage"
)

type Allocator struct {
	Files       []File
	HasExisting bool
	HasMissing  bool
	Error       error

	closeC chan struct{}
	doneC  chan struct{}
}

type File struct {
	Storage storage.File
	Name    string
	Padding bool
}

type Progress struct {
	AllocatedSize int64
}

func New() *Allocator {
	return &Allocator{
		closeC: make(chan struct{}),
		doneC:  make(chan struct{}),
	}
}

func (a *Allocator) Close() {
	close(a.closeC)
	<-a.doneC
}

func (a *Allocator) Run(info *metainfo.Info, sto storage.Storage, progressC chan Progress, resultC chan *Allocator) {
	defer close(a.doneC)

	defer func() {
		if a.Error != nil {
			for _, f := range a.Files {
				if f.Storage != nil {

					f.Storage.Close()
				}
			}
		}

		select {
		case resultC <- a:
		case <-a.closeC:
		}
	}()

	var allocatedSize int64
	a.Files = make([]File, len(info.Files))
	for i, f := range info.Files {
		var sf storage.File
		var exists bool
		if f.Padding {
			sf = storage.NewPaddingFile(f.Length)
		} else {
			sf, exists, a.Error = sto.Open(f.Path, f.Length)
			if a.Error != nil {
				return
			}

			if exists {
				a.HasExisting = true
			} else {
				a.HasMissing = true
			}
		}
		a.Files[i] = File{
			Storage: sf,
			Name:    f.Path,
			Padding: f.Padding,
		}
		allocatedSize += f.Length
		a.sendProgress(progressC, allocatedSize)
	}
}

func (a *Allocator) sendProgress(progressC chan Progress, size int64) {
	select {
	case progressC <- Progress{
		AllocatedSize: size,
	}:
	case <-a.closeC:
		return
	}
}
