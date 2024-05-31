package internal

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/gvcgo/goutils/pkgs/gtea/confirm"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gtea/selector"
	"github.com/gvcgo/goutils/pkgs/gutils"
)

var projectDir string

const (
	winSuffix string = ".exe"
)

type GoBuilder struct {
	ArchOSList         []string `json:"arch_os_list"`
	EnableCompress     bool     `json:"compress"`
	EnableUPX          bool     `json:"upx"`
	EnableGarble       bool     `json:"garble"`
	EnableOsslsigncode bool     `json:"osslsigncode"`
	PfxFilePath        string   `json:"pfx_file_path"`
	PfxPassword        string   `json:"pfx_password"`
	PfxCompany         string   `json:"pfx_company"`
	PfxWebsite         string   `json:"pfx_website"`
	BuildArgs          []string `json:"build_args"`
	WorkDir            string   `json:"work_dir"`
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

	// Enable garble for windows.
	cfm = confirm.NewConfirm(confirm.WithTitle("Use garble to obfuscate for windows binary or not?"))
	cfm.Run()
	g.EnableGarble = cfm.Result()

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
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	if len(args) == 0 {
		return
	}

	for idx, v := range args {
		v = strings.ReplaceAll(v, "#", "$")
		args[idx] = v
	}

	g.BuildArgs = args
}

func (g *GoBuilder) PackWithUPX(osInfo, archInfo, binDir, bName string) {
	// UPX cannot pack binaries for MacOS. Segment fault occurrs.
	if !g.EnableUPX || osInfo == gutils.Darwin || (osInfo == gutils.Windows && archInfo != "amd64") {
		return
	}

	if !IsUPXInstalled() {
		gprint.PrintWarning("upx if not found!")
		return
	}

	fmt.Println(gprint.YellowStr("Packing with UPX..."))
	if g.EnableUPX {
		binPath := filepath.Join(binDir, bName)
		packedBinPath := filepath.Join(binDir, fmt.Sprintf("packed_%s", bName))

		_, err := gutils.ExecuteSysCommand(true, binDir, "upx", "-9", "-o", packedBinPath, binPath)
		if err != nil {
			gprint.PrintError("Failed to pack binary: %+v", err)
			os.RemoveAll(packedBinPath)
			return
		}
		os.RemoveAll(binPath)
		os.Rename(packedBinPath, binPath)
	}
}

func (g *GoBuilder) SignWithOsslsigncode(osInfo, archInfo, binDir, binName string) {
	// Only sign windows binaries.
	if osInfo != gutils.Windows {
		return
	}
	if !IsOsslsigncodeInstalled() {
		return
	}
	if ok, _ := gutils.PathIsExist(g.PfxFilePath); !ok || g.PfxPassword == "" {
		return
	}

	gprint.PrintInfo("Signing with osslsigncode...")
	binPath := filepath.Join(binDir, binName)
	signedBinPath := filepath.Join(binDir, fmt.Sprintf("signed_%s", binName))

	/*
		osslsigncode sign -addUnauthenticatedBlob -pkcs12
		/home/moqsien/golang/src/gvcgo/version-manager/scripts/vmr.pfx
		-pass Vmr2024 -n "GVC" -i https://github.com/gvcgo/ -in vmr.exe -out vmr_signed.exe
	*/
	_, err := gutils.ExecuteSysCommand(
		true,
		binDir,
		"osslsigncode",
		"sign",
		"-addUnauthenticatedBlob",
		"-pkcs12",
		g.PfxFilePath,
		"-pass",
		g.PfxPassword,
		"-n",
		g.PfxCompany,
		"-i",
		g.PfxWebsite,
		"-in",
		binPath,
		"-out",
		signedBinPath,
	)
	if err != nil {
		gprint.PrintError("Failed to sign binary: %+v", err)
		os.RemoveAll(signedBinPath)
		return
	}
	os.RemoveAll(binPath)
	os.Rename(signedBinPath, binPath)
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

func (g *GoBuilder) Zip(binDir, osInfo, archInfo, binName string) {
	if !g.EnableCompress {
		return
	}
	fmt.Println(gprint.YellowStr("Zipping binaries..."))

	binPath := filepath.Join(binDir, binName)
	dirPrefix := strings.Split(binName, ".")[0]
	zipPath := filepath.Join(filepath.Dir(binDir), fmt.Sprintf("%s_%s-%s.zip", dirPrefix, osInfo, archInfo))
	g.zipDir(binPath, zipPath, binName)
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

func (g *GoBuilder) clearArgs(args []string) {
	cmdReg := regexp.MustCompile(`(\$\([\w\W]+?\))`)
	for idx, v := range args {
		l := cmdReg.FindAllString(v, -1)
		for _, cc := range l {
			ccc := strings.TrimLeft(cc, "$(")
			ccc = strings.TrimRight(ccc, ")")
			bf, _ := gutils.ExecuteSysCommand(true, g.WorkDir, strings.Split(ccc, " ")...)
			v = strings.ReplaceAll(v, cc, bf.String())
		}
		if len(l) > 0 {
			args[idx] = v
		}
	}
}

func (g *GoBuilder) prepareArgs(osInfo, archInfo string) (args []string, targetDir, binName string) {
	inputArgs := append([]string{}, g.BuildArgs...) // deepcopy

	if len(inputArgs) == 0 {
		inputArgs = append(inputArgs, g.WorkDir)
	}

	// main func position.
	lastArg := inputArgs[len(inputArgs)-1]

	if !strings.HasPrefix(lastArg, string([]rune{filepath.Separator})) && !strings.HasPrefix(lastArg, ".") {
		inputArgs = append(inputArgs, g.WorkDir)
		lastArg = g.WorkDir
	} else if lastArg == "." && g.WorkDir != "" {
		inputArgs[len(inputArgs)-1] = g.WorkDir
		lastArg = g.WorkDir
	} else if lastArg == ".." && g.WorkDir != "" {
		inputArgs[len(inputArgs)-1] = filepath.Dir(g.WorkDir)
		lastArg = g.WorkDir
	}

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

	targetDir = filepath.Join(g.ProjectDir(), "build", fmt.Sprintf("%s-%s", osInfo, archInfo))
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

func (g *GoBuilder) build(osInfo, archInfo string) {
	gprint.PrintInfo("Building for %s/%s...", osInfo, archInfo)
	inputArgs, binDir, binName := g.prepareArgs(osInfo, archInfo)

	compiler := []string{"go", "build"}
	if g.EnableGarble && osInfo == gutils.Windows {
		// use garble to obfuscate go binary.
		CheckAndInstallGarble()
		compiler = []string{"garble", "-literals", "-tiny", "-seed=random", "build"}
	}

	args := append(compiler, inputArgs...)
	g.clearArgs(args)

	os.Setenv("GOOS", osInfo)
	os.Setenv("GOARCH", archInfo)
	if _, err := gutils.ExecuteSysCommand(false, g.WorkDir, args...); err != nil {
		gprint.PrintError("Failed to build binaries: %+v", err)
		os.Exit(1)
	}
	g.PackWithUPX(osInfo, archInfo, binDir, binName)
	if g.EnableOsslsigncode {
		g.SignWithOsslsigncode(osInfo, archInfo, binDir, binName)
	}
	g.Zip(binDir, osInfo, archInfo, binName)
}
