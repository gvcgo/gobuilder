package internal

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gvcgo/goutils/pkgs/gtea/confirm"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gtea/selector"
	"github.com/gvcgo/goutils/pkgs/gutils"
)

var projectDir string

type GoBuilder struct {
	ArchOSList     []string `json:"arch_os_list"`
	EnableCompress bool     `json:"compress"`
	EnableUPX      bool     `json:"upx"`
	BuildArgs      []string `json:"build_args"`
	WorkDir        string   `json:"work_dir"`
}

func NewGoBuilder() (g *GoBuilder) {
	g = &GoBuilder{
		ArchOSList: []string{},
		BuildArgs:  []string{},
	}
	g.LoadBuilder()
	return
}

func (g *GoBuilder) ProjectDir() string {
	if projectDir == "" {
		projectDir = FindGoProjectDir(GetCurrentWorkingDir())
	}
	// If no go project found, exit.
	if projectDir == "" {
		os.Exit(1)
	}
	return projectDir
}

func (g *GoBuilder) LoadBuilder() {
	projectDir := g.ProjectDir()
	buildFile := filepath.Join(projectDir, "build", "gbuild.json")
	if ok, _ := gutils.PathIsExist(buildFile); ok {
		data, _ := os.ReadFile(buildFile)
		if err := json.Unmarshal(data, g); err != nil {
			gprint.PrintError("Failed to load build config file: %+v", err)
			os.Exit(1)
		}
	} else {
		g.saveBuilder(buildFile)
	}
}

