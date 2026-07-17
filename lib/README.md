# 运行时库

`lib/` 保存 ASRL 的 Bash 运行时。

- `core/`：共享常量、日志、配置、工具链检测和服务辅助函数。
- `commands/`：由 `bin/lab` 调用的 CLI 适配层。
- `services/`：已实现的领域工作流。

详见 [项目架构文档](../docs/architecture.md)。
