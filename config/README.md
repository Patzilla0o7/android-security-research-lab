# Configuration

本目录是 ASRL 的唯一配置来源。所有服务都应通过 `lib/core/config.sh`
加载配置，禁止在 Bash 业务逻辑中写死路径、Java 位置、AOSP 分支或构建 target。

## 配置文件

- `doctor.conf`：受版本控制的 Ubuntu 24.04 环境检查阈值。
- `tools.conf`：受版本控制的工具能力清单，定义检测方式、必需级别和安装方式。
- `lab.conf.example`：受版本控制的本地实验室配置模板。
- `lab.conf`：当前 Ubuntu 主机的实际配置，已被 Git 忽略，不能提交凭据或 token。

首次配置：

```bash
cp config/lab.conf.example config/lab.conf
```

然后按本机实际环境修改 `config/lab.conf`。常用配置项包括：

- `ANDROID_WORKSPACE`
- `ANDROID_MANIFEST_URL`
- `ANDROID_BRANCH`
- `ANDROID_BUILD_TARGET`
- `JAVA_HOME`
- `ADB_PATH`
- `CCACHE_DIR`

`LAB_ROOT` 由 `bin/lab` 自动推导，不能在本地配置中覆盖。服务必须在使用上述配置项前调用
`config_load`，并使用 `config_require` 校验必填项。

工具不能分别维护在 Doctor 与 Bootstrap 中。应在 `tools.conf` 声明一次：

- Doctor 依据清单检测工具是否已安装、版本是否满足要求。
- Bootstrap 复用同一检测结果，自动执行 `apt:<package>` 安装方式。
- `manual:<URL>` 方式仅显示给操作者，不会被自动执行。

当前清单中，ADB 可通过 Ubuntu 的 `adb` 包自动安装；Android Emulator 仅进行
`emulator` 命令检测，并指向其官方手动安装方式。
