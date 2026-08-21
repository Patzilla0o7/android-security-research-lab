# 项目架构

ASRL 的目标平台为 Ubuntu 24.04。Go 负责 CLI、配置、安全校验、状态管理和
工作流编排；Shell 只保留必须与系统包管理器或 AOSP Shell 环境交互的适配。

## 命令调用链

```text
cmd/lab + internal/*
  -> go build
     -> bin/lab
        -> scripts/bootstrap-apt.sh (仅在显式 --apply 时)
        -> scripts/aosp-build.sh (加载 envsetup、执行 lunch 和 m)
        -> adb (Device 状态与启动等待)

tools/DroidForge (Git submodule)
  -> 独立管理 SDK、AVD、Emulator 和自定义 AOSP 镜像
  -> 不由 lab 运行时直接调用
```

| 层级 | 职责 |
|---|---|
| `bin/lab` | Go 编译产物，不受 Git 管理 |
| `cmd/lab` | Go 程序入口 |
| `internal/` | CLI、配置、安全、工具链与领域工作流 |
| `scripts/build-go.sh` | 构建本机 Go CLI |
| `scripts/bootstrap-apt.sh` | 最小 apt 特权适配器 |
| `scripts/aosp-build.sh` | AOSP envsetup、lunch 和 m 的最小 Shell 适配器 |
| `scripts/build-all.sh` | 初始化 DroidForge submodule 并构建两个独立 CLI |
| `tools/DroidForge` | 独立的模拟设备工具链 Git submodule |
| `config/` | 工具清单、环境阈值与本机 Workspace 档案 |
| `tests/` | Go 测试统一入口和 CLI 二进制冒烟测试 |

## Shell 保留边界

当前保留 Shell 用于本机/联合 Go 构建、测试入口、apt 和 AOSP 构建环境适配。
`scripts/aosp-build.sh` 负责加载 `build/envsetup.sh` 并执行 `lunch`、`m`；
参数解析、路径保护、日志和结果处理由 Go 负责。

## 数据保留原则

- 长期资产：`research/` 以及未来的 `knowledge/`、`automation/`。
- 本机状态：`config/workspaces/*.conf`、`.local/active-workspace`。
- 可清理资产：构建日志、logcat、bugreport、截图和 tombstone。
- 证据索引：`output/evidence/<workspace>/<case>/<timestamp>/` 下的 manifest 与 SHA-256。
- 可重建资产：AOSP 源码工作区、构建产物和 `bin/lab`。
