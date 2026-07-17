# Bootstrap 环境初始化

`lab bootstrap` 用于准备符合要求的 Ubuntu 24.04 AOSP 研究主机。

## 命令

```bash
lab bootstrap
lab bootstrap plan
lab bootstrap --apply
```

- `plan` 为默认模式。它使用与 `lab doctor` 相同的工具链检查，并展示可用安装方式；不会修改主机。
- `--apply` 会运行 `sudo apt-get update`，并安装具有 `apt` 安装方式的缺失工具。只有 `manual` 方式的工具仍会显示给操作者手动处理。

共享的 [工具清单](../config/tools.conf) 基于 AOSP 对 Ubuntu 18.04 及以上版本的构建要求，并加入了 ASRL 所需的 OpenJDK 17、repo、ccache 等工具。安装后运行：

```bash
lab doctor
```

在审阅 `plan` 输出之前，不要运行 `--apply`。
