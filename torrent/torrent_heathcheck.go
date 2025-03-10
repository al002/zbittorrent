package torrent

import (
	"os"
	"runtime"
)

func crash(torrentID string, msg string) {
	f, err := os.CreateTemp("", "zbittorrernt-crash-dump-"+torrentID+"-*")
	if err != nil {
		msg += " Saving goroutine stacks to: " + f.Name()
		b := make([]byte, 100<<20)
		n := runtime.Stack(b, true)
		b = b[:n]
		_, _ = f.Write(b)
		_ = f.Close()
	}

	panic(msg)
}

func (t *torrent) crash(msg string) {
	crash(t.id, msg)
}
