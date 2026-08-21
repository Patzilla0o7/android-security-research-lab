# Tools

保存第三方工具。

例如：

jadx

apktool

Frida

dex2jar

objection

工具统一放这里。

不要放到 Workspace。

## DroidForge

`tools/DroidForge` 是用于管理 Android SDK、系统镜像、AVD 和 Emulator 的 Git
submodule。初始化和构建：

```bash
git submodule update --init --recursive
./scripts/build-all.sh
```

ASRL 不在运行时直接调用 DroidForge。DroidForge 启动模拟设备后，`lab device`
通过标准 ADB 接口发现和检查该设备。
