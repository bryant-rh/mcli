# mcli
mcli is a tool for pull container images from imagelist or manifests or charts

这个工具同时支持三种方式解析镜像列表，打包成离线镜像。

1. 从k8s，manifests yaml文件中解析镜像列表打包。
2. 从charts包中解析镜像列表打包。
3. 可以把要打包的镜像写入imageList文件中（一行一个），打包成离线镜像。

可以用于k3s场景中部署一些集群基础设施组件，打包离线镜像包，放置于${K3S_ROOT}/k3s/agent/images 目录中，manifes文件放置于${K3S_ROOT}/k3s/server/manifests目录中，k3s启动时即可自动导入离线镜像。

# Usage

```bash
mcli is a tool for pull container images from imagelist and manifests amd charts

Usage:
  mcli [flags]
  mcli [command]

Available Commands:
  auth        Log in or access credentials
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  pull        Pull remote images by reference and store their contents locally

Flags:
  -h, --help                help for mcli
      --insecure            Allow image references to be fetched without TLS
      --platform platform   Specifies the platform in the form os/arch[/variant][:osversion] (e.g. linux/amd64). (default all)
  -v, --verbose             Enable debug logs
      --version             version for mcli

Use "mcli [command] --help" for more information about a command.
```
## 1. 登录
如果镜像仓库是私有仓库，需要登录认证，可通过如下命令先进行登录
```bash
Login to a registry

Usage:
  mcli auth login [OPTIONS] [SERVER] [flags]

Examples:
  # Log in to reg.example.com
  mcli auth login reg.example.com -u AzureDiamond -p hunter2

Flags:
  -h, --help              help for login
  -p, --password string   Password
      --password-stdin    Take the password from stdin
  -u, --username string   Username

Global Flags:
      --insecure            Allow image references to be fetched without TLS
      --platform platform   Specifies the platform in the form os/arch[/variant][:osversion] (e.g. linux/amd64). (default all)
  -v, --verbose             Enable debug logs
```

```bash
# Log in to reg.example.com
mcli auth login reg.example.com -u AzureDiamond -p hunter2
```

## 2. 打包镜像
```bash
Pull container images from imagelist or manifests or charts

Usage:
  mcli pull IMAGE TARBALL [flags]

Flags:
  -c, --cache_path string   Path to cache image layers
      --format string       Format in which to save images ("tarball", "legacy", or "oci") (default "tarball")
  -h, --help                help for pull

Global Flags:
      --insecure            Allow image references to be fetched without TLS
      --platform platform   Specifies the platform in the form os/arch[/variant][:osversion] (e.g. linux/amd64). (default all)
  -v, --verbose             Enable debug logs
```

```bash
#解析当前目录下的manifests目录，或者charts目录，或者imageList的镜像，进行打包
mcli  pull ./  image.tar --platform linux/amd64
```

# Demo
步骤一：
```bash
#在当前目录创建manifests、charts 目录，imageList文件,拷贝要部署的文件，目录结构如下：
.
├── charts
│   └── helm_kafka
├── imageList
└── manifests
    ├── minio
    ├── nginx-deploy.yaml
    └── pgsql
```
+ charts 目录放置需部署的charts部署文件
+ manifests 目录放置需要部署的yaml文件（k8s支持的任意资源文件，包括crd也支持解析）
+ imagesList 填写镜像列表，一行一个，举例如下:
```bash
#cat imageList
nginx:latest
busybox:latest
```

步骤二：
```bash
#第一个参数为上述目录文件的根目录，第二个参数为打包文件名称，最后会进行gzip压缩，生成tar.gz 格式的镜像包文件

mcli pull ./ image.tar 
```

```bash
#支持--platform 参数来指定需要打包的镜像文件CPU架构，最后打包的文件会带上架构
#如执行以下命令，最终打包的文件未image-linux-amd64.tar.gz

mcli pull ./ image.tar --platform linux/amd64
```

```bash
#支持--format （默认：tarball）选择打包的镜像格式,通常默认格式即足够
#注意经测试发现k3s无法导入oci格式文件，因为oci默认是打包所有架构镜像文件

mci pull ./ image.tar --format legacy --platform linux/amd64
```