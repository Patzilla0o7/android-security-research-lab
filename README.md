# Android Security Research Lab (ASRL)

ASRL 是面向 Android Framework、安全研究与漏洞分析的长期实验平台。它不是 AOSP 源码镜像；AOSP 工作区可以重新获取，研究记录、知识与自动化才是长期资产。

## 当前可用功能

- 统一命令入口：`bin/lab`
- Ubuntu 24.04 环境检查：`lab doctor`
- 共享工具清单：Doctor 与 Bootstrap 使用同一份 `config/tools.conf`
- 安装计划与 apt 安装：`lab bootstrap plan`、`lab bootstrap --apply`
- 本地实验室配置模板与配置校验
- Bash 语法与配置/工具清单/Bootstrap 参数测试

`workspace`、`repo`、`build`、`research` 命令已预留，但尚未实现业务逻辑。

## 架构

```text
bin/lab
  -> lib/commands/<command>.sh
    -> lib/services/<domain>.sh
      -> lib/core/* + config/*
```

- `bin/`：CLI 分发，不放业务逻辑。
- `lib/commands/`：命令参数与服务调用。
- `lib/services/`：领域工作流，例如 Doctor、Bootstrap。
- `lib/core/`：日志、配置、工具清单解析和通用函数。
- `config/`：唯一配置来源。
- `tests/`：不依赖 AOSP 工作区的自动化测试。

完整说明见 [docs/README.md](docs/README.md)。

## 快速开始

所有命令应在 Ubuntu 24.04 中执行：

```bash
./bin/lab help
./bin/lab doctor
./bin/lab bootstrap plan
```

确认 Bootstrap 计划后，才执行会修改主机的安装操作：

```bash
./bin/lab bootstrap --apply
./bin/lab doctor
```

## 本地配置

```bash
cp config/lab.conf.example config/lab.conf
```

`config/lab.conf` 是机器本地配置，已被 Git 忽略。不要将 token、密码、私钥或研究中的敏感证据提交到仓库。
