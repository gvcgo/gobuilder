[中文](https://github.com/gvcgo/gobuilder/blob/main/docs/README_CN.md) | [En](https://github.com/gvcgo/gobuilder)
### 何为gobuilder？

gobuilder是一个用于编译go项目的工具。它功能上与go build类似，但是做了增强。能够同时编译到不同平台和架构，也不需要单独写脚本。

### 功能特点

- 同时编译到go build支持的任何一个或者多个平台；
- 使用**xgo**对CGO进行交叉编译(可选)；
- 使用**UPX**对binary进行压缩(可选)；
- 使用**garble**对windows可执行文件进行混淆(可选)；
- 使用**osslsigncode**对windows可执行文件进行数字签名(可选)；
- 自动对binary进行zip压缩打包(可选)；
- 在go项目下的任何文件夹中，都可以一键编译该项目；
- 记住编译参数，后续任何时间再编译时，无需要输入任何参数；
- 无需编写任何脚本；
- 整洁，会在项目主目录下创建build文件夹，编译配置文件build.json以及二进制文件、压缩文件均存放在此处；

**注意**: 建议使用[VMR](https://github.com/gvcgo/version-manager)安装**upx** 和 **go compiler**。 **osslsigncode** 的安装则需要手动编译。 **garble** 和 **xgo** 可以通过 **go install xxx**来安装。 **xgo docker镜像** 是 **ghcr.io/crazy-max/xgo** 或 **crazymax/xgo**。

windows自签名证书生成方法，详见[这里](https://blog.csdn.net/Think88666/article/details/125947720)。

```bash
New-SelfSignedCertificate -Type Custom -Subject "CN=姓名, O=公司名称, C=CN, L=上海, S=上海" -KeyUsage DigitalSignature -FriendlyName "MailTool" -CertStoreLocation "Cert:\CurrentUser\My" -TextExtension @("2.5.29.37={text}1.3.6.1.5.5.7.3.3", "2.5.29.19={text}") -NotAfter (Get-Date).AddYears(10)
```

### 如何使用？

- 安装

```bash
go install github.com/gvcgo/gobuilder/cmd/gber@v0.1.5
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

- [go compiler](https://go.dev/dl/) (必需)
- [garble](https://github.com/burrowers/garble) (可选)
- [osslsigncode](https://github.com/mtrojnar/osslsigncode) (可选)
- [upx](https://github.com/upx/upx) (可选)
- [xgo](https://github.com/crazy-max/xgo) (可选)

**注意**：xgo的docker镜像，国内可以使用**ghcr.m.daocloud.io/crazy-max/xgo**加速安装。
