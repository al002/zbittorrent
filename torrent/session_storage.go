package torrent

import (
	"io/fs"

	"github.com/al002/zbittorrent/internal/storage"
	"github.com/al002/zbittorrent/internal/storage/filestorage"
)

type fileStorageProvider struct {
	DataDir         string
	FilePermissions fs.FileMode
}

func newFileStorageProvider(cfg *Config) *fileStorageProvider {
	return &fileStorageProvider{
		DataDir:         cfg.DataDir,
		FilePermissions: cfg.FilePermissions,
	}
}

func (p *fileStorageProvider) GetStorage(torrentID string) (storage.Storage, error) {
	return filestorage.New(p.getDataDir(torrentID), p.FilePermissions)
}

func (p *fileStorageProvider) getDataDir(torrentID string) string {
	//  if p.DataDirIncludesTorrentID {
	// 	return filepath.Join(p.DataDir, torrentID)
	// }
	return p.DataDir
}
