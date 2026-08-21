# 命令行接口

统一入口为 `./bin/lab`。不传子命令时会显示帮助信息。

## 已实现命令

| 命令 | 说明 |
|---|---|
| `lab help` | 显示命令列表 |
| `lab version` | 从 `VERSION` 输出项目名称和版本 |
| `lab doctor [--json]` | 检查 Ubuntu 主机硬件、本地配置、工具链与 Git 身份；硬失败返回非零 |
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
| `lab repo init [options]` | 预览或执行 Repo 初始化，支持 Partial Clone |
| `lab repo sync [options]` | 支持项目、重试和实时进度；`--apply` 后执行并归档日志 |
| `lab repo branch list|create` | 列出或显式创建跨项目研究分支 |
| `lab repo patch export|import` | 导出研究修改或检查、导入 Patch |
| `lab build [plan] [options]` | 预览或执行完整、模块和增量 AOSP 构建 |
| `lab build status` | 查看最近一次构建结果、耗时和日志 |
| `lab device list` | 列出 ADB 设备及其连接状态 |
| `lab device status` | 安全选择设备并读取系统、构建和 SELinux 状态 |
| `lab device wait` | 在超时限制内等待设备连接并完成 Android 启动 |
| `lab collect device-info` | 采集设备与构建属性、SELinux 状态 |
| `lab collect logcat` | 保存当前 logcat 缓冲区 |
| `lab collect screenshot` | 通过 ADB 保存并验证 PNG 截图 |
| `lab collect bugreport` | 生成 bugreport ZIP |
| `lab collect tombstones` | 将可读取的 tombstone 打包为 TAR |
| `lab collect bundle` | 尝试所有采集项并生成 manifest 和 SHA-256 bundle |

`lab --help` 和 `lab --version` 分别等价于 `lab help` 和 `lab version`。

## 已预留命令

以下命令已注册，但目前只会报告模块尚未实现：

| 命令 | 计划职责 |
|---|---|
| `lab research` | CVE/0day 项目创建和研究资产管理 |

## 执行约定

- 未知命令会打印帮助并以非零状态退出。
- 参数错误使用退出码 `2`；尚未实现的占位命令使用退出码 `3`。
- `bootstrap plan` 为只读操作。
- `bootstrap --apply` 会执行 `sudo apt-get update` 和 `sudo apt-get install`，因此必须显式指定。
- 所有命令应在 Ubuntu 24.04 中运行，而不是 macOS 编辑主机。
