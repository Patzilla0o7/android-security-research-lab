# 服务层

服务层保存 ASRL 的业务工作流。命令层负责调用服务，但不应实现具体工作流。

已实现服务：

- `doctor.sh`：检查 Ubuntu 主机、本地配置、共享工具链与 Git 身份。
- `bootstrap.sh`：使用共享工具链结果生成安装计划，或安装缺失的 apt 管理工具。

服务遵循项目生命周期约定：

```text
initialize -> execute -> summary -> cleanup
```

详见 [Doctor](../../docs/doctor.md)、[Bootstrap](../../docs/bootstrap.md) 和 [工具链](../../docs/toolchain.md)。
