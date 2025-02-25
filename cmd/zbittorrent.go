package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/al002/zbittorrent/internal/config"
	"github.com/al002/zbittorrent/internal/logger"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	cfgRegistry *config.Registry
	cfg         *config.Config
	log         *logger.Logger

	rootCmd = &cobra.Command{
		Use:   "zbittorrent",
		Short: "A simple bittorrent client",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("zbittorrent is starting...")
		},
	}
)

func Execute() error {
	defer func() {
		if log != nil {
			if err := log.Close(); err != nil {
				fmt.Printf("Failed to close logger: %v\n", err)
			}
		}
	}()
	return rootCmd.Execute()
}

func init() {
	cfgRegistry = config.NewRegistry()
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.zbittorrent/config.yaml)")

  rootCmd.AddCommand(validateCmd)
  rootCmd.AddCommand(infoCmd)
}

func initConfig() {
	var err error
	cfg, err = cfgRegistry.LoadConfig(cfgFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err = logger.New(&cfg.Log)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	log.Debug("Configuration loaded successfully",
		"config_file", cfgRegistry.ConfigFile(),
		"listen_port", cfg.ListenPort,
		"download_dir", cfg.DownloadDir,
		"enable_dht", cfg.EnableDHT,
		"enable_pex", cfg.EnablePEX,
		"upload_rate_limit", cfg.UploadRateLimit,
		"download_rate_limit", cfg.DownloadRateLimit,
	)

	dirs := []string{cfg.DownloadDir}
	for _, dir := range dirs {
		if err := verifyDirectory(dir); err != nil {
			log.Error("Directory verification failed", "dir", dir, "error", err)
			os.Exit(1)
		}
	}
}

func verifyDirectory(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("Failed to create directory: %v", err)
	}

	tmpFile := filepath.Join(dir, ".zbittorrent_write_test")
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %v", err)
	}
	f.Close()
	os.Remove(tmpFile)

	return nil
}
