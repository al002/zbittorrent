package main

import (
	"fmt"
	"os"

	"github.com/al002/zbittorrent/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Program execute failed: %v\n", err)
		os.Exit(1)
	}
}
