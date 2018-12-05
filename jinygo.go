package jinygo

import (
	"os"
	"fmt"
	"path"
	"strings"
	"strconv"
	"github.com/jinycoo/jinygo/db"
	"github.com/jinycoo/jinygo/web"
	"github.com/jinycoo/jinygo/log"
	"github.com/jinycoo/jinygo/cache"
	"github.com/jinycoo/jinygo/mqueue"
	"github.com/jinycoo/jinygo/constants"
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
		if dbFile, ok := jiny.config.Components[constants.ConfigFileDB]; ok && dbFile != "" {
			if file := jiny.getModConfigFile(dbFile); file != "" {
				db.Init(file)
				defer db.Close()
			} else {
				log.Error(dbFile + ".yml 配置文件未找到，请检查配置是否正确")
			}
		}
		if cacheFile, ok := jiny.config.Components[constants.ConfigFileCache]; ok && cacheFile != "" {
			if file := jiny.getModConfigFile(cacheFile); file != "" {
				cache.Init(file)
			} else {
				log.Error(cacheFile + ".yml 配置文件未找到，请检查配置是否正确")
			}
		}
		if mqFile, ok := jiny.config.Components[constants.ConfigFileMQ]; ok && mqFile != "" {
			if file := jiny.getModConfigFile(mqFile); file != "" {
				mqueue.Init(file)
				defer mqueue.Mqueue.Close()
			} else {
				log.Error(mqFile + ".yml 配置文件未找到，请检查配置是否正确")
			}
		}
		if paramsFile, ok := jiny.config.Components[constants.ConfigFileParams]; ok && paramsFile != "" {
			if file := jiny.getModConfigFile(paramsFile); file != "" {
				initParams(file)
			} else {
				log.Error(paramsFile + ".yml 配置文件未找到，请检查配置是否正确")
			}
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