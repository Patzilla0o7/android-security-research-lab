# 测试

在仓库根目录运行：

```bash
./tests/run.sh
```

测试入口执行：

1. `go test ./...`，覆盖 CLI、配置、路径安全、工具链、Bootstrap 和多 Workspace。
2. `tests/cli_test.sh`，验证 `bin/lab` 启动器和进程退出码。

其他验证命令：

```bash
go vet ./...
./scripts/build-go.sh
bash -n bin/lab scripts/*.sh tests/*.sh
git diff --check
```

测试不依赖 AOSP 源码。Doctor、apt 和未来 Repo/AOSP 行为的主机级验证应在
Ubuntu 24.04 中执行。
