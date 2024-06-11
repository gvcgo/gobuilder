package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gvcgo/gobuilder/internal/utils"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
)

/* Use xgo for CGO cross-compile.
https://github.com/crazy-max/xgo

crazymax/xgo
ghcr.io/crazy-max/xgo
*/

func FindGoProxy() (p string) {
	return os.Getenv("GOPROXY")
}

func IsXgoInstalled() bool {
	homeDir, _ := os.UserHomeDir()
	_, err := gutils.ExecuteSysCommand(true, homeDir, "xgo", "-h")
	return err == nil
}

func FindXgoDockerImage() (imgName string) {
	b, _ := gutils.ExecuteSysCommand(true, "", "docker", "images")
	for _, line := range strings.Split(b.String(), "\n") {
		if strings.Contains(line, "crazy-max/xgo") {
			sList := strings.Split(line, " ")
			return sList[0]
		}
	}
	return
}

func (b *Builder) UseXGO(osInfo, archInfo, binDir, binName string, oldArgs []string) (newArgs []string) {
	if !IsXgoInstalled() {
		gprint.PrintWarning("xgo is not installed.")
		os.Exit(1)
	}
	imgName := b.XGoImage
	if imgName == "" {
		imgName = FindXgoDockerImage()
	}
	if imgName == "" {
		gprint.PrintWarning("xgo docker image is not found.")
		os.Exit(1)
	}

	goProxy := FindGoProxy()

	var (
		ldflags  string
		vv       string
		xx       string
		trimpath string
	)

	for idx, arg := range oldArgs {
		if arg == "-ldflags" && idx < len(oldArgs)-1 {
			ldflags = oldArgs[idx+1]
		} else if strings.HasPrefix(arg, "-ldflags=") {
			sList := strings.Split(arg, "=")
			ldflags = sList[1]
		} else if arg == "-v" {
			vv = "-v"
		} else if arg == "-x" {
			xx = "-x"
		} else if arg == "-trimpath" {
			trimpath = "-trimpath"
		}
	}

	importDir := strings.ReplaceAll(oldArgs[len(oldArgs)-1], b.WorkDir, "")
	if importDir == "" {
		importDir = "."
	}

	targets := fmt.Sprintf("%s/%s", osInfo, archInfo)

	newArgs = append(newArgs, "xgo", "-race")
	if b.XGoDeps != "" {
		newArgs = append(newArgs, fmt.Sprintf(`-deps=%s`, b.XGoDeps))
	}
	if b.XGoDepsArgs != "" {
		newArgs = append(newArgs, fmt.Sprintf(`-depsargs=%s`, b.XGoDepsArgs))
	}

	destDir := strings.ReplaceAll(binDir, b.WorkDir, "")
	newArgs = append(newArgs, fmt.Sprintf(`-dest=%s`, strings.TrimPrefix(destDir, utils.GetPathSeparator())))

	newArgs = append(newArgs, fmt.Sprintf(`-docker-image=%s`, imgName))

	if goProxy != "" {
		newArgs = append(newArgs, fmt.Sprintf(`-goproxy=%s`, goProxy))
	}

	if ldflags != "" {
		newArgs = append(newArgs, fmt.Sprintf(`-ldflags=%s`, ldflags))
	}

	newArgs = append(newArgs, fmt.Sprintf(`-out=%s`, binName))

	newArgs = append(newArgs, fmt.Sprintf(`-targets=%s`, targets))

	if trimpath != "" {
		newArgs = append(newArgs, trimpath)
	}

	if vv != "" {
		newArgs = append(newArgs, vv)
	}

	if xx != "" {
		newArgs = append(newArgs, xx)
	}

	newArgs = append(newArgs, importDir)

	fmt.Println(newArgs)
	return
}

func (b *Builder) FixBinaryName(osInfo, archInfo, binDir, binName string) {
	dList, _ := os.ReadDir(binDir)
	for _, d := range dList {
		if !d.IsDir() && strings.Contains(d.Name(), binName) && strings.Contains(d.Name(), osInfo) {
			binPath := filepath.Join(binDir, d.Name())
			newBinPath := filepath.Join(binDir, binName)
			os.Rename(binPath, newBinPath)
		}
	}
	if osInfo != gutils.Windows {
		user := os.Getenv("USER")
		if user == "" {
			return
		}
		gutils.ExecuteSysCommand(true, b.WorkDir, "chown", user, filepath.Join(binDir, binName))
	}
}
