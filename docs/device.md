# 设备管理

`lab device` 通过标准 ADB 接口发现和检查已经运行的 Android 设备。设备可以由
DroidForge、Android Studio、其他 Emulator 管理器或真实硬件提供；ASRL 不直接
调用或控制 DroidForge。

## 列出设备

```bash
./bin/lab device list
```

输出包括 serial、ADB 状态、型号、产品和 transport ID。`offline` 与
`unauthorized` 设备会显示在列表中，但不能用于状态读取或等待操作。

## 查看状态

```bash
./bin/lab device status
./bin/lab device status --serial emulator-5554
```

状态包括 Android 版本、API level、build fingerprint、build type、debuggable、
启动完成属性与 SELinux 模式。如果同时连接多个可用设备，必须使用 `--serial`，
避免读取或操作错误的研究目标。

## 等待启动

```bash
./bin/lab device wait --serial emulator-5554 --timeout 2m
```

该命令先等待 ADB 连接，再轮询 `sys.boot_completed`。连接或启动超时会返回非零
退出码。默认超时时间为两分钟。

## 与 DroidForge 的边界

DroidForge 作为 `tools/DroidForge` Git submodule 管理 SDK、系统镜像、AVD 和
Emulator 生命周期。ASRL 只观察 DroidForge 启动后暴露的 ADB 设备。两个项目可
独立升级，也不会让 `lab` 在运行时依赖 `droidforge` 二进制。
