# Configuration

本目录保存 ASRL 的全局配置与本机 Workspace 档案。Go 服务通过 `internal/config` 读取数据配置；保留的 Shell 适配仍使用 `lib/core/config.sh`。

## 配置文件

- `doctor.conf`：受版本控制的 Ubuntu 24.04 环境检查阈值。
- `tools.conf`：受版本控制的工具能力清单，定义检测方式、必需级别和安装方式。
- `lab.conf.example`：受版本控制的本地实验室配置模板。
- `lab.conf`：兼容的本机全局配置，已被 Git 忽略。
- `workspaces/*.conf`：由 `lab workspace add` 创建的本机 AOSP 档案，已被 Git 忽略。
- `.local/active-workspace`：当前活动档案名称，已被 Git 忽略。

首次使用多 Workspace：

```bash
./bin/lab workspace add android-14 --path /data/aosp/android-14 --branch android-14.0.0_r75
./bin/lab workspace add android-15 --path /data/aosp/android-15 --branch android-15.0.0_r1
./bin/lab workspace use android-15
```

档案名称只允许字母、数字、`-` 和 `_`。不同档案不能指向同一源码路径。

工具能力只在 `tools.conf` 声明一次：

- Doctor 依据清单检测工具是否安装及版本是否满足要求。
- Bootstrap 复用同一检测结果，自动执行 `apt:<package>` 安装方式。
- `manual:<URL>` 方式仅显示给操作者，不会被自动执行。

当前清单中，ADB 可通过 Ubuntu 的 `adb` 包自动安装；Android Emulator 仅进行
`emulator` 命令检测，并指向其官方手动安装方式。
