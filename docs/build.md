# AOSP Build 工作流

`lab build` 使用 Workspace 档案中的源码路径、默认 target、Java 和 ccache
配置。Go 负责校验、计划、进程、日志与状态；Shell 适配器只加载 AOSP 定义的
Shell 函数。

## 前置检查

Workspace 必须已经完成 Repo 同步，并包含：

```text
.repo/
build/envsetup.sh
```

先查看构建计划：

```bash
./bin/lab build --workspace android-14
```

计划显示源码路径、branch、target、模块、并发数、Java、ccache、构建模式、
可用磁盘和执行模式。默认不会修改 Workspace。

## 完整与增量构建

```bash
./bin/lab build --workspace android-14 --jobs 16 --apply
```

未指定模块时运行当前 target 的 `m -j<jobs>`。首次没有 `out/` 时标记为
`full`；已有 `out/` 时标记为 `incremental`。命令不会自动清理 `out/`，
避免不可逆地删除大型构建产物。

临时覆盖 Workspace target：

```bash
./bin/lab build --workspace android-14 \
  --target aosp_x86_64-userdebug \
  --jobs 16 \
  --apply
```

覆盖只对本次执行生效，不写回 Workspace 档案。

## 模块构建

```bash
./bin/lab build --workspace android-14 \
  --module framework-minus-apex \
  --module services \
  --jobs 16 \
  --apply
```

`--module` 可以重复，模块名称作为独立参数传给 AOSP `m`，不会交给 Shell
重新解析。

## 构建状态与日志

```bash
./bin/lab build status --workspace android-14
```

每次实际执行都会在 `output/build/<workspace>/` 保存：

- `<timestamp>-build.log`：完整 stdout/stderr。
- `<timestamp>-build.json`：本次 target、模块、耗时、退出码和日志路径。
- `latest.json`：最近一次构建结果，供 status 查询。

失败也会写入元数据和退出码。构建日志与 AOSP `out/` 都是可重建资产，不应
作为长期漏洞研究记录；重要结论、Patch 和证据应另外归档。
