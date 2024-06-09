package builder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/gvcgo/gobuilder/internal/utils"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
)

func IsGoCompilerInstalled() bool {
	_, err := gutils.ExecuteSysCommand(true, "", "go", "version")
	return err == nil
}

const (
	ConfFileName = "build.json"
)

type Builder struct {
	WorkDir            string   `json:"work_dir"`
	ArchOSList         []string `json:"arch_os_list"`
	BuildArgs          []string `json:"build_args"`
	EnableCGO          bool     `json:"enable_cgo"`
	EnableZip          bool     `json:"enable_zip"`
	EnableGarble       bool     `json:"enable_garble"`
	EnableUPX          bool     `json:"enable_upx"`
	EnableOsslsigncode bool     `json:"enable_osslsigncode"`
	OsslPfxFilePath    string   `json:"ossl_pfx_file_path"`
	OsslPfxPassword    string   `json:"ossl_pfx_password"`
	OsslPfxCompany     string   `json:"ossl_pfx_company"`
	OsslPfxWebsite     string   `json:"ossl_pfx_website"`
}

func NewBuilder() (b *Builder) {
	b = &Builder{
		ArchOSList: []string{},
		BuildArgs:  []string{},
	}
	b.LoadConf()
	return b
}

func (b *Builder) ProjectDir() string {
	var projectDir string
	cwd, _ := os.Getwd()
	if cwd != "" {
		projectDir = utils.FindGoProjectDir(cwd)
	}
	return projectDir
}

func (b *Builder) LoadConf() {
	projectDir := b.ProjectDir()
	if projectDir == "" {
		gprint.PrintError("not found go project")
		os.Exit(1)
	}
	buildConf := filepath.Join(projectDir, "build", ConfFileName)
	if ok, _ := gutils.PathIsExist(buildConf); ok {
		data, _ := os.ReadFile(buildConf)
		if err := json.Unmarshal(data, b); err != nil {
			gprint.PrintError("Failed to load build config file: %+v", err)
			os.Exit(1)
		}
	} else {
		b.saveBuilder(buildConf)
	}
}

func (b *Builder) build(osInfo, archInfo string) {
	gprint.PrintInfo("Building for %s/%s...", osInfo, archInfo)
	inputArgs, binDir, binName := b.PrepareArgs(osInfo, archInfo)

	compiler := []string{
		"go",
		"build",
	}

	if b.EnableGarble {
		// Enable garble
		compiler = []string{"garble", "-literals", "-tiny", "-seed=random", "build"}
	}

	args := append(compiler, inputArgs...)
	b.handleInjections(args)

	os.Setenv("GOOS", osInfo)
	os.Setenv("GOARCH", archInfo)
	os.Setenv("CGO_ENABLED", "0") // disable CGO by default.

	// CGO
	if b.EnableCGO {
		if !SupportedArchOSForCGO(osInfo, archInfo) {
			gprint.PrintError("CGO is not supported for %s/%s", osInfo, archInfo)
			return
		}
		b.SetZigForCGO(osInfo, archInfo, binDir, binName)
		args = b.modifyBuildArgsForCGO(args)
	}

	if _, err := gutils.ExecuteSysCommand(false, b.WorkDir, args...); err != nil {
		gprint.PrintError("Failed to build binaries: %+v", err)
		os.Exit(1)
	}

	// UPX
	b.PackWithUPX(osInfo, archInfo, binDir, binName)

	// Osslsigncode
	b.SignWithOsslsigncode(osInfo, archInfo, binDir, binName)

	// Zip
	b.Zip(osInfo, archInfo, binDir, binName)
}

func (b *Builder) Build() {
	if !IsGoCompilerInstalled() {
		gprint.PrintError("go compiler is not installed.")
		return
	}

	if len(b.ArchOSList) == 0 {
		return
	}
	for _, osArch := range b.ArchOSList {
		sList := strings.Split(osArch, "/")
		b.build(sList[0], sList[1])
	}
}
