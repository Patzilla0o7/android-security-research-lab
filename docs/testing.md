# 测试

在仓库根目录运行测试套件：

```bash
./tests/run.sh
```

## 当前测试

| 测试 | 覆盖范围 |
|---|---|
| `config_test.sh` | 本地配置模板加载与必填项校验 |
| `toolchain_test.sh` | 工具清单解析、Java 版本解析、apt 与手动安装方式 |
| `bootstrap_test.sh` | Bootstrap 参数解析 |

这些测试不依赖 AOSP 工作区。还可使用以下命令检查 Shell 语法：

```bash
bash -n bin/lab lib/core/*.sh lib/commands/*.sh lib/services/*.sh tests/*.sh
```

行为级验证必须在 Ubuntu 24.04 中执行：

```bash
./tests/run.sh
./bin/lab doctor
./bin/lab bootstrap plan
```
