# CLI 命令入口

`bin/lab` 是 ASRL 的命令入口。

它只负责三项工作：

1. 定位并导出 `LAB_ROOT`。
2. 从 `lib/core/init.sh` 加载共享核心运行时。
3. 将 `lab <command>` 分发到 `lib/commands/<command>.sh`。

业务逻辑必须放在 `lib/services/`，不能放在本目录。详见 [CLI 文档](../docs/cli.md)。
