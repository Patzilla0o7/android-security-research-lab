# Android Security Research Lab (ASRL)

ASRL 是面向 Android Framework、安全研究与漏洞分析的长期实验平台。它不是 AOSP 源码镜像；AOSP 工作区可以重新获取，研究记录、知识与自动化才是长期资产。

## 当前可用功能

- 统一命令入口：`bin/lab`
- Go CLI、配置解析、路径保护和工具链检测
- Ubuntu 24.04 环境检查：`lab doctor`、`lab doctor --json`
- 共享工具清单：Doctor 与 Bootstrap 使用同一份 `config/tools.conf`
- 安装计划与 apt 安装：`lab bootstrap plan`、`lab bootstrap --apply`
- 本地实验室配置模板与配置校验
- Go 单元测试与 CLI 二进制冒烟测试

多 Workspace、Repo `status|init|sync|branch|patch` 和 AOSP Build 已实现；`research` 仍为预留命令。

## Go 与 Shell 边界

项目 CLI、Doctor、Bootstrap 计划、Workspace 和 Repo 工作流已使用 Go。`bootstrap --apply` 以及未来需要加载 AOSP `envsetup.sh`、执行 `lunch`/`m` 的适配继续保留 Shell。

```bash
./scripts/build-go.sh
./tests/run.sh
./bin/lab doctor
```

`bin/lab` 是 `scripts/build-go.sh` 生成的 Go 二进制，不受 Git 管理。

## 架构

```text
cmd/lab + internal/*
  -> go build
    -> bin/lab (Go CLI)
    -> scripts/bootstrap-apt.sh (仅显式安装)
```

- `bin/`：存放本机 Go 编译产物。
- `cmd/` 与 `internal/`：Go CLI、基础能力与领域工作流。
- `scripts/`：Go 构建及必须保留的 Shell 适配。
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

## Workspace 配置

先创建一个独立的 AOSP Workspace 档案：

```bash
./bin/lab workspace add android-14 \
  --path /data/aosp/android-14 \
  --manifest https://android.googlesource.com/platform/manifest \
  --branch android-14.0.0_r75 \
  --target aosp_x86_64-eng \
  --java-home /usr/lib/jvm/java-17-openjdk-amd64

./bin/lab workspace use android-14
./bin/lab workspace init
```

`workspace add` 参数：

| 参数 | 必需 | 默认值 | 说明 |
|---|---|---|---|
| `<name>` | 是 | 无 | 档案名称；只允许字母、数字、`-`、`_` |
| `--path` | 是 | 无 | AOSP 源码目录，必须是安全的绝对路径 |
| `--manifest` | 否 | `https://android.googlesource.com/platform/manifest` | Repo manifest 仓库地址 |
| `--branch` | 否 | `android-latest-release` | Repo manifest 分支或 Android tag |
| `--target` | 否 | `aosp_x86_64-eng` | 传给 `lunch` 的默认构建目标 |
| `--java-home` | 否 | 空 | 该 Workspace 使用的 JDK 路径 |

`--target` 会保存到档案的 `ANDROID_BUILD_TARGET`，作为 Build 工作流的默认构建目标。`CCACHE_DIR` 无需传入，系统会按档案名称自动生成独立路径。

管理多个版本：

```bash
./bin/lab workspace add android-15 \
  --path /data/aosp/android-15 \
  --branch android-15.0.0_r1 \
  --target aosp_x86_64-eng

./bin/lab workspace list
./bin/lab workspace use android-15
./bin/lab workspace current
./bin/lab workspace status android-14
./bin/lab workspace init android-14
```

### Workspace 档案与全局配置

| 文件 | 用途 |
|---|---|
| `config/workspaces/<name>.conf` | 当前正式的多 Workspace 配置；保存源码路径、manifest、分支、构建目标、Java 和 ccache 路径 |
| `.local/active-workspace` | 保存 `workspace use` 选择的活动档案名称 |
| `config/lab.conf` | 旧单 Workspace/本机全局配置的兼容文件，不用于多 Workspace 的选择和切换 |
| `config/lab.conf.example` | 旧配置格式模板；新增 AOSP 版本应优先使用 `workspace add` |

