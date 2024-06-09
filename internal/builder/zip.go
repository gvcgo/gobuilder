package builder

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
)

func (b *Builder) zipDir(src, dst, binName string) (err error) {
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

func (b *Builder) Zip(osInfo, archInfo, binDir, binName string) {
	if !b.EnableZip {
		return
	}
	fmt.Println(gprint.YellowStr("Zipping binaries..."))

	binPath := filepath.Join(binDir, binName)
	dirPrefix := strings.Split(binName, ".")[0]
	zipPath := filepath.Join(filepath.Dir(binDir), fmt.Sprintf("%s_%s-%s.zip", dirPrefix, osInfo, archInfo))
	b.zipDir(binPath, zipPath, binName)
}
