# CLI 命令入口

`bin/lab` 是 ASRL 的命令入口。

它只负责三项工作：
1. 定位并导出 `ASRL_ROOT`。
2. 优先运行已经构建的 `build/lab`。
3. 二进制不存在时调用 `scripts/build-go.sh` 完成首次构建。

业务逻辑位于 `internal/`，不能放在本目录。详见 [CLI 文档](../docs/cli.md)。
