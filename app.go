package jinygo

import (
	"os"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"gopkg.in/yaml.v2"
)

const (
	AppName    = "jinygo"
	AppVersion = "0.0.1"

	Port = 8080
	ConfigDir      = "conf"
	ConfigFileName = "app"
	ConfigFileType = "yml"

	EnvKeyConfigDir = "config_path"
)

func (jiny *Jinygo) getModConfigFile(name string) string {
	var file string
	if name != "" {
		if mod,ok := jiny.config.Components[name]; ok {
			filename := fmt.Sprintf("%s.%s", mod, ConfigFileType)
			file = filepath.Join(jiny.configPath, filename)
			if _, err := os.Stat(file); err != nil {
				return file
			}
		}
	}
	return file
}

func (jiny *Jinygo) initApp() {
	configFile := fmt.Sprintf("%s.%s", ConfigFileName, ConfigFileType)
	cfgFile := filepath.Join(jiny.configPath, configFile)
	if _, err := os.Stat(cfgFile); err == nil {
		buf, _ := ioutil.ReadFile(cfgFile)
		yaml.Unmarshal(buf, &cfg)
		if cfg.WebPort == 0 {
			cfg.WebPort = Port
		}
		if cfg.Logger != nil {
			if cfg.Logger.LogPath == "" {
				cfg.Logger.LogPath = root
			}
			if cfg.Logger.LogFile == "" {
				cfg.Logger.LogFile = cfg.AppName
			}
		}
	}
	jiny.config = cfg
}