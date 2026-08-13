# Tests

用于测试整个实验室。

包括：

CLI

Bootstrap

Workspace

Automation

以后：

所有新增功能建议增加对应测试。

## 运行

```bash
./tests/run.sh
```

`run.sh` 执行全部 Go 单元测试和一个 Shell CLI 启动器冒烟测试。测试覆盖配置、路径安全、工具链、Bootstrap、CLI 和多 Workspace 工作流，不依赖 AOSP 源码。
