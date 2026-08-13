# 开发路线图

ASRL 当前已经完成统一 CLI、Ubuntu Doctor、共享工具链、Bootstrap 安装计划、本地配置加载和基础测试。后续开发按安全依赖关系推进。

## 当前：Go 迁移

- 已完成 Go CLI、配置解析、路径安全、工具清单检测和 Doctor。
- Bootstrap 计划已迁移到 Go，`apt-get` 执行层保留 Shell。
- 多 Workspace 以及 Repo `status|init|sync` 已实现；Repo branch/patch 是下一项。

## 近期：CLI 与配置基础

- 统一全局及子命令帮助、退出码和错误信息。
- 增加绝对路径校验、危险路径保护和非交互约定。
- 扩展 CLI 分发、配置加载与安全辅助函数测试。

## 下一阶段：AOSP Workspace 与 Repo

- 已完成 `lab repo status|init|sync`。
- 下一步实现 `lab repo branch` 与 `lab repo patch import|export`。
- 所有删除、清理和覆盖操作必须进行路径保护并要求显式确认。

## 中期：Build、设备与证据采集

- 构建目标、ccache、增量/完整构建和日志归档。
- Emulator/ADB 健康检查与启动流程。
- logcat、bugreport、tombstone 和截图采集。
- 按时间与操作组织 `output/`，并在失败时给出日志位置。

## 中长期：Research、Knowledge 与 Automation

- 结构化 CVE/0day 研究项目和元数据。
- PoC、根因、补丁、复现步骤及报告模板。
- Framework、Binder、权限、SELinux 和内存安全知识分类。
- 可审计、显式启用的研究自动化工作流。

## 长期：质量与发布

- Shell 格式化、静态检查和 CI。
- Workspace 路径保护、模板生成及核心 CLI 回归测试。
- 版本、变更日志和 Ubuntu 24.04 发布验证流程。
