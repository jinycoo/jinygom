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

var cacheCfg *caches

type caches struct {
	Redis  *redisConfig  `yaml:"redis"`
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