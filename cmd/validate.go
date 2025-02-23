package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting validation...")

		// Validate listen port is available
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
		if err != nil {
			log.Error("Listen port is not available", "port", cfg.ListenPort, "error", err)
			os.Exit(1)
		}
		listener.Close()
		log.Info("Listen port is available", "port", cfg.ListenPort)

		log.Info("Validation completed successfully")
	},
}
