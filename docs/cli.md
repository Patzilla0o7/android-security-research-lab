# 命令行接口

统一入口为 `./bin/lab`。不传子命令时会显示帮助信息。

## 已实现命令

| 命令 | 说明 |
|---|---|
| `lab help` | 显示命令列表 |
| `lab version` | 从 `VERSION` 输出项目名称和版本 |
| `lab doctor` | 检查 Ubuntu 主机硬件、本地配置、工具链与 Git 身份 |
| `lab bootstrap` | 等同于 `lab bootstrap plan`，不会修改系统 |
| `lab bootstrap plan` | 显示缺失工具及可用安装方式 |
| `lab bootstrap --apply` | Go 生成计划，并由最小 Shell 适配器执行 apt |
| `lab workspace list` | 列出档案并标记活动 Workspace |
| `lab workspace add` | 添加独立的 AOSP 版本档案 |
| `lab workspace use` | 切换活动档案，不触发同步或构建 |
| `lab workspace current` | 输出活动档案名称 |
| `lab workspace status [name]` | 查看活动或指定 Workspace 状态 |
| `lab workspace init [name]` | 创建活动或指定 Workspace 目录 |
| `lab repo status [--workspace <name>]` | 查看 Repo 工具与初始化状态 |
| `lab repo init [--workspace <name>]` | 使用档案 manifest 和 branch 初始化 Repo |
| `lab repo sync [options]` | 默认显示计划；`--apply` 后执行同步并归档日志 |

`lab --help` 和 `lab --version` 分别等价于 `lab help` 和 `lab version`。

## 已预留命令

以下命令已注册，但目前只会报告模块尚未实现：

| 命令 | 计划职责 |
|---|---|
| `lab build` | AOSP 构建 target、日志与输出采集 |
| `lab research` | CVE/0day 项目创建和研究资产管理 |

## 执行约定

- 未知命令会打印帮助并以非零状态退出。
- 参数错误使用退出码 `2`；尚未实现的占位命令使用退出码 `3`。
- `bootstrap plan` 为只读操作。
- `bootstrap --apply` 会执行 `sudo apt-get update` 和 `sudo apt-get install`，因此必须显式指定。
- 所有命令应在 Ubuntu 24.04 中运行，而不是 macOS 编辑主机。
