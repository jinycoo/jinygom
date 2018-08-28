package db

import (
	"fmt"
	"strings"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/go-xorm/xorm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinygo/log"
)

var (
	dbCfg  *dbConfig
	DataGroup map[string]*xorm.EngineGroup
)

type dbGroupConfig struct {
	OpenConns  int `yaml:"openConns"`
	IdleConns  int `yaml:"idleConns"`
	Master     *engineConfig `yaml:"master`
	Slaves     []string `yaml:"slaves"`
}
type dbConfig struct {
	Adapter string `yaml:"adapter"`
	Db      map[string]*dbGroupConfig   `yaml:"db"`
}
type engineConfig struct {
	Dsn      string  `yaml:"dsn"`
	Username string  `yaml:"username"`
	Password string  `yaml:"password"`
	Protocol string  `yaml:"protocol"`
	Addr     string  `yaml:"addr"`
	Host     string  `yaml:"host"`
	Port     int     `yaml:"port"`
	Params   map[string]string  `yaml:"params"`
}

func initDataGroup() map[string]*xorm.EngineGroup {
	var groups = make(map[string]*xorm.EngineGroup)
	for g, e := range dbCfg.Db {
		dataSourceSlice := make([]string, 0)
		dataSourceSlice = append(dataSourceSlice, e.Master.parseDns(g))
		for _, sn := range dbCfg.Db[g].Slaves {
			dataSourceSlice = append(dataSourceSlice, sn)
		}
		if len(dataSourceSlice) > 0 {
			group, err := xorm.NewEngineGroup(dbCfg.Adapter, dataSourceSlice)
			if err != nil {
				log.Warn("创建数据组链接错误：" + err.Error())
			}
			group.SetMaxOpenConns(dbCfg.Db[g].OpenConns)
			group.SetMaxIdleConns(dbCfg.Db[g].IdleConns)
			groups[g] = group
			log.Info(fmt.Sprintf("%s EngineGroup Opened", g))
		}
	}
	return groups
}

func Use(dbName string) *xorm.Engine {
	if DataGroup == nil {
		DataGroup = initDataGroup()
	}
	if g, ok := DataGroup[dbName]; ok {
		return g.Engine
	} else {
		log.Error(dbName + " - Database does not exist.")
	}
	return nil
}

func Init(dbCfgFile string) {
	buf, err := ioutil.ReadFile(dbCfgFile)
	if err != nil {
		log.Warn(dbCfgFile + "文件读取失败")
	}
	err = yaml.Unmarshal(buf, &dbCfg)
	if err != nil {
		log.Warn(dbCfgFile + "解析失败")
	}
	DataGroup = initDataGroup()
}
func (e *engineConfig) parseDns(dbname string) string {
	if e.Dsn == "" {
		var addr string
		switch e.Protocol {
		case "", "tcp":
			if e.Host == "" {
				e.Host = "127.0.0.1"
			}
			if e.Port == 0 {
				e.Port = 3306
			}
			addr = fmt.Sprintf("tcp(%s:%d)", e.Host, e.Port)
		default:
			addr = fmt.Sprintf("%s(%s)", e.Protocol, e.Addr)
		}
		var params= make([]string, 0)
		for k, v := range e.Params {
			params = append(params, fmt.Sprintf("%s=%v", k, v))
		}
		var dsnParams = ""
		if len(params) > 0 {
			dsnParams = "?" + strings.Join(params, "&")
		}
		e.Dsn = fmt.Sprintf("%s:%s@%s/%s%s",
			e.Username,
			e.Password,
			addr,
			dbname,
			dsnParams,
		)
	}
	return e.Dsn
}

func Close() {
	for n,db := range DataGroup {
		db.Close()
		log.Info(fmt.Sprintf("%s EngineGroup Closed", n))
	}
}