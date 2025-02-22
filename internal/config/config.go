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

type Config struct {
	ListenPort        int       `mapstructure:"listen_port" validate:"required,min=1024,max=65535"`
	EnableDHT         bool      `mapstructure:"enable_dht"`
	EnablePEX         bool      `mapstructure:"enable_pex"`
	MaxUploads        int       `mapstructure:"max_uploads"`
	MaxPeers          int       `mapstructure:"max_peers" validate:"required,min=10,max=200"`
	MinPeers          int       `mapstructure:"min_peers" validate:"required,min=1"`
	DownloadDir       string    `mapstructure:"download_dir" validate:"required,dir"`
	UploadRateLimit   int       `mapstructure:"upload_rate_limit"`
	DownloadRateLimit int       `mapstructure:"download_rate_limit"`
	Log               LogConfig `mapstructure:"log"`
}
