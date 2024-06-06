[中文](https://github.com/gvcgo/gobuilder/blob/main/docs/README_CN.md) | [En](https://github.com/gvcgo/gobuilder)
### 何为gobuilder？

gobuilder是一个用于编译go项目的工具。它功能上与go build类似，但是做了增强。能够同时编译到不同平台和架构，也不需要单独写脚本。

### 功能特点

- 同时编译到go build支持的任何一个或者多个平台；
- 使用**UPX**对binary进行压缩(可选)；
- 使用**garble**对windows可执行文件进行混淆(可选)；
- 使用**osslsigncode**对windows可执行文件进行数字签名(可选)；
- 自动对binary进行zip压缩打包(可选)；
- 在go项目下的任何文件夹中，都可以一键编译该项目；
- 记住编译参数，后续任何时间再编译时，无需要输入任何参数；
- 无需编写任何脚本；

**注意**：在使用osslsigncode对windows可执行文件进行数字签名时，需要提前安装**osslsigncode**工具。另外，需要手动在gobuilder的配置文件build/gbuild.json中配置如下字段：

```json
// Example:
{
    "osslsigncode": true,
    "pfx_file_path":"/home/moqsien/golang/src/gvcgo/version-manager/scripts/vmr.pfx",
    "pfx_password":"Vmr2024",
    "pfx_company":"GVC",
    "pfx_website":"https://github.com/gvcgo/",
}
```

### 如何使用？

- 安装

```bash
go install github.com/gvcgo/gobuilder/cmd/gber@v0.1.3
```

- 使用方法

```bash
gber build <your-go-build-flags-and-args>
```

**注意**: 如果你需要在编译时动态地注入一些变量，也许你需要将"$"符号替换为"#"符号，以此避免$()中的命令立即执行，gber会自动识别这类替换。例如：

```bash
# original
gber build -ldflags "-X main.GitTag=$(git describe --abbrev=0 --tags) -X main.GitHash=$(git show -s --format=%H)  -s -w" ./cmd/vmr/

# replaced
gber build -ldflags "-X main.GitTag=#(git describe --abbrev=0 --tags) -X main.GitHash=#(git show -s --format=%H)  -s -w" ./cmd/vmr
```

### 项目依赖

- [garble](https://github.com/burrowers/garble) (可选)
- [osslsigncode](https://github.com/mtrojnar/osslsigncode) (可选)
- [upx](https://github.com/upx/upx) (可选)
