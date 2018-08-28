package jinygo

import (
	"os"
	"fmt"
	"path"
	"strings"
	"strconv"
	"github.com/jinygo/db"
	"github.com/jinygo/web"
	"github.com/jinygo/log"
	"github.com/jinygo/cache"
	"github.com/jinygo/constants"
)

type Jinygo struct {
	version    string
	basePath   string
	configPath string
	config     *Config
}

func New() *Jinygo {
	var jiny = new(Jinygo)
	jiny.version = AppVersion
	jiny.basePath = root
	jiny.configPath = path.Join(jiny.basePath, ConfigDir)
	return jiny
}

func (jiny *Jinygo) SetEnvPrefix(prefix string) {
	if prefix != "" {
		runModeKey := strings.ToUpper(fmt.Sprintf("%s_%s", prefix, constants.RunMode))
		cfgPathKey := strings.ToUpper(fmt.Sprintf("%s_%s", prefix, EnvKeyConfigDir))
		if runMode := os.Getenv(runModeKey); runMode != "" {
			cfg.RunMode = runMode
		}
		if cfgPath := os.Getenv(cfgPathKey); cfgPath != "" {
			jiny.configPath = cfgPath
		}
	}
}

func (jiny *Jinygo) RGroup(name string) *web.RuGroup {
	return &web.RuGroup{
		Name: constants.Separator + strings.Trim(name, constants.Separator),
		Child: make([]*web.Route, 0),
	}
}

func (jiny *Jinygo) Run(params ...string) {
	jiny.initApp()
	log.New(jiny.config.Logger)
	defer log.Sync()
	if len(jiny.config.Components) > 0 {
		if file := jiny.getModConfigFile(constants.ConfigFileDB); file != "" {
			db.Init(file)
			defer db.Close()
		} else {
			log.Warn("db 配置文件未找到，请检查配置是否正确")
		}
		if file := jiny.getModConfigFile(constants.ConfigFileCache); file != "" {
			cache.Init(file)
		} else {
			log.Warn("db 配置文件未找到，请检查配置是否正确")
		}
	}
	if len(params) > 0 && params[0] != "" {
		addr := strings.Split(params[0], ":")
		if len(addr) > 0 && addr[0] != "" {
			jiny.config.WebHost = addr[0]
		}
		if len(addr) > 1 && addr[1] != "" {
			jiny.config.WebPort, _ = strconv.Atoi(addr[1])
		}
	}
	web.Run(jiny.config.RunMode, fmt.Sprintf("%s:%d", cfg.WebHost, cfg.WebPort))
}