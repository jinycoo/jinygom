package jinygo

import (
	"os"
	"github.com/jinycoo/jinygo/log"
	"github.com/jinycoo/jinygo/utils"
	"github.com/jinycoo/jinygo/constants"
)

var (
	cfg *Config
	root string
)

type Config struct {
	AppName     string                  `yaml:"appName"`
	AppPath     string                  `yaml:"appPath"`
	RunMode     string                  `yaml:"runMode"`
	WebHost     string                  `yaml:"host"`
	WebPort     int                     `yaml:"port"`
	Logger      *log.JLogConfig         `yaml:"log"`
	Components  map[string]string       `yaml:"components"`
}

func init() {
	root = utils.RootDir()
	cfg = &Config {
		AppName: AppName,
		AppPath: root,
		RunMode: constants.RunModeDebug,
		WebPort: Port,
		Components: make(map[string]string),
	}
	if runMode := os.Getenv(constants.RunMode); runMode != "" {
		cfg.RunMode = runMode
	}
}