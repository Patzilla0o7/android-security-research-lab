# CLI 命令入口

`bin/lab` 是由 Go 编译生成的 ASRL CLI，不受 Git 管理。

```bash
./scripts/build-go.sh
./bin/lab --help
```

Go 源码入口位于 `cmd/lab`，业务逻辑位于 `internal/`。
