package torrent

type Torrent struct {
	torrent *torrent
}

func (t *Torrent) Start() error {
	t.torrent.Start()
	return nil
}
