# 测试

在仓库根目录运行：

```bash
./tests/run.sh
```

测试入口执行：

1. `go test ./...`，覆盖 CLI、配置、路径安全、工具链、Bootstrap 和多 Workspace。
2. `tests/cli_test.sh`，验证 `bin/lab` 二进制和进程退出码。

其他验证命令：

```bash
go vet ./...
./scripts/build-go.sh
bash -n scripts/*.sh tests/*.sh
git diff --check
```

单元测试和 CLI 冒烟测试不依赖 AOSP 源码。Doctor 和 apt 的主机级验证，以及
Repo 下载、AOSP 构建的端到端验证，应在 Ubuntu 24.04 的专用环境中执行。

GitHub Actions 负责运行不依赖真实 AOSP Workspace 的 Go、Shell 和 CLI 检查。
真实 Repo 下载和 AOSP 编译不在公共 CI 中执行。
