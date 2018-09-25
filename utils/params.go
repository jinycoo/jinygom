package utils

import (
	"strconv"
)

func ParseParamInt8(v interface{}, def int8) int8 {
	var val = def
	switch v.(type) {
	case string:
		res, err := strconv.Atoi(v.(string))
		if err == nil {
			val = int8(res)
		}
	case float64:
		val = int8(v.(float64))
	default:
		val = def
	}
	return val
}

func ParseParamInt(v interface{}, def int) int {
	var val = def
	switch v.(type) {
	case string:
		res, err := strconv.Atoi(v.(string))
		if err == nil {
			val = res
		}
	case float64:
		val = int(v.(float64))
	default:
		val = def
	}
	return val
}

func ParseParamFloat64(v interface{}, def float64) float64 {
	var val = def
	switch v.(type) {
	case string:
		res, err := strconv.ParseFloat(v.(string), 64)
		if err == nil {
			val = res
		}
	case float64:
		val = v.(float64)
	default:
		val = def
	}
	return val
}

func ParseParamString(v interface{}, def string) string {
	var val = def
	switch v.(type) {
	case string:
		val = v.(string)
	case float64:
		val = strconv.FormatFloat(v.(float64), 'f', -1, 64)
	case int:
		val = strconv.Itoa(v.(int))
	default:
		val = def
	}
	return val
}