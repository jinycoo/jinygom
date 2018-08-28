package web

import (
	"time"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
	"github.com/jinygo/log"
	"github.com/jinygo/constants"
)

func Run(runMode, addr string) {
	gin.SetMode(runMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(incLogger())
	r.NoRoute(JsonHandle404)
	if len(Routes) == 0 {
		r.GET(constants.Separator, JsonHandleIndex)
	} else {
		if _,ok := Routes[constants.Separator]; !ok {
			r.GET(constants.Separator, JsonHandleIndex)
		}
		for _,v := range Routes {
			module := r.Group(v.Name)
			for _,re := range v.Child {
				module.Handle(re.Method, re.Name, re.Controller)
			}
		}
	}
	log.Info("Listening and serving HTTP on " + addr)
	r.Run(addr)
}

func incLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req  = make(map[string]interface{}, 0)
		req["start"] = time.Now().Unix()
		req["action"] = c.Request.URL.Path
		req["method"] = c.Request.Method
		req["query"] = strings.Split(c.Request.URL.RawQuery, "&")
		req["clientIP"] = c.ClientIP()
		//body := c.Request.Body
		//var bodyBytes []byte
		//if body != nil {
		//	var params map[string]interface{}
		//	bodyBytes, _ = ioutil.ReadAll(body)
		//	jsoniter.Unmarshal(bodyBytes, &params)
		//	req["params"] = params
		//}
		b,_ := jsoniter.Marshal(req)
		log.Info(string(b))
		c.Next()
		var res  = make(map[string]interface{}, 0)
		res["end"] = time.Now().Unix()
		res["status"] = c.Writer.Status()
		b2,_ := jsoniter.Marshal(res)
		log.Info(string(b2))
	}
}