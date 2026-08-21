# 开发路线图

ASRL 当前已经完成统一 CLI、Ubuntu Doctor、共享工具链、Bootstrap 安装计划、本地配置加载和基础测试。后续开发按安全依赖关系推进。

## 当前：Go 迁移

- 已完成 Go CLI、配置解析、路径安全、工具清单检测和 Doctor。
- Bootstrap 计划已迁移到 Go，`apt-get` 执行层保留 Shell。
- 多 Workspace、Repo 工作流以及 AOSP Build 已实现；下一步进入设备与证据采集。

## 近期：CLI 与配置基础

- 统一全局及子命令帮助、退出码和错误信息。
- 增加绝对路径校验、危险路径保护和非交互约定。
- 扩展 CLI 分发、配置加载与安全辅助函数测试。

## 下一阶段：AOSP Workspace 与 Repo

- 已完成 `lab repo status|init|sync`。
- 已完成 `lab repo branch list|create` 与 `lab repo patch import|export`。
- 所有删除、清理和覆盖操作必须进行路径保护并要求显式确认。

## 已完成：Build

- 已完成构建目标、ccache、模块/增量/完整构建、日志和状态归档。

## 当前：基础加固、设备与证据采集

- 同步实现与文档，建立 Ubuntu 24.04 CI。
- Doctor 硬失败退出码、结构化输出与测试。
- 已完成 ADB 设备发现、状态读取、多设备保护与启动完成等待。
- Emulator 生命周期由 DroidForge submodule 独立管理，ASRL 不直接调用。
- 已完成 device-info、logcat、截图、bugreport、tombstones、完整 bundle、采集元数据和 SHA-256 校验。
- 已完成 Research 项目创建、模板、元数据、校验和证据 bundle 关联。
- 下一步完善证据脱敏和长期归档策略。
- 已按 Workspace、case、UTC 时间组织 `output/evidence/`，失败时保留部分证据。

当前阶段完成标准：CLI 能识别并等待指定设备，启动受管 Emulator，并生成带设备、
Workspace、时间、命令和 SHA-256 信息的证据包。Repo 下载与 AOSP 编译继续在专用
环境中验证，不纳入公共 CI。

## 中长期：Research、Knowledge 与 Automation

- 已完成结构化 CVE/0day 研究项目、基础元数据、根因与复现模板。
- 下一步扩展 PoC、补丁和报告自动化。
- Framework、Binder、权限、SELinux 和内存安全知识分类。
- 可审计、显式启用的研究自动化工作流。

## 长期：质量与发布

- Shell 格式化、静态检查和 CI。
- Workspace 路径保护、模板生成及核心 CLI 回归测试。
- 版本、变更日志和 Ubuntu 24.04 发布验证流程。
