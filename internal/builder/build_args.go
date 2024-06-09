package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gvcgo/goutils/pkgs/gutils"
)

const (
	winSuffix string = ".exe"
)

/*
prepare build args
*/
func (b *Builder) handleInjections(args []string) {
	// handle injection args
	cmdReg := regexp.MustCompile(`(\$\([\w\W]+?\))`)
	for idx, v := range args {
		l := cmdReg.FindAllString(v, -1)
		for _, cc := range l {
			ccc := strings.TrimLeft(cc, "$(")
			ccc = strings.TrimRight(ccc, ")")
			bf, _ := gutils.ExecuteSysCommand(true, b.WorkDir, strings.Split(ccc, " ")...)
			v = strings.ReplaceAll(v, cc, bf.String())
		}
		if len(l) > 0 {
			args[idx] = v
		}
	}
}

func (b *Builder) PrepareArgs(osInfo, archInfo string) (args []string, targetDir, binName string) {
	inputArgs := append([]string{}, b.BuildArgs...) // deepcopy

	if len(inputArgs) == 0 {
		inputArgs = append(inputArgs, b.WorkDir)
	}

	// main func position.
	lastArg := inputArgs[len(inputArgs)-1]

	if !strings.HasPrefix(lastArg, string([]rune{filepath.Separator})) && !strings.HasPrefix(lastArg, ".") {
		inputArgs = append(inputArgs, b.WorkDir)
		lastArg = b.WorkDir
	} else if lastArg == "." && b.WorkDir != "" {
		inputArgs[len(inputArgs)-1] = b.WorkDir
		lastArg = b.WorkDir
	} else if lastArg == ".." && b.WorkDir != "" {
		inputArgs[len(inputArgs)-1] = filepath.Dir(b.WorkDir)
		lastArg = b.WorkDir
	}

	for idx, arg := range b.BuildArgs {
		if arg == "-o" && len(b.BuildArgs) > idx+1 {
			inputArgs = append(inputArgs[:idx], inputArgs[idx+2:]...)
			// If binName has been specified.
			binName = filepath.Base(b.BuildArgs[idx+1])
		}
	}

	if binName == "" {
		binName = filepath.Base(lastArg)
	}

	targetDir = filepath.Join(b.ProjectDir(), "build", fmt.Sprintf("%s-%s", osInfo, archInfo))
	os.MkdirAll(targetDir, os.ModePerm)

	target := targetDir
	if binName != "" {
		if osInfo == gutils.Windows && !strings.HasSuffix(binName, winSuffix) {
			binName += winSuffix
		}
		target = filepath.Join(targetDir, binName)
	}

	if len(inputArgs) == 1 {
		inputArgs = append([]string{"-o", target}, inputArgs...)
		return inputArgs, targetDir, binName
	}

	maxIndex := len(inputArgs) - 1
	front := inputArgs[:maxIndex]
	// incase overwrite
	back := append([]string{}, inputArgs[maxIndex:]...)
	inputArgs = append(append(front, "-o", target), back...)
	return inputArgs, targetDir, binName
}
