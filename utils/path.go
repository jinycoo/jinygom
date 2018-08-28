package utils

import (
	"os"
	"strings"
	"path/filepath"
)

func RootDir() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return strings.Replace(dir, "\\", "/", -1)
}