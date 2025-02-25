package cmd

import (
	"fmt"
	"os"

	"github.com/al002/zbittorrent/pkg/metainfo"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get torrent file metainfo",
	Run: func(cmd *cobra.Command, args []string) {
		mi, err := metainfo.LoadFromFile(args[0])
		if err != nil {
			log.Error("parse torrent file error", err)
			os.Exit(1)
		}

    b := mi.HashInfoBytes()
    fmt.Printf("info hash %s\n", b.HexString())

    info, err := mi.UnmarshalInfo()
    if err != nil {
      fmt.Printf("unmarshal info error %w\n", err)
    }


    fmt.Printf("unmarshal info %v\n", info.Name)
	},
}
