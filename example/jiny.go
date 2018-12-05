package main

import (
	"github.com/jinycoo/jinygo"
	"github.com/jinycoo/jinygo/log"
	"github.com/gin-gonic/gin"
)

func main() {
	jiny := jinygo.New()
	jiny.SetEnvPrefix("jiny")
	v1 := jiny.RGroup("v1")
	{
		v1.Get("index", Index)
	}
	jiny.Run()
}

func Index(c *gin.Context) {
	// db.Use("jiny_db").Table("jiny_table")....
	//cache.RCache.Set("dd", "dddddd", 0)
	//log.Info(cache.RCache.Get("dd"))
	log.Info("good")
	c.JSON(200, gin.H{"redis-test": "good"})
}
