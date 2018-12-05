package web

import (
	"strings"
	"reflect"
	"runtime"
	"github.com/gin-gonic/gin"
	"github.com/jinycoo/jinygo/utils"
	"github.com/jinycoo/jinygo/constants"
)

var (
	Routes = make(map[string]*RuGroup, 0)
	Methods = map[string]string{
		"Index": constants.MethodGet,
		"Create": constants.MethodPost,
		"Show": constants.MethodGet,
		"Update": constants.MethodPut,
		"Delete": constants.MethodDelete,
	}
)

type (
	RuGroup struct {
		Name string
		Child []*Route
	}
	Route struct {
		Name string
		Method string
		Controller gin.HandlerFunc
	}
)

func (g *RuGroup) Resource(resourceName string, controllers ...gin.HandlerFunc) {
	if len(controllers) > 0 {
		relPath := strings.ToLower(resourceName)
		relative := utils.Ucfirst(relPath)
		for _,c := range controllers {
			fn := runtime.FuncForPC(reflect.ValueOf(c).Pointer()).Name()
			runes := []rune(fn)
			length := len(runes)
			llen := strings.LastIndex(fn, relative)
			if llen > -1 {
				flen := llen + len(relative)
				fnName := string(runes[flen:length])
				switch fnName {
				case "Show", "Update", "Delete":
					relativePath := relPath + constants.Separator + strings.ToLower(fnName) + constants.Separator + ":id"
					g.add(Methods[fnName], relativePath, c)
				default:
					g.add(Methods[fnName], relPath + constants.Separator + strings.ToLower(fnName), c)
				}
			}
		}
	}
}

func (g *RuGroup) Get(relativePath string, controller gin.HandlerFunc) {
	g.add(constants.MethodGet, relativePath, controller)
}

func (g *RuGroup) Post(relativePath string, controller gin.HandlerFunc) {
	g.add(constants.MethodPost, relativePath, controller)
}

func (g *RuGroup) Put(relativePath string, controller gin.HandlerFunc) {
	g.add(constants.MethodPut, relativePath, controller)
}
func (g *RuGroup) Del(relativePath string, controller gin.HandlerFunc) {
	g.add(constants.MethodDelete, relativePath, controller)
}

func (g *RuGroup) add(method, relativePath string, controller gin.HandlerFunc) {
	r := &Route {
		Name: constants.Separator + strings.Trim(relativePath, constants.Separator),
		Method: method,
		Controller: controller,
	}
	g.Child = append(g.Child, r)
	Routes[g.Name] = g
}