# 工具链清单

`config/tools.conf` 是主机工具与 AOSP 构建依赖的唯一清单。

## 配置格式

每个 `TOOL_SPECS` 条目由七个以管道符分隔的字段组成：

```text
id|label|level|check type|check target|min version|install methods
```

示例：

```bash
"java|Java|required|java-major|java|17|apt:openjdk-17-jdk"
"emulator|Android Emulator|recommended|command|emulator||manual:https://developer.android.com/studio/run/emulator-commandline"
```

## 字段说明

| 字段 | 含义 |
|---|---|
| `id` | 稳定的内部标识符 |
| `label` | 面向用户的报告名称 |
| `level` | `required`、`recommended` 或 `optional` |
| `check type` | `command`、`java-major` 或 `package` |
| `check target` | 命令名或 dpkg 软件包名 |
| `min version` | `java-major` 所需的最低 Java 主版本；其他检查类型为空 |
| `install methods` | 以逗号分隔的 `apt:<package>` 和/或 `manual:<URL>` 方法 |

## 工作方式

- `lab doctor` 检查每个条目，并依据 `level` 输出通过、警告或失败。
- `lab bootstrap plan` 使用相同检查结果，按安装方式归类缺失工具。
- `lab bootstrap --apply` 仅自动安装可用的 `apt:<package>` 条目。
- `manual:<URL>` 只会展示给操作者，绝不会自动下载或执行。

## 当前设备调试工具

- **ADB：** 使用 `adb` 命令检测，可通过 Ubuntu 的 `adb` 软件包安装。
- **Android Emulator：** 使用 `emulator` 命令检测。目前提供官方手动安装方式，因为 Ubuntu SDK 软件包不提供当前的 Emulator 命令。

新增工具时，只添加一条工具清单记录；若使用新的检查或安装方法，再补充相应解析测试。不要在 Doctor 或 Bootstrap 中分别硬编码重复检查。
