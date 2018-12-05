package jinygo

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/jinycoo/jinygo/log"
	"strings"
	"reflect"
	"strconv"
)

var Params map[string]interface{}

func GetParams(key string) interface{} {
	if Params != nil {
		if out, ok := Params[key]; ok {
			return out
		}
	}
	return nil
}
func GetParamInt(key string) int {
	var result = 0
	keys := strings.Split(key, ".")
	switch len(keys) {
	case 1:
		val := GetParams(keys[0])
		if val != nil {
			result = int(val.(float64))
		}
	case 2:
		valBase := GetParams(keys[0]).(map[interface {}]interface{})
		if valBase != nil {
			if val, ok := valBase[keys[1]]; ok {
				if reflect.TypeOf(val).Name() == "int" {
					result = val.(int)
				}
			}
		}
	default:
		result = 0
	}
	return result
}
func GetParamFloat(key string) float64 {
	var result float64 = 0
	keys := strings.Split(key, ".")
	switch len(keys) {
	case 1:
		val := GetParams(keys[0])
		if val != nil {
			result = val.(float64)
		}
	case 2:
		valBase := GetParams(keys[0]).(map[interface {}]interface{})
		if valBase != nil {
			if val, ok := valBase[keys[1]]; ok {
				typeName := reflect.TypeOf(val).Name()
				if typeName == "float64" || typeName == "int" {
					result = val.(float64)
				}
			}
		}
	default:
		result = 0
	}
	return result
}
func GetParamString(key string) string {
	var result = ""
	keys := strings.Split(key, ".")
	switch len(keys) {
	case 1:
		val := GetParams(keys[0])
		if val != nil {
			result = val.(string)
		}
	case 2:
		valBase := GetParams(keys[0]).(map[interface {}]interface{})
		if valBase != nil {
			if val, ok := valBase[keys[1]]; ok {
				switch val.(type) {
				case int:
					result = strconv.Itoa(val.(int))
				case float64:
					result = strconv.FormatFloat(val.(float64), 'f', -1, 64)
				case string:
					result = val.(string)
				default:
					result = ""
				}
			}
		}
	default:
		result = ""
	}
	return result
}

func initParams(file string) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		log.Warn(file + "文件读取失败")
	}
	err = yaml.Unmarshal(buf, &Params)
	if err != nil {
		log.Warn(file + "解析失败")
	}
}