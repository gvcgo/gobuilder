package utils

import (
	"path/filepath"
	"strings"

	"github.com/gvcgo/goutils/pkgs/gutils"
)

func FindGoProjectDir(dirName ...string) string {
	var currentDir string
	sep := string([]rune{filepath.Separator})
	if len(dirName) > 0 && strings.Trim(dirName[0], sep) != "" {
		currentDir = dirName[0]
	} else {
		return currentDir
	}

	modPath := filepath.Join(currentDir, "go.mod")

	if ok, _ := gutils.PathIsExist(modPath); ok {
		return currentDir
	} else {
		parentDir := filepath.Dir(currentDir)
		return FindGoProjectDir(parentDir)
	}
}

func GetCommanlyUsedArchOS() []string {
	return []string{
		"darwin/amd64",
		"darwin/arm64",
		"linux/amd64",
		"linux/arm64",
		"windows/amd64",
		"windows/arm64",
	}
}

func isCommanlyUsed(archOS string) bool {
	cList := GetCommanlyUsedArchOS()
	for _, v := range cList {
		if v == archOS {
			return true
		}
	}
	return false
}

func GetOtherArchOS() []string {
	buff, _ := gutils.ExecuteSysCommand(true, "", "go", "tool", "dist", "list")
	aoList := strings.Split(buff.String(), "\n")
	r := []string{}
	for _, v := range aoList {
		if isCommanlyUsed(v) {
			continue
		}
		r = append(r, v)
	}
	return r
}
