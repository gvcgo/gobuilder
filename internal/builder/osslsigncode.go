package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/gutils"
)

func IsOsslsigncodeInstalled() (ok bool) {
	binName := "osslsigncode"
	if runtime.GOOS == gutils.Windows {
		binName += ".exe"
	}
	_, err := gutils.ExecuteSysCommand(true, "", binName, "--version")
	return err == nil
}

func (b *Builder) SignWithOsslsigncode(osInfo, archInfo, binDir, binName string) {
	if !b.EnableOsslsigncode {
		return
	}

	// Only sign windows binaries.
	if osInfo != gutils.Windows {
		gprint.PrintWarning("only sign windows binaries.")
		return
	}

	if !IsOsslsigncodeInstalled() {
		return
	}
	if ok, _ := gutils.PathIsExist(b.OsslPfxFilePath); !ok || b.OsslPfxPassword == "" {
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
		b.OsslPfxFilePath,
		"-pass",
		b.OsslPfxPassword,
		"-n",
		b.OsslPfxCompany,
		"-i",
		b.OsslPfxWebsite,
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
