package web

import (
	"strconv"
	"io/ioutil"
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
	"github.com/jinygo/log"
	"github.com/jinygo/errno"
)

type InContext struct {
	Ctx       *gin.Context
	ApiRes    *ApiResponse
	ApiOldRes *OldApiResponse
}

type ApiResponse struct {
	ErrCode  int `json:"msg_code"`
	Message string `json:"message"`
	Data    interface{} `json:"attachment"`
}

type OldApiResponse struct {
	ErrCode  int `json:"status"`
	Message string `json:"message"`
	Data    interface{} `json:"attachment"`
}

func (ic *InContext) QueryInt(pkey string, def int) int {
	val := ic.Ctx.DefaultQuery(pkey, strconv.Itoa(def))
	queryKey, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return queryKey
}

func (ic *InContext) QueryString(pkey string, def string) string {
	return ic.Ctx.DefaultQuery(pkey, def)
}

func (ic *InContext) ParamsInt(pname string, def int) int {
	val := ic.Ctx.Param(pname)
	queryKey, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return queryKey
}
func (ic *InContext) ParamsString(pname, def string) string {
	val := ic.Ctx.Param(pname)
	if val == "" {
		return def
	}
	return val
}
func (ic *InContext) JsonPost() map[string]interface{} {
	body := ic.Ctx.Request.Body
	var params map[string]interface{}
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = ioutil.ReadAll(body)
		jsoniter.Unmarshal(bodyBytes, &params)
	}
	return params
}

func (ic *InContext) JsonResponse() {
	if ic.ApiOldRes != nil {
		if ic.ApiOldRes.Data == nil {
			ic.ApiOldRes.Data = gin.H{}
		}
		b, _ := jsoniter.Marshal(ic.ApiOldRes)
		log.Info(string(b))
		ic.Ctx.JSON(200, ic.ApiOldRes)
	} else if ic.ApiRes != nil {
		if ic.ApiRes.Data == nil {
			ic.ApiRes.Data = gin.H{}
		}
		b, _ := jsoniter.Marshal(ic.ApiRes)
		log.Info(string(b))
		ic.Ctx.JSON(200, ic.ApiRes)
	} else {
		ic.ApiRes = &ApiResponse{
			ErrCode: 0,
			Message: errno.ErrCode[0],
			Data: gin.H{},
		}
		b, _ := jsoniter.Marshal(ic.ApiRes)
		log.Info(string(b))
		ic.Ctx.JSON(200, ic.ApiRes)
	}
}
/**
 * 默认首页
 */
func JsonHandleIndex(c *gin.Context) {
	c.JSON(200, &ApiResponse{
		ErrCode: 0,
		Message: errno.ErrCode[1],
		Data: gin.H{},
	})
}
/**
 * 404 处理
 */
func JsonHandle404(c *gin.Context) {
	c.JSON(404, &ApiResponse{
		ErrCode: 404,
		Message: errno.ErrCode[404],
		Data: gin.H{},
	})
}