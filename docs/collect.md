# 证据采集

`lab collect` 将设备证据绑定到 Workspace、研究 case 和设备 serial。第一阶段支持
设备信息、当前 logcat 缓冲区、截图、bugreport、tombstones 和完整 bundle。

```bash
./bin/lab collect device-info --case CVE-2026-0001
./bin/lab collect logcat --case CVE-2026-0001 --serial emulator-5554
./bin/lab collect screenshot --case CVE-2026-0001
./bin/lab collect bugreport --case CVE-2026-0001 --timeout 15m
./bin/lab collect tombstones --case CVE-2026-0001
./bin/lab collect bundle --case CVE-2026-0001 --workspace android-15
```

`--case` 必须指定，只允许字母、数字、点、短横线和下划线。多设备环境必须使用
`--serial`。省略 `--workspace` 时使用活动 Workspace。

## 输出结构

```text
output/evidence/<workspace>/<case-id>/<UTC timestamp>/
├── device.json
├── logcat.txt
├── screenshot.png
├── bugreport.zip
├── tombstones.tar
├── manifest.json
└── SHA256SUMS
```

`manifest.json` 记录 schema 版本、操作、Workspace、case、设备、采集时间、执行过的
ADB 参数、成功/失败状态以及每个证据文件的大小和 SHA-256。`SHA256SUMS` 同时覆盖
证据文件和 manifest。

`logcat` 使用 `adb logcat -d -v threadtime` 获取当前缓冲区，不会持续阻塞。
截图使用原始 `exec-out screencap -p` 数据并检查 PNG 文件头；bugreport 由 ADB
直接写入 ZIP；tombstones 通过设备端 `tar` 输出，通常需要 root 或对
`/data/tombstones` 的读取权限。

默认总超时为 10 分钟，可用 `--timeout` 调整。完整 bundle 依次尝试所有采集项；
某一步失败时其他步骤继续执行，已经取得的文件仍会保留，manifest 状态记为
`partial`，命令整体返回非零。只有完全没有取得证据时状态才是 `failed`。

证据目录属于可清理的 `output/`，重要材料应在确认隐私和敏感信息后归档到对应的
Research 项目或外部证据存储。

## 检查与验证

```bash
./bin/lab collect inspect output/evidence/android-15/case-1/20260821T120000Z
./bin/lab collect verify output/evidence/android-15/case-1/20260821T120000Z
```

`inspect` 显示 manifest 摘要、文件、大小、采集状态，并执行完整性验证。`verify`
执行同样的严格验证，适合脚本和 CI。验证会拒绝：

- SHA-256 不匹配、缺失或重复条目；
- 未被 `SHA256SUMS` 覆盖的额外文件；
- 未覆盖 `manifest.json` 的 bundle；
- 路径穿越、符号链接和其他非普通文件；
- 不支持的 manifest schema。

## 脱敏计划

```bash
./bin/lab collect redact <bundle> --plan
```

计划模式要求 bundle 先通过完整性验证，只读扫描文本、ZIP 和 TAR 内容。当前规则
覆盖邮箱、IPv4、MAC、电话号码、常见 credential 字段、Android 标识符、Wi-Fi
标识符和用户路径。输出只包含文件位置、规则名称和命中数量，不显示原始敏感值。

PNG 等图像会标记为人工检查。单个文本或归档条目最多扫描 16 MiB，较大的条目会
跳过，避免异常证据造成过量内存消耗。`--plan` 不修改任何文件；生成独立脱敏副本的
`--apply` 尚未实现。
