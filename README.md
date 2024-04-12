[中文](https://github.com/gvcgo/gobuilder/blob/main/docs/README_CN.md) | [En](https://github.com/gvcgo/gobuilder)
### What's gobuilder?

Gobuilder is a tool for building Go binaries. It is similar to the Go tool, but it supports building multiple binaries at once and supports custom build configurations without creating any script.

### Features

- Builds binaries for any platform from go source code.
- Packs binaries with UPX.
- Zip binaries automatically.
- Builds binaries at anywhere in a go project.
- Remembers the build operations forever.
- No script is needed.

### How to use?

- Install

```bash
go install github.com/gvcgo/gobuilder/cmd/gber@v0.0.8
```

- Usage

```bash
gber build <your-go-build-flags-and-args>
```

**Note**: If you need to inject variables when building go source code, "$" should be replaced with "#".
```bash
# original
gber build -ldflags "-X main.GitTag=$(git describe --abbrev=0 --tags) -X main.GitHash=$(git show -s --format=%H)  -s -w" ./cmd/vmr/

# replaced
gber build -ldflags "-X main.GitTag=#(git describe --abbrev=0 --tags) -X main.GitHash=#(git show -s --format=%H)  -s -w" ./cmd/vmr
```
