package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
)

func IsUPXInstalled() bool {
	_, err := gutils.ExecuteSysCommand(true, "", "upx", "--version")
	return err == nil
}

func (b *Builder) PackWithUPX(osInfo, archInfo, binDir, binName string) {
	if !b.EnableUPX {
		return
	}

	if !IsUPXInstalled() {
		gprint.PrintWarning("upx is not found!")
		return
	}

	// UPX cannot pack binaries for MacOS. Segment fault occurrs.
	if osInfo == gutils.Darwin || (osInfo == gutils.Windows && archInfo != "amd64") {
		gprint.PrintWarning("pack with UPX is not supported for %s/%s", osInfo, archInfo)
		return
	}

	fmt.Println(gprint.YellowStr("Packing with UPX..."))

	binPath := filepath.Join(binDir, binName)
	packedBinPath := filepath.Join(binDir, fmt.Sprintf("packed_%s", binName))

	_, err := gutils.ExecuteSysCommand(true, binDir, "upx", "-9", "-o", packedBinPath, binPath)
	if err != nil {
		gprint.PrintError("Failed to pack binary: %+v", err)
		os.RemoveAll(packedBinPath)
		return
	}
	os.RemoveAll(binPath)
	os.Rename(packedBinPath, binPath)
}
