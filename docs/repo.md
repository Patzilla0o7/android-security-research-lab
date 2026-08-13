# Repo 工作流

Repo 命令默认使用活动 Workspace；`--workspace <name>` 可以操作指定档案而不切换活动项。

## 状态

```bash
./bin/lab repo status
./bin/lab repo status --workspace android-14
```

状态命令只读，显示 Workspace、manifest URL、branch、Repo 工具位置、`.repo` 初始化状态和 manifest 文件位置。

## 初始化

```bash
./bin/lab repo init --workspace android-14
```

该命令在档案的源码目录执行 `repo init -u <ANDROID_MANIFEST_URL> -b <ANDROID_BRANCH>`。输出同时显示在终端并保存到 `output/repo/<workspace>/<timestamp>-init.log`。

## 同步

默认只显示计划：

```bash
./bin/lab repo sync --workspace android-14 --jobs 8
```

显式应用后才执行网络同步：

```bash
./bin/lab repo sync --workspace android-14 --jobs 8 --apply
```

未指定 `--jobs` 时使用主机逻辑 CPU 数。实际命令为 `repo sync -c -j <jobs>`，日志保存到 `output/repo/<workspace>/<timestamp>-sync.log`。同步前必须已存在 `.repo`。
