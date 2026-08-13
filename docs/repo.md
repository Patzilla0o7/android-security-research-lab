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

可选同步参数：

| 参数 | 说明 |
|---|---|
| `--project <path>` | 只同步指定项目；可重复，例如 `frameworks/base` 和 `system/core` |
| `--retry-fetches <count>` | Git fetch 失败后的重试次数 |
| `--no-clone-bundle` | 禁用 clone bundle，直接从远端抓取 |
| `--force-sync` | 必要时覆盖指向不同 object directory 的 Git 目录；可能丢失 refs，谨慎使用 |
| `--jobs <count>` | 并发任务数；默认使用主机逻辑 CPU 数 |
| `--apply` | 实际执行；缺少该参数时只显示计划 |

例如只同步 Framework 和 System Core：

```bash
./bin/lab repo sync --workspace android-14 \
  --project frameworks/base \
  --project system/core \
  --retry-fetches 3 \
  --no-clone-bundle \
  --jobs 4 \
  --apply
```

交互式终端运行时，ASRL 保留 Repo/Git 的终端 stderr，使其显示项目级完成数、
当前 fetch 的对象数、百分比和已接收字节。Repo/Git 在同步前不能准确计算所有
项目的最终下载字节总量，所以不会显示可靠的全局“剩余 GB”。为了保留终端
进度检测，动态 stderr 进度可能不会重复写入日志；普通命令输出仍保存到
`output/repo/<workspace>/<timestamp>-sync.log`。同步前必须已存在 `.repo`。
