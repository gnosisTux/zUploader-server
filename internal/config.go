package internal

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	UploadDir   string `toml:"upload_dir"`
	Host        string `toml:"host"`
	Port        int    `toml:"port"`
	MaxUploadMB int    `toml:"max_upload_mb"`
}

var ConfigData Config
var MaxUploadSize int64

func LoadConfig(path string) {
	if _, err := toml.DecodeFile(path, &ConfigData); err != nil {
		fmt.Println("Error loading config.toml:", err)
		os.Exit(1)
	}
	MaxUploadSize = int64(ConfigData.MaxUploadMB) * 1024 * 1024
}
