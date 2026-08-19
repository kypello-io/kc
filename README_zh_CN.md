# Kypello Client (kc)

[![license](https://img.shields.io/badge/license-AGPL%20V3-blue)](https://github.com/kypello-io/kc/blob/master/LICENSE)

kc是[MinIO mc](https://github.com/minio/mc)的分支，由Kypello维护。
本项目与MinIO, Inc.无任何隶属关系，也未获得其认可或支持。详见[署名与许可证](#署名与许可证)。

Kypello Client (kc)为ls，cat，cp，mirror，diff，find等UNIX命令提供了一种替代方案。它支持文件系统和兼容Amazon S3的云存储服务（AWS Signature v2和v4）。

```
ls        列出文件和文件夹。
mb        创建一个存储桶或一个文件夹。
cat       显示文件和对象内容。
pipe      将一个STDIN重定向到一个对象或者文件或者STDOUT。
share     生成用于共享的URL。
cp        拷贝文件和对象。
mirror    给存储桶和文件夹做镜像。
find      基于参数查找文件。
diff      对两个文件夹或者存储桶比较差异。
rm        删除文件和对象。
events    管理对象通知。
watch     监听文件和对象的事件。
anonymous 管理访问策略。
session   为cp命令管理保存的会话。
config    管理kc配置文件。
version   输出版本信息。
```

## Docker容器

```
docker pull ghcr.io/kypello-io/kc:latest
docker run ghcr.io/kypello-io/kc:latest ls play
```

**注意:** 上述示例默认使用MinIO[演示环境](#验证)做演示，如果想用`kc`操作其它S3兼容的服务，采用下面的方式来启动容器：

```
docker run -it --entrypoint=/bin/sh ghcr.io/kypello-io/kc:latest
```

然后使用[`kc alias`命令](#添加一个云存储服务)。

## GNU/Linux、macOS和FreeBSD

### 下载二进制文件

发布文件名中包含版本号，因此下面的命令会先解析出最新的版本。

```sh
VERSION=$(curl -fsSL https://api.github.com/repos/kypello-io/kc/releases/latest \
  | sed -n 's/.*"tag_name": *"v\([^"]*\)".*/\1/p')
PLATFORM=linux_amd64   # 参见下表

curl -fsSLO "https://github.com/kypello-io/kc/releases/download/v${VERSION}/kc_${VERSION}_${PLATFORM}.tar.gz"
tar xzf "kc_${VERSION}_${PLATFORM}.tar.gz"
./kc --help
```

| 平台 | CPU架构 | `PLATFORM` |
| ---------- | -------- | ------ |
| GNU/Linux | 64-bit Intel | `linux_amd64` |
| GNU/Linux | 64-bit ARM | `linux_arm64` |
| GNU/Linux | 32-bit ARMv7 | `linux_armv7` |
| GNU/Linux | ppc64le | `linux_ppc64le` |
| GNU/Linux | s390x | `linux_s390x` |
| macOS | Intel | `darwin_amd64` |
| macOS | Apple Silicon | `darwin_arm64` |
| FreeBSD | 64-bit Intel | `freebsd_amd64` |

### Linux软件包

deb、rpm和apk软件包提供`amd64`、`arm64`、`armv7`、`ppc64le`和`s390x`架构。

```sh
VERSION=$(curl -fsSL https://api.github.com/repos/kypello-io/kc/releases/latest \
  | sed -n 's/.*"tag_name": *"v\([^"]*\)".*/\1/p')
BASE="https://github.com/kypello-io/kc/releases/download/v${VERSION}"

# Debian/Ubuntu
curl -fsSLO "${BASE}/kc_${VERSION}_linux_amd64.deb" && sudo dpkg -i "kc_${VERSION}_linux_amd64.deb"

# RHEL/Fedora
curl -fsSLO "${BASE}/kc_${VERSION}_linux_amd64.rpm" && sudo rpm -i "kc_${VERSION}_linux_amd64.rpm"

# Alpine
curl -fsSLO "${BASE}/kc_${VERSION}_linux_amd64.apk" && sudo apk add --allow-untrusted "kc_${VERSION}_linux_amd64.apk"
```

## Microsoft Windows
### 下载二进制文件

从[GitHub Releases](https://github.com/kypello-io/kc/releases/latest)下载`kc_<版本号>_windows_amd64.zip`（或`windows_arm64`），解压后运行：

```
kc.exe --help
```

## 通过源码安装
通过源码安装仅适用于开发人员和高级用户。

如果您没有Golang环境，请参照[如何安装Golang](https://golang.org/doc/install)。所需的最低版本见[go.mod](https://github.com/kypello-io/kc/blob/master/go.mod)，当前为go1.26。

```sh
go install github.com/kypello-io/kc@latest
```

## 关于更新

`kc`不会替换自身的二进制文件。请通过安装时使用的软件包管理器或容器镜像仓库进行更新，或按上文重新下载归档文件。`kc update`只会显示当前版本以及获取新版本的地址。

## 添加一个云存储服务
如果你打算仅在POSIX兼容文件系统中使用`kc`,那你可以直接略过本节，跳到[日常使用](#everyday-use)。

添加一个或多个S3兼容的服务，请参考下面说明。`kc`将所有的配置信息都存储在``~/.kc/config.json``文件中。

```
kc alias set <ALIAS> <YOUR-S3-ENDPOINT> <YOUR-ACCESS-KEY> <YOUR-SECRET-KEY> [--api API-SIGNATURE]
```

别名就是给你的云存储服务起了一个短点的外号。S3 endpoint,access key和secret key是你的云存储服务提供的。API签名是可选参数，默认情况下，它被设置为"S3v4"。

### 示例-MinIO云存储
从MinIO服务获得URL、access key和secret key。

```
kc alias set minio http://192.168.1.51 BKIKJAA5BMMU2RHO6IBB V7f1CwQqAcwo80UEIJEjc5gVQUSSx5ohQ9GSrr12 --api s3v4
```

### 示例-Amazon S3云存储
参考[AWS Credentials指南](http://docs.aws.amazon.com/general/latest/gr/aws-security-credentials.html)获取你的AccessKeyID和SecretAccessKey。

```
kc alias set s3 https://s3.amazonaws.com BKIKJAA5BMMU2RHO6IBB V7f1CwQqAcwo80UEIJEjc5gVQUSSx5ohQ9GSrr12 --api s3v4
```

### 示例-Google云存储
参考[Google Credentials Guide](https://cloud.google.com/storage/docs/migrating?hl=en#keys)获取你的AccessKeyID和SecretAccessKey。

```
kc alias set gcs  https://storage.googleapis.com BKIKJAA5BMMU2RHO6IBB V8f1CwQqAcwo80UEIJEjc5gVQUSSx5ohQ9GSrr12 --api s3v2
```

注意：Google云存储只支持旧版签名版本V2，所以你需要选择S3v2。

## 验证
`kc`预先配置了云存储服务URL：https://play.min.io，别名"play"。它是一个用于研发和测试的MinIO服务。如果想测试Amazon S3,你可以将"play"替换为"s3"。

*示例:*

列出https://play.min.io上的所有存储桶。

```
kc ls play
[2016-03-22 19:47:48 PDT]     0B my-bucketname/
[2016-03-22 22:01:07 PDT]     0B mytestbucket/
[2016-03-22 20:04:39 PDT]     0B mybucketname/
[2016-01-28 17:23:11 PST]     0B newbucket/
[2016-03-20 09:08:36 PDT]     0B s3git-test/
```
<a name="everyday-use"></a>
## 日常使用

### Shell别名
你可以添加shell别名来覆盖默认的Unix工具命令。

```
alias ls='kc ls'
alias cp='kc cp'
alias cat='kc cat'
alias mkdir='kc mb'
alias pipe='kc pipe'
alias find='kc find'
```

### Shell自动补全
如果你使用bash、zsh或fish，Shell补全已默认嵌入`kc`中，使用`kc --autocompletion`安装自动补全。重启shell后，kc将自动补全命令，如下所示。

```
kc <TAB>
admin    config   diff     ls       mirror   policy   session  version  watch
cat      cp       events   mb       pipe     rm       share
```

## 贡献
请遵守[贡献者指南](https://github.com/kypello-io/kc/blob/master/CONTRIBUTING.md)。

## 署名与许可证

kc是[MinIO Client (mc)](https://github.com/minio/mc)的分支，mc的版权归MinIO, Inc.所有，并以GNU AGPLv3许可证发布。本分支由Kypello维护。MinIO, Inc.不维护、不认可、也不支持kc，购买MinIO商业订阅不会获得对kc的任何权利。

`kc`的使用受GNU AGPLv3许可证约束，详见[LICENSE](https://github.com/kypello-io/kc/blob/master/LICENSE)文件。按照该许可证第4条的要求，上游的版权声明保留在[NOTICE](https://github.com/kypello-io/kc/blob/master/NOTICE)文件以及各源码文件中。

`kc admin`、`kc support`和`kc idp`等命令用于管理MinIO服务器，其帮助信息中出现的MinIO指的是被管理的服务器，而非kc的作者。
