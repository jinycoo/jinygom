package cache

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/jinygo/log"
)

const (
	Redis = "redis"
	Mongo = "mongo"
)

var cacheCfg *CheConfig

type CheConfig struct {
	Redis  *RedisConfig  `yaml:"redis"`
	Mongo  *mongoConfig  `yaml:"mongo"`
}

func Init(cfgFile string) {
	buf, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.Warn(cfgFile + "文件读取失败")
	}
	err = yaml.Unmarshal(buf, &cacheCfg)
	if err != nil {
		log.Warn(cfgFile + "解析失败")
	}
	redisCfg = cacheCfg.Redis
	mongoCfg = cacheCfg.Mongo
	if redisCfg != nil {
		redisConn(redisCfg.Cluster)
	}
	if mongoCfg != nil {
		mongoConn("")
	}
}

func InitCache(cfg *CheConfig) {
	redisCfg = cfg.Redis
	mongoCfg = cfg.Mongo
	if redisCfg != nil {
		redisConn(redisCfg.Cluster)
	}
	if mongoCfg != nil {
		mongoConn("")
	}
}