# Doctor 环境检查

`lab doctor` 用于验证 Ubuntu 主机是否满足当前已实现的 ASRL 工作流要求。

```bash
./bin/lab doctor
```

## 检查项

| 范围 | 配置来源 | 结果级别 |
|---|---|---|
| Ubuntu 版本、CPU、内存、磁盘 | `config/doctor.conf` | 根据阈值给出失败或警告 |
| 本地实验室配置 | `config/lab.conf` | 缺失时警告 |
| AOSP 构建依赖与主机工具 | `config/tools.conf` | 必需工具缺失为失败；推荐工具缺失为警告 |
| Git 身份 | 本地 Git 配置 | 缺失时警告 |

工具链检查包括构建依赖、Git、Python 3、Java 17、repo、ccache、ADB 和 Android Emulator。完整工具定义与安装方式见 [toolchain.md](toolchain.md)。

## 结果解释

- **通过：** 要求已满足。
- **警告：** 推荐工具或可选的本地配置缺失。
- **失败：** 必需主机能力或工具缺失、不受支持，或低于最低版本要求。

执行 `lab bootstrap --apply` 后，应始终再次运行 `lab doctor`。Bootstrap 用于执行安装，Doctor 是最终环境验证报告。
