package cache

import (
	"net"
	"github.com/go-redis/redis"
	"github.com/jinycoo/jinygo/log"
)

var (
	RCache *redis.Client
	redisCfg *RedisConfig
)

type RedisMaster struct {
	Protocol  string  `yaml:"protocol"`
	Host      string  `yaml:"host"`
	Port      string  `yaml:"port"`
	Password  string  `yaml:"password"`
	Db        int     `yaml:"db"`
}

type RedisConfig struct {
	Cluster    string        `yaml:"cluster"`
	Master     *RedisMaster  `yaml:"master"`
	Sentinel   []string      `yaml:"sentinel"`
}

func redisConn(cluster string) {
	if redisCfg.Master != nil {
		master := redisCfg.Master
		switch cluster {
		case "", "standalone":
			RCache = redis.NewClient(&redis.Options{
				Network: master.Protocol,
				Addr: net.JoinHostPort(master.Host, master.Port),
				Password: master.Password,
				DB: master.Db,
			})
		case "sentinel":
			if redisCfg.Sentinel != nil {
				RCache = redis.NewFailoverClient(&redis.FailoverOptions{
					MasterName: master.Host,
					Password: master.Password,
					DB: master.Db,
					SentinelAddrs: redisCfg.Sentinel,
				})
			} else {
				log.Error("cache config setting error")
			}
		default:
			log.Error("cache config - cluster setting error")
		}
	} else {
		log.Error("cache config - master must be setting")
	}
	if RCache != nil {
		_, err := RCache.Ping().Result()
		if err != nil {
			log.Warn(err)
		}
		log.Info("successful connection to redis-server")
	}
}