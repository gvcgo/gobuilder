package builder

import (
	"os"
	"runtime"
	"strings"

	"github.com/gvcgo/goutils/pkgs/gutils"
)

const (
	CGO_Additional_BUILD_ARGS_ENV string = "CGO_ADDITIONAL_BUILD_ARGS"
)

func SupportedArchOSForCGO(osInfo, archInfo string) bool {
	if osInfo == gutils.Linux || osInfo == gutils.Darwin {
		switch archInfo {
		case "amd64", "arm64":
			return true
		default:
			return false
		}
	} else if osInfo == gutils.Windows {
		return archInfo == "amd64"
	}
	return false
}

func IsZigInstalled() bool {
	binName := "zig"
	if runtime.GOOS == gutils.Windows {
		binName += ".exe"
	}
	_, err := gutils.ExecuteSysCommand(true, "", binName, "--help")
	return err == nil
}

func (b *Builder) SetZigForCGO(osInfo, archInfo, binDir, binName string) {
	if !IsZigInstalled() {
		return
	}
	if !SupportedArchOSForCGO(osInfo, archInfo) {
		return
	}
	os.Setenv("CGO_ENABLED", "1")
	// CC='zig cc -target x86_64-linux-musl' CXX='zig cc -target x86_64-linux-musl' CGO_CFLAGS='-D_LARGEFILE64_SOURCE' go build -ldflags='-linkmode=external -extldflags -static'
	if osInfo == gutils.Linux && archInfo == "amd64" {
		os.Setenv("CC", "zig cc -target x86_64-linux-musl")
		os.Setenv("CXX", "zig cc -target x86_64-linux-musl")
		os.Setenv("CGO_CFLAGS", "-D_LARGEFILE64_SOURCE")
		os.Setenv(CGO_Additional_BUILD_ARGS_ENV, `-ldflags='-linkmode=external -extldflags -static'`)
	} else if osInfo == gutils.Linux && archInfo == "arm64" {
		os.Setenv("CC", "zig cc -target aarch64-linux-musl")
		os.Setenv("CXX", "zig cc -target aarch64-linux-musl")
		os.Setenv("CGO_CFLAGS", "-D_LARGEFILE64_SOURCE")
		os.Setenv(CGO_Additional_BUILD_ARGS_ENV, `-ldflags='-linkmode=external -extldflags -static'`)
	} else if osInfo == gutils.Windows && archInfo == "amd64" {
		os.Setenv("CC", "zig cc -target x86_64-windows-gnu")
		os.Setenv("CXX", "zig cc -target x86_64-windows-gnu")
	} else if osInfo == gutils.Darwin && archInfo == "amd64" {
		// https://github.com/ziglang/zig/issues/9050
		os.Setenv("CC", "zig cc -target x86_64-macos-gnu")
		os.Setenv("CXX", "zig cc -target x86_64-macos-gnu")
		os.Setenv(CGO_Additional_BUILD_ARGS_ENV, `-v -x -a -buildmode=pie -ldflags="-s -w"`)
	} else if osInfo == gutils.Darwin && archInfo == "arm64" {
		os.Setenv("CC", "zig cc -target aarch64-macos-gnu")
		os.Setenv("CXX", "zig cc -target aarch64-macos-gnu")
		os.Setenv(CGO_Additional_BUILD_ARGS_ENV, `-v -x -a -buildmode=pie -ldflags="-s -w"`)
	}
}

func (b *Builder) modifyBuildArgsForCGO(args []string) (result []string) {
	buildIdx := findBuild(args)
	if buildIdx == -1 {
		return args
	}

	oldArgs := args[buildIdx+1:]
	filter := map[string]string{}
	for _, arg := range oldArgs {
		if strings.Contains(arg, "=") {
			sList := strings.Split(arg, "=")
			filter[sList[0]] = sList[1]
		} else {
			filter[arg] = ""
		}
	}
	additional := strings.Split(os.Getenv(CGO_Additional_BUILD_ARGS_ENV), " ")

	toAddList := []string{}
	delOld := map[string]struct{}{}
	for _, arg := range additional {
		if strings.Contains(arg, "=") {
			sList := strings.Split(arg, "=")
			name, value := sList[0], sList[1]
			if _, ok := filter[name]; !ok {
				toAddList = append(toAddList, arg)
			} else {
				newArg := name + "" + mergeNamedArgs(filter[name], value)
				toAddList = append(toAddList, newArg)
				delOld[name] = struct{}{}
			}
		} else {
			if _, ok := filter[arg]; !ok {
				toAddList = append(toAddList, arg)
			}
		}
	}

	newArgs := []string{}
	for _, old := range oldArgs {
		if strings.Contains(old, "=") {
			sList := strings.Split(old, "=")
			name := sList[0]
			if _, ok := delOld[name]; ok {
				continue
			}
		}
		newArgs = append(newArgs, old)
	}

	newArgs = append(newArgs, args[:buildIdx+1]...)
	newArgs = append(newArgs, toAddList...)
	newArgs = append(newArgs, toAddList...)
	return newArgs
}

func mergeNamedArgs(oldValue, newValue string) string {
	if !strings.HasPrefix(oldValue, `"`) {
		return newValue
	}
	oldValue = strings.Trim(oldValue, `"`)
	newValue = strings.Trim(newValue, `"`)
	sList := strings.Split(newValue, " ")
	for _, s := range sList {
		if strings.Contains(oldValue, s) {
			continue
		}
		oldValue += " " + s
	}
	return oldValue
}

func findBuild(args []string) (idx int) {
	idx = -1
	for i, arg := range args {
		if arg == "build" {
			return i
		}
	}
	return idx
}