Workspace 档案与活动选择都是机器本地数据，已被 Git 忽略。`workspace use` 只切换活动配置，不会同步、构建或删除源码。不要将 token、密码、私钥或研究中的敏感证据提交到仓库。

## Repo 使用

Repo 命令默认操作活动 Workspace，也可以通过 `--workspace <name>` 指定档案：

```bash
./bin/lab repo status
./bin/lab repo init --workspace android-14 --partial-clone --no-use-superproject
./bin/lab repo init --workspace android-14 --partial-clone --no-use-superproject --apply

# 只显示同步计划，不访问网络
./bin/lab repo sync --workspace android-14 --jobs 8

# 显式执行同步
./bin/lab repo sync --workspace android-14 --jobs 8 --apply
```

实际 `repo init` 和同步的日志保存在 `output/repo/<workspace>/`。`repo sync` 未指定
`--jobs` 时默认使用主机逻辑 CPU 数；未指定 `--apply` 时不会修改 Workspace。
交互式终端运行时会显示 Repo/Git 的项目级进度，以及当前 fetch 的对象数、百分比
和已接收字节。由于 Git 对象数量和压缩复用是动态的，无法在开始前准确显示所有
项目最终需要下载的总字节数。Partial Clone 仅对新初始化的 Workspace 生效；
`--clone-filter` 必须与 `--partial-clone` 一起使用。

选择项目和增强重试：

```bash
./bin/lab repo sync --workspace android-14 \
  --project frameworks/base \
  --project system/core \
  --retry-fetches 3 \
  --no-clone-bundle \
  --jobs 4 \
  --apply
```

`--project` 可以重复；不指定时同步 manifest 中的所有项目。还支持 `--force-sync`，
它会在必要时覆盖指向不同 object directory 的 Git 目录并可能丢失 refs，只应在
明确理解影响时使用。

## AOSP Build

Build 默认只检查环境并显示计划，不执行编译：

```bash
./bin/lab build --workspace android-14
```

完整构建需要显式指定 `--apply`：

```bash
./bin/lab build \
  --workspace android-14 \
  --jobs 16 \
  --apply
```

构建指定模块时可以重复 `--module`：

```bash
./bin/lab build \
  --workspace android-14 \
  --target aosp_x86_64-userdebug \
  --module framework-minus-apex \
  --module services \
  --jobs 16 \
  --apply
```

`--target` 临时覆盖 Workspace 的默认构建目标，不修改档案。省略 `--module`
时执行完整 target；指定模块时执行模块构建。构建前会检查 Workspace、Repo、
`build/envsetup.sh`、target、磁盘空间、Java 和 ccache 配置。Go 负责参数、
路径保护、日志和状态；`scripts/aosp-build.sh` 只负责加载 AOSP Shell 环境并
执行 `lunch`、`m`。

查看最近一次结果：

```bash
./bin/lab build status --workspace android-14
```

日志、历史元数据和 `latest.json` 保存在
`output/build/<workspace>/`。详细说明见 [Build 工作流](docs/build.md)。

### 研究 Branch 与 Patch

Branch 用于在一个或多个 AOSP Git 项目中创建同名研究分支，隔离漏洞复现、
补丁移植和实验修改。它不会切换 Android 版本；Android 版本仍由 Workspace
档案中的 manifest branch/tag 决定。

#### 前置条件

开始前确认 Workspace 已完成 Repo 初始化，并检查当前状态：

```bash
./bin/lab workspace status android-14
./bin/lab repo status --workspace android-14
```

Branch 和 Patch 命令默认使用活动 Workspace；以下示例显式指定
`--workspace android-14`，避免在多个版本间误操作。

#### 查看研究分支

```bash
./bin/lab repo branch list --workspace android-14
```

