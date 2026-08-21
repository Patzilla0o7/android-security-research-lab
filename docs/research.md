# Research 项目

`lab research` 管理长期漏洞研究资产。一个 case 对应 `research/<case-id>/` 下的一个
独立项目；大型原始证据仍保存在 `output/evidence/` 或外部受控存储中。

## 创建项目

```bash
./bin/lab research new CVE-2026-0001 \
  --title "Binder permission bypass" \
  --workspace android-15
```

省略 `--workspace` 时使用活动 Workspace。case ID 只允许字母、数字、点、短横线和
下划线。命令拒绝覆盖已有项目。

生成结构：

```text
research/<case-id>/
├── case.yaml
├── README.md
├── timeline.md
├── reproduction.md
├── root-cause.md
├── patches/
├── poc/
├── artifacts/
│   └── evidence.json
└── reports/
```

`case.yaml` 使用 schema version 1，记录 case ID、标题、状态、Workspace、受影响
组件、披露状态和 UTC 时间。当前支持的研究状态为 `investigating`、`reproduced`、
`patched`、`disclosed` 和 `closed`。

## 查看和校验

```bash
./bin/lab research list
./bin/lab research show CVE-2026-0001
./bin/lab research validate CVE-2026-0001
```

`validate` 检查 case schema、必需模板和目录，并验证所有已关联证据。

## 关联证据

```bash
./bin/lab research evidence add CVE-2026-0001 \
  --bundle output/evidence/android-15/CVE-2026-0001/20260821T120000Z

./bin/lab research evidence list CVE-2026-0001
./bin/lab research evidence verify CVE-2026-0001
```

添加前会验证 bundle 中的 `SHA256SUMS`、拒绝路径穿越和符号链接，并要求 evidence
manifest 的 case ID 与 Research 项目一致。索引只保存 bundle 路径、设备、
Workspace、状态、采集时间和 manifest SHA-256，不复制 bugreport 等大型文件。

`verify` 会重新校验所有文件和 manifest 摘要；任何证据被修改、删除或替换都会返回
非零状态。移动外部证据后，需要在确认新位置和完整性后重新建立关联。
