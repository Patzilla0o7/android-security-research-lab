# 项目架构

## 执行环境

ASRL 脚本的目标平台为 Ubuntu 24.04。代码可通过共享目录在 macOS 中编辑，但功能验证、Bootstrap、AOSP 同步和构建必须在 Ubuntu 中执行。

## 命令调用链

```text
bin/lab
  -> lib/core/init.sh
  -> lib/commands/<command>.sh
  -> lib/services/<domain>.sh
  -> lib/core/* and config/*
```

| 层级 | 职责 |
|---|---|
| `bin/lab` | 定位项目根目录、初始化核心模块、分发子命令 |
| `lib/commands/` | 轻量命令适配层；不包含业务流程 |
| `lib/services/` | Doctor、Bootstrap 等业务工作流 |
| `lib/core/` | 常量、日志、配置加载、工具清单解析、通用工具函数 |
| `config/` | 受版本控制的阈值和工具定义，以及被忽略的本地配置 |
| `tests/` | 无需 AOSP 工作区的 Bash 测试 |

## 当前目录结构

```text
bin/                 CLI 命令入口
config/              环境与工具链配置
docs/                项目文档
lib/core/            共享运行时基础能力
lib/commands/        CLI 命令适配层
lib/services/        已实现的业务工作流
output/              可清理的运行输出
research/            未来的长期漏洞研究记录
templates/           未来的研究与报告模板
tests/               自动化测试
tools/               未来的第三方工具资产
```

## 工具链调用关系

`config/tools.conf` 是主机工具与 AOSP 构建依赖的唯一事实来源。

```text
config/tools.conf
       -> lib/core/toolchain.sh
          -> lab doctor     (检测与报告)
          -> lab bootstrap  (生成计划与安装)
```

Bootstrap 不会解析 Doctor 的终端输出。两个命令调用相同的共享检测函数，避免检查规则与安装规则发生漂移。

## 数据保留原则

- 长期资产：未来的 `research/`、`knowledge/` 和 `automation/` 内容。
- 可清理资产：构建日志、logcat、bugreport、截图、tombstone 及其他 `output/` 数据。
- 可重建资产：AOSP 源码工作区与构建产物。
