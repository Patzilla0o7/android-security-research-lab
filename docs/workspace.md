# 多 Workspace 管理

每个 AOSP 版本使用独立档案、源码目录、分支、构建目标和 ccache 目录。

## 添加与切换

```bash
./bin/lab workspace add android-14 --path /data/aosp/android-14 --branch android-14.0.0_r75
./bin/lab workspace add android-15 --path /data/aosp/android-15 --branch android-15.0.0_r1 --target aosp_x86_64-eng
./bin/lab workspace list
./bin/lab workspace use android-15
./bin/lab workspace current
```

第一个档案会自动成为活动 Workspace。`use` 只更新当前选择，不会同步源码、
执行构建或删除任何数据。

## 状态与初始化

不指定名称时操作活动档案；指定名称时不改变当前选择：

```bash
./bin/lab workspace status
./bin/lab workspace init
./bin/lab workspace status android-14
./bin/lab workspace init android-14
```

`init` 只创建经过安全检查的源码目录，不执行 `repo init`。

## 本机数据

- `config/workspaces/<name>.conf`：Workspace 配置档案。
- `.local/active-workspace`：当前活动档案名称。

两者都被 Git 忽略。档案名称只允许字母、数字、`-`、`_`，且不能让两个
档案指向同一路径。删除配置或源码的命令目前未提供。
