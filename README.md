# Android Security Research Lab (ASRL)

ASRL 是面向 Android Framework、安全研究与漏洞分析的长期实验平台。它不是 AOSP 源码镜像；AOSP 工作区可以重新获取，研究记录、知识与自动化才是长期资产。

## 当前可用功能

- 统一命令入口：`bin/lab`
- Go CLI、配置解析、路径保护和工具链检测
- Ubuntu 24.04 环境检查：`lab doctor`
- 共享工具清单：Doctor 与 Bootstrap 使用同一份 `config/tools.conf`
- 安装计划与 apt 安装：`lab bootstrap plan`、`lab bootstrap --apply`
- 本地实验室配置模板与配置校验
- Go 单元测试与 CLI 二进制冒烟测试

多 Workspace 档案管理已实现；`repo`、`build`、`research` 仍为预留命令。

## Go 与 Shell 边界

项目 CLI、Doctor、Bootstrap 计划和 Workspace 已使用 Go。`bootstrap --apply` 以及未来需要加载 AOSP `envsetup.sh`、执行 `lunch`/`m` 的适配继续保留 Shell。

```bash
./scripts/build-go.sh
./tests/run.sh
./bin/lab doctor
```

`bin/lab` 是 `scripts/build-go.sh` 生成的 Go 二进制，不受 Git 管理。

## 架构

```text
cmd/lab + internal/*
  -> go build
    -> bin/lab (Go CLI)
    -> scripts/bootstrap-apt.sh (仅显式安装)
```

- `bin/`：存放本机 Go 编译产物。
- `cmd/` 与 `internal/`：Go CLI、基础能力与领域工作流。
- `scripts/`：Go 构建及必须保留的 Shell 适配。
- `config/`：唯一配置来源。
- `tests/`：不依赖 AOSP 工作区的自动化测试。

完整说明见 [docs/README.md](docs/README.md)。

## 快速开始

所有命令应在 Ubuntu 24.04 中执行：

```bash
./bin/lab help
./bin/lab doctor
./bin/lab bootstrap plan
```

确认 Bootstrap 计划后，才执行会修改主机的安装操作：

```bash
./bin/lab bootstrap --apply
./bin/lab doctor
```

## Workspace 配置

先创建一个独立的 AOSP Workspace 档案：

```bash
./bin/lab workspace add android-14 \
  --path /data/aosp/android-14 \
  --manifest https://android.googlesource.com/platform/manifest \
  --branch android-14.0.0_r75 \
  --target aosp_x86_64-eng \
  --java-home /usr/lib/jvm/java-17-openjdk-amd64

./bin/lab workspace use android-14
./bin/lab workspace init
```

`workspace add` 参数：

| 参数 | 必需 | 默认值 | 说明 |
|---|---|---|---|
| `<name>` | 是 | 无 | 档案名称；只允许字母、数字、`-`、`_` |
| `--path` | 是 | 无 | AOSP 源码目录，必须是安全的绝对路径 |
| `--manifest` | 否 | `https://android.googlesource.com/platform/manifest` | Repo manifest 仓库地址 |
| `--branch` | 否 | `android-latest-release` | Repo manifest 分支或 Android tag |
| `--target` | 否 | `aosp_x86_64-eng` | 未来传给 `lunch` 的默认构建目标 |
| `--java-home` | 否 | 空 | 该 Workspace 使用的 JDK 路径 |

`--target` 当前会保存到档案的 `ANDROID_BUILD_TARGET`，待 Build 工作流实现后用于选择构建目标。`CCACHE_DIR` 无需传入，系统会按档案名称自动生成独立路径。

管理多个版本：

```bash
./bin/lab workspace add android-15 \
  --path /data/aosp/android-15 \
  --branch android-15.0.0_r1 \
  --target aosp_x86_64-eng

./bin/lab workspace list
./bin/lab workspace use android-15
./bin/lab workspace current
./bin/lab workspace status android-14
./bin/lab workspace init android-14
```

### Workspace 档案与全局配置

| 文件 | 用途 |
|---|---|
| `config/workspaces/<name>.conf` | 当前正式的多 Workspace 配置；保存源码路径、manifest、分支、构建目标、Java 和 ccache 路径 |
| `.local/active-workspace` | 保存 `workspace use` 选择的活动档案名称 |
| `config/lab.conf` | 旧单 Workspace/本机全局配置的兼容文件，不用于多 Workspace 的选择和切换 |
| `config/lab.conf.example` | 旧配置格式模板；新增 AOSP 版本应优先使用 `workspace add` |

Workspace 档案与活动选择都是机器本地数据，已被 Git 忽略。`workspace use` 只切换活动配置，不会同步、构建或删除源码。不要将 token、密码、私钥或研究中的敏感证据提交到仓库。