func (g *GoBuilder) saveBuilder(buildFile string) {
	buildDir := filepath.Dir(buildFile)
	if ok, _ := gutils.PathIsExist(buildDir); !ok {
		os.MkdirAll(buildDir, os.ModePerm)
	}
	// Choose target Os/Arch
	items := selector.NewItemList()
	items.Add("Current Os/Arch Only", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
	items.Add("Frequently Used Os/Arch", "frequently")
	others := GetOtherArchOS()
	for _, osArch := range others {
		if osArch != "" {
			items.Add(osArch, osArch)
		}
	}
	sel := selector.NewSelector(
		items,
		selector.WithTitle("Select your target Os/Arch: "),
		selector.WidthEnableMulti(true),
		selector.WithEnbleInfinite(true),
		selector.WithWidth(40),
		selector.WithHeight(20),
	)
	sel.Run()

	targetOsArch := sel.Values()
	if len(targetOsArch) > 0 {
		g.ArchOSList = []string{}
	}

	for _, osArch := range targetOsArch {
		if oa, ok := osArch.(string); ok {
			if oa == "frequently" {
				g.ArchOSList = append(g.ArchOSList, GetCommanlyUsedArchOS()...)
			} else {
				g.ArchOSList = append(g.ArchOSList, oa)
			}
		}
	}

	// Enable zip compress.
	cfm := confirm.NewConfirm(confirm.WithTitle("To zip binaries or not?"))
	cfm.Run()
	g.EnableCompress = cfm.Result()

	// Enable upx.
	cfm = confirm.NewConfirm(confirm.WithTitle("To pack binaries with UPX or not?"))
	cfm.Run()
	g.EnableUPX = cfm.Result()

	g.parseArgs()

	g.WorkDir = GetCurrentWorkingDir()

	// Save build info to build file.
	data, _ := json.MarshalIndent(g, "", "    ")
	if len(data) > 0 {
		os.WriteFile(buildFile, data, os.ModePerm)
	}
}

func (g *GoBuilder) parseArgs() {
	args := []string{}
	if len(os.Args) > 2 && os.Args[1] == "build" {
		args = os.Args[2:]
	}
	if len(args) == 0 {
		return
	}

	g.WorkDir = GetCurrentWorkingDir()

	for idx, v := range args {
		v = strings.ReplaceAll(v, "#", "$")
		args[idx] = v
	}

	g.BuildArgs = args
}

func (g *GoBuilder) PackWithUPX(osInfo, archInfo, binDir string) {
	if !IsUPXInstalled() {
		gprint.PrintWarning("upx if not found!")
		return
	}
	var ok bool
	if osInfo == gutils.Windows && archInfo == "amd64" {
		ok = true
	}
	if osInfo == gutils.Linux {
		switch archInfo {
		case "amd64", "arm64":
			ok = true
		default:
			ok = false
		}
	}
	if osInfo == gutils.Darwin {
		ok = false
	}
	if ok {
		dList, _ := os.ReadDir(binDir)
		if len(dList) == 1 {
			bName := dList[0].Name()
			binPath := filepath.Join(binDir, bName)
			packedBinPath := filepath.Join(binDir, fmt.Sprintf("packed_%s", bName))
			_, err := gutils.ExecuteSysCommand(true, binDir, "upx", "-o", packedBinPath, binPath)
			if err != nil {
				gprint.PrintError("Failed to pack binary: %+v", err)
				os.RemoveAll(packedBinPath)
				return
			}

			os.RemoveAll(binPath)
			os.Rename(packedBinPath, binPath)
		}
	}
}

func (g *GoBuilder) zipDir(src, dst, binName string) (err error) {
	fr, err := os.Open(src)
	if err != nil {
		return
	}
	defer fr.Close()

	info, err := fr.Stat()
	if err != nil || info.IsDir() {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	fw, err := os.Create(dst)
	if err != nil {
		return
	}
	defer fw.Close()
	header.Name = binName
	header.Method = zip.Deflate
	zw := zip.NewWriter(fw)
	writer, err := zw.CreateHeader(header)
	if err != nil {
		return
	}
	defer zw.Close()

	if _, err = io.Copy(writer, fr); err != nil {
		return
	}
	return nil
}

func (g *GoBuilder) Zip(binDir string) {
	dList, _ := os.ReadDir(binDir)
	if len(dList) == 1 {
		binName := dList[0].Name()
		binPath := filepath.Join(binDir, binName)
		zipPath := filepath.Join(filepath.Dir(binDir), fmt.Sprintf("%s.zip", filepath.Base(binDir)))
		g.zipDir(binPath, zipPath, binName)
	}
}

func (g *GoBuilder) Build() {
	if len(g.ArchOSList) == 0 {
		return
	}
	for _, osArch := range g.ArchOSList {
		sList := strings.Split(osArch, "/")
		g.build(sList[0], sList[1])
	}
}

func (g *GoBuilder) prepareArgs(osInfo, archInfo string) ([]string, string) {
	inputArgs := g.BuildArgs

	if len(inputArgs) == 0 {
		inputArgs = append(inputArgs, g.WorkDir)
	}

	// main func position.
	lastArg := inputArgs[len(inputArgs)-1]

	if !strings.Contains(lastArg, string([]rune{filepath.Separator})) && !strings.HasPrefix(lastArg, ".") {
		inputArgs = append(inputArgs, g.WorkDir)
	} else if lastArg == "." && g.WorkDir != "" {
		inputArgs[len(inputArgs)-1] = g.WorkDir
	} else if lastArg == ".." && g.WorkDir != "" {
		inputArgs[len(inputArgs)-1] = filepath.Dir(g.WorkDir)
	}

	var binName string
	for idx, arg := range g.BuildArgs {
		if arg == "-o" && len(g.BuildArgs) > idx+1 {
			inputArgs = append(inputArgs[:idx], inputArgs[idx+2:]...)
			// If binName has been specified.
			binName = filepath.Base(g.BuildArgs[idx+1])
		}
	}

	if binName == "" {
		binName = filepath.Base(lastArg)
	}

	targetDir := filepath.Join(g.ProjectDir(), "build", fmt.Sprintf("%s_%s-%s", binName, osInfo, archInfo))

	target := targetDir
	if binName != "" {
		if osInfo == gutils.Windows && !strings.HasSuffix(binName, ".exe") {
			binName += ".exe"
		}
		target = filepath.Join(targetDir, binName)
	}

	if len(inputArgs) == 1 {
		inputArgs = append([]string{"-o", target}, inputArgs...)
		return inputArgs, targetDir
	}

	maxIndex := len(inputArgs) - 1
	front := inputArgs[:maxIndex]
	back := inputArgs[maxIndex:]
	inputArgs = append(append(front, "-o", target), back...)
	return inputArgs, targetDir
}

func (g *GoBuilder) build(osInfo, archInfo string) {
	inputArgs, binDir := g.prepareArgs(osInfo, archInfo)
	args := append([]string{"go", "build"}, inputArgs...)

	fmt.Println(args)
	fmt.Println(g.WorkDir)

	if _, err := gutils.ExecuteSysCommand(false, g.WorkDir, args...); err != nil {
		gprint.PrintError("Failed to build binaries: %+v", err)
		os.Exit(1)
	}
	g.PackWithUPX(osInfo, archInfo, binDir)
	g.Zip(binDir)
}
