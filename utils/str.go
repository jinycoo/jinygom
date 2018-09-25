package utils

import (
	"unicode"
	"strings"
)

func Ucfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

func Lcfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

func CamelStrings(s string) string {
	var result string
	if re := strings.Replace(s, " ", "", -1); re != "" {
		for i, v := range strings.Split(re, "_") {
			if i == 0 {
				result = v
			} else {
				result += Ucfirst(v)
			}
		}
	}
	return result
}

func SnakeString(s string) string {
	j := false
	num := len(s)
	data := make([]byte, 0, num*2)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}