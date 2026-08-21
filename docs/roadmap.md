# 开发路线图

ASRL 已达到当前使用范围的功能基线，进入稳定维护阶段。项目聚焦 AOSP 环境、设备
连接和漏洞研究项目管理，不内置证据采集、脱敏或归档系统。

## 已完成

- Go CLI、配置解析、路径保护、Doctor 和 Bootstrap。
- 多 AOSP Workspace 添加、切换、状态和目录初始化。
- Repo 初始化、同步、研究分支以及 Patch 导入导出。
- AOSP 完整、增量和模块构建编排、日志与状态。
- DroidForge submodule 和 ASRL/DroidForge 联合构建。
- ADB 设备发现、状态读取、多设备保护与启动等待。
- Research case 创建、列表、查看、schema 校验和标准模板。
- Ubuntu 24.04 CI、Go 单元测试和 CLI 冒烟测试。

## 当前维护原则

- 不主动扩展超出当前研究需求的大型子系统。
- 优先修复真实使用中发现的问题和兼容性回归。
- 保持 CLI、Workspace profile 和 Research case schema 向后兼容。
- Repo 下载和 AOSP 编译继续在专用实验环境中验证，不纳入公共 CI。
- 所有具有修改或覆盖影响的操作继续要求显式 `--apply`。

## 按需演进

- 根据实际研究需求补充 Research 模板或小型辅助命令。
- 定期验证 Ubuntu、Go、Repo、ADB 和 DroidForge 兼容性。
- 在准备正式发布时补充 CHANGELOG、版本策略和发布验收清单。
