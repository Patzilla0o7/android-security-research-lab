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
| `lab bootstrap --apply` | 通过 `sudo` 安装具有 `apt` 安装方式的缺失工具 |

## 已预留命令

以下命令已注册，但目前只会报告模块尚未实现：

| 命令 | 计划职责 |
|---|---|
| `lab workspace` | AOSP 工作区初始化与状态管理 |
| `lab repo` | Repo 同步、状态、分支与补丁管理 |
| `lab build` | AOSP 构建 target、日志与输出采集 |
| `lab research` | CVE/0day 项目创建和研究资产管理 |

## 执行约定

- 未知命令会打印帮助并以非零状态退出。
- `bootstrap plan` 为只读操作。
- `bootstrap --apply` 会执行 `sudo apt-get update` 和 `sudo apt-get install`，因此必须显式指定。
- 所有命令应在 Ubuntu 24.04 中运行，而不是 macOS 编辑主机。
