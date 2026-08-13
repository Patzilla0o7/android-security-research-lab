# 项目架构

ASRL 的目标平台为 Ubuntu 24.04。Go 负责 CLI、配置、安全校验、状态管理和
工作流编排；Shell 只保留必须与系统包管理器或 AOSP Shell 环境交互的适配。

## 命令调用链

```text
cmd/lab + internal/*
  -> go build
     -> bin/lab
        -> scripts/bootstrap-apt.sh (仅在显式 --apply 时)
```

| 层级 | 职责 |
|---|---|
| `bin/lab` | Go 编译产物，不受 Git 管理 |
| `cmd/lab` | Go 程序入口 |
| `internal/` | CLI、配置、安全、工具链与领域工作流 |
| `scripts/build-go.sh` | 构建本机 Go CLI |
| `scripts/bootstrap-apt.sh` | 最小 apt 特权适配器 |
| `config/` | 工具清单、环境阈值与本机 Workspace 档案 |
| `tests/` | Go 测试统一入口和 CLI 二进制冒烟测试 |

## Shell 保留边界

当前保留 Shell 用于构建、测试入口和 apt。未来 AOSP 构建需要加载
`build/envsetup.sh` 并执行 `lunch`、`m` 时，也应通过小型 Shell 适配器完成；
参数解析、路径保护、日志和结果处理仍由 Go 负责。

## 数据保留原则

- 长期资产：`research/` 以及未来的 `knowledge/`、`automation/`。
- 本机状态：`config/workspaces/*.conf`、`.local/active-workspace`。
- 可清理资产：构建日志、logcat、bugreport、截图和 tombstone。
- 可重建资产：AOSP 源码工作区、构建产物和 `bin/lab`。
