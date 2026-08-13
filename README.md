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

```bash
./bin/lab workspace add android-14 --path /data/aosp/android-14 --branch android-14.0.0_r75
./bin/lab workspace use android-14
./bin/lab workspace init
```

Workspace 档案与活动选择是机器本地配置，已被 Git 忽略。不要将 token、密码、私钥或研究中的敏感证据提交到仓库。
