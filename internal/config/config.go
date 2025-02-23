package config

type LogConfig struct {
	Dir    string `mapstructure:"dir"`
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // "text" or "json"
	// MaxSize    int    `mapstructure:"max_size"`
	// MaxAge     int    `mapstructure:"max_age"`
	// MaxBackups int    `mapstructure:"max_backups"`
	// Compress   bool   `mapstructure:"compress"`
	// ToStderr   bool   `mapstructure:"to_stderr"`
}

type TrackerConfig struct {
	Timeout     int `mapstructure:"timeout" validate:"required,min=1"`
	MaxRetries  int `mapstructure:"max_retries" validate:"required,min=0"`
	MinInterval int `mapstructure:"min_interval" validate:"required,min=1"`
}

type PeerConfig struct {
	MaxConnections int `mapstructure:"max_connections" validate:"required,min=1,max=200"`
	Timeout        int `mapstructure:"timeout" validate:"required,min=1"`   // connection timeout in seconds
	KeepAlive      int `mapstructure:"keepalive" validate:"required,min=1"` // keepalive interval in seconds
}

type Config struct {
	ListenPort        int           `mapstructure:"listen_port" validate:"required,min=1024,max=65535"`
	DownloadDir       string        `mapstructure:"download_dir" validate:"required,dir"`

  // Feature flags
	EnableDHT         bool          `mapstructure:"enable_dht"`
	EnablePEX         bool          `mapstructure:"enable_pex"`

  // Rate limits
	UploadRateLimit   int           `mapstructure:"upload_rate_limit"`
	DownloadRateLimit int           `mapstructure:"download_rate_limit"`

	Log               LogConfig     `mapstructure:"log"`
	Tracker           TrackerConfig `mapstructure:"tracker"`
	Peer              PeerConfig    `mapstructure:"peer"`
}
