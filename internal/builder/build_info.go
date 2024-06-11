package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/gvcgo/gobuilder/internal/utils"
	"github.com/gvcgo/goutils/pkgs/gtea/confirm"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gtea/input"
	"github.com/gvcgo/goutils/pkgs/gtea/selector"
	"github.com/gvcgo/goutils/pkgs/gutils"
)

func (b *Builder) saveBuilder(buildConfPath string) {
	buildDir := filepath.Dir(buildConfPath)
	if ok, _ := gutils.PathIsExist(buildDir); !ok {
		os.MkdirAll(buildDir, os.ModePerm)
	}

	// Choose target Os/Arch
	b.chooseArchOs()

	if len(b.ArchOSList) == 0 {
		gprint.PrintError("No target Os/Arch chosen.")
		os.Exit(1)
	}

	// Enable CGO or not.
	b.enableCGO()

	// Enable zip after compilation.
	b.enableZip()

	// Enable upx or not.
	b.enableUpx()

	// Enable garble or not.
	b.enableGarble()

	// Enable osslsigncode or not.
	b.enableOsslsigncode()

	// process build args.
	b.processArgs()

	// process work dir.
	b.processWorkDir()

	// save conf file.
	b.saveConf(buildConfPath)
}

func (b *Builder) chooseArchOs() {
	items := selector.NewItemList()
	items.Add("Current Os/Arch Only", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
	items.Add("Frequently Used Os/Arch", "frequently")
	others := utils.GetOtherArchOS()
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
		b.ArchOSList = []string{}
	}
	for _, osArch := range targetOsArch {
		if oa, ok := osArch.(string); ok {
			if oa == "frequently" {
				b.ArchOSList = append(b.ArchOSList, utils.GetCommanlyUsedArchOS()...)
			} else {
				b.ArchOSList = append(b.ArchOSList, oa)
			}
		}
	}
}

func (b *Builder) enableCGO() {
	cfm := confirm.NewConfirmation(confirm.WithPrompt("To enable CGO or not?"))
	cfm.Run()
	b.EnableCGoWithXGo = cfm.Result()
}

func (b *Builder) enableZip() {
	cfm := confirm.NewConfirmation(confirm.WithPrompt("To zip binaries or not?"))
	cfm.Run()
	b.EnableZip = cfm.Result()
}

func (b *Builder) enableUpx() {
	cfm := confirm.NewConfirmation(confirm.WithPrompt("To pack binaries with UPX or not?"))
	cfm.Run()
	b.EnableUPX = cfm.Result()
}

func (b *Builder) enableGarble() {
	cfm := confirm.NewConfirmation(confirm.WithPrompt("Use garble to obfuscate for the binaries or not?"))
	cfm.Run()
	b.EnableGarble = cfm.Result()
}

func (b *Builder) enableOsslsigncode() {
	cfm := confirm.NewConfirmation(confirm.WithPrompt("Use osslsigncode to sign the binaries or not?"))
	cfm.Run()
	b.EnableOsslsigncode = cfm.Result()

	if b.EnableOsslsigncode {
		mInput := input.NewMultiInput()
		var (
			pfxFilePath string = "Pfx file path"
			pfxPassword string = "Pfx password"
			pfxCompany  string = "Pfx company"
			pfxWebsite  string = "Pfx website"
		)
		mInput.AddOneItem(pfxFilePath, input.MWithWidth(120))
		mInput.AddOneItem(pfxPassword, input.MWithWidth(60), input.MWithEchoMode(textinput.EchoPassword), input.MWithEchoChar("*"))
		mInput.AddOneItem(pfxCompany, input.MWithWidth(60))
		mInput.AddOneItem(pfxWebsite, input.MWithWidth(120))
		mInput.Run()

		result := mInput.Values()
		b.OsslPfxFilePath = result[pfxFilePath]
		b.OsslPfxPassword = result[pfxPassword]
		b.OsslPfxCompany = result[pfxCompany]
		b.OsslPfxWebsite = result[pfxWebsite]
		if b.OsslPfxFilePath == "" {
			b.EnableOsslsigncode = false
		}
	}
}

func (b *Builder) processArgs() {
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

	b.BuildArgs = args
}

func (b *Builder) processWorkDir() {
	cwd, _ := os.Getwd()
	b.WorkDir = cwd
}

func (b *Builder) saveConf(buildConfPath string) {
	data, _ := json.MarshalIndent(b, "", "    ")
	if len(data) > 0 {
		os.WriteFile(buildConfPath, data, os.ModePerm)
	}
}