该命令在 Workspace 根目录执行 `repo branches`，只读取各项目的分支状态。

#### 创建研究分支

先查看计划，不修改源码：

```bash
./bin/lab repo branch create binder-cve \
  --workspace android-14 \
  --project frameworks/base \
  --project frameworks/native
```

确认项目范围后显式创建：

```bash
./bin/lab repo branch create binder-cve \
  --workspace android-14 \
  --project frameworks/base \
  --project frameworks/native \
  --apply
```

实际等价于：

```text
repo start binder-cve frameworks/base frameworks/native
```

`--project` 可以重复。不指定任何 `--project` 时，Repo 会对 manifest 中的全部
项目创建同名分支，范围很大，建议研究时明确列出项目。Branch 名称会经过 Git
兼容性检查。当前不提供高风险的 branch delete。

#### 导出 Patch bundle

只导出工作树中已跟踪文件的未提交修改：

```bash
./bin/lab repo patch export \
  --workspace android-14 \
  --project frameworks/base
```

同时导出最近两个已提交修改：

```bash
./bin/lab repo patch export \
  --workspace android-14 \
  --project frameworks/base \
  --commits 2
```

可以重复 `--project`，一次导出多个项目：

```bash
./bin/lab repo patch export \
  --workspace android-14 \
  --project frameworks/base \
  --project frameworks/native \
  --commits 1
```

导出结构：

```text
output/repo/android-14/patches/<timestamp>/
  frameworks__base/
    0001-example.patch       # git format-patch 生成的已提交修改
    working-tree.diff       # git diff --binary HEAD 生成的未提交修改
    metadata.json           # Workspace、项目、manifest branch、基线 commit
    SHA256SUMS              # bundle 文件完整性校验
```

`--commits 0` 或不指定时不生成提交 Patch。当前未跟踪文件不会自动包含在
`working-tree.diff` 中；如需导出新文件，先在对应项目执行 `git add <file>`，
再运行 export。Patch bundle 位于可清理的 `output/`，重要研究成果应复制到未来
的 `research/<case>/` 长期资产目录。

验证导出文件：

```bash
cd output/repo/android-14/patches/<timestamp>/frameworks__base
sha256sum -c SHA256SUMS
```

#### 检查和导入 Patch

普通工作树 diff 默认只检查，不修改目标 Workspace：

```bash
./bin/lab repo patch import \
  --workspace android-15 \
  --project frameworks/base \
  --file /absolute/path/to/working-tree.diff
```

输出 `Patch check passed` 后，再显式应用：

```bash
./bin/lab repo patch import \
  --workspace android-15 \
  --project frameworks/base \
  --file /absolute/path/to/working-tree.diff \
  --apply
```

导入 `git format-patch` 文件：

```bash
./bin/lab repo patch import \
  --workspace android-15 \
  --project frameworks/base \
  --file /absolute/path/to/0001-example.patch \
  --apply
```

导入行为：

| 文件类型 | 检查 | `--apply` 行为 |
|---|---|---|
| `*.patch` | `git apply --check` | `git am`，保留作者、提交说明和提交元数据 |
| 其他 diff | `git apply --check` | `git apply`，只修改工作树，不自动创建提交 |

导入失败时不会自动使用 `--3way`、`--reject` 或强制覆盖。应先检查目标项目的
本地修改和冲突，再手动处理。`patch import` 每次只接受一个 `--project` 和一个
`--file`，避免 Patch 被错误应用到多个仓库。

#### 典型漏洞研究流程

```text
同步干净的 Android 14 Workspace
  -> 为相关项目创建 binder-cve 研究分支
  -> 修改源码并完成复现/修复实验
  -> 导出已提交 Patch 和未提交 diff
  -> 校验 SHA256SUMS 并归档研究资产
  -> 在 Android 15 Workspace 中先检查 Patch
  -> 显式导入、构建并验证跨版本影响
```

完整 Repo 说明见 [Repo 工作流](docs/repo.md)。
