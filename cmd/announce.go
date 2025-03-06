package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/al002/zbittorrent/torrent"
	"github.com/spf13/cobra"
)

var announceCmd = &cobra.Command{
	Use:   "announce",
	Short: "Announce to torrent trackers",
	Run: func(cmd *cobra.Command, args []string) {
		fp := args[0]
		f, err := os.Open(fp)
		if err != nil {
      fmt.Printf("open file err: %w\n", err)
			os.Exit(1)
		}
		defer f.Close()

		var buf bufio.Reader
		buf.Reset(f)

    s, err := torrent.NewSession(torrent.DefaultConfig, *log)
    if err != nil {
      fmt.Printf("new session error: %w\n", err)
    }

    s.AddTorrent(&buf, &torrent.AddTorrentOptions{
      ID: "test1",
    })

    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

    for {
      select {
      case <-ch:
        return 
      }
    }
	},
}
