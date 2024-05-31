package internal

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gvcgo/goutils/pkgs/gutils"
)

func IsGoCompilerInstalled() bool {
	_, err := gutils.ExecuteSysCommand(true, "", "go", "version")
	return err == nil
}

func IsUPXInstalled() bool {
	_, err := gutils.ExecuteSysCommand(true, "", "upx", "--version")
	return err == nil
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

var CurrentWorkingDirEnv string = "GBER_CURRENT_WORKING_DIR"

func SetCurrentWorkingDir(dPath string) {
	os.Setenv(CurrentWorkingDirEnv, dPath)
}

func GetCurrentWorkingDir() string {
	return os.Getenv(CurrentWorkingDirEnv)
}

func CheckAndInstallGarble() {
	garbleBin := "garble"
	if runtime.GOOS == gutils.Windows {
		garbleBin += ".exe"
	}
	_, err := gutils.ExecuteSysCommand(true, "", "garble", "version")
	if err != nil {
		// install garble
		gutils.ExecuteSysCommand(true, "", "go", "install", "mvdan.cc/garble@latest")
	}
}

func IsOsslsigncodeInstalled() (ok bool) {
	binName := "osslsigncode"
	if runtime.GOOS == gutils.Windows {
		binName += ".exe"
	}
	_, err := gutils.ExecuteSysCommand(true, "", binName, "--version")
	return err == nil
}
