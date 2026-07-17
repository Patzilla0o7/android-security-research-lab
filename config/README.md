# Configuration

本目录是 ASRL 的唯一配置来源。所有服务都应通过 `lib/core/config.sh`
加载配置，禁止在 Bash 业务逻辑中写死路径、Java 位置、AOSP 分支或构建 target。

## 配置文件

- `doctor.conf`：受版本控制的 Ubuntu 24.04 环境检查阈值。
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
