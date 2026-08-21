# Research 项目

`lab research` 管理长期漏洞研究资产。一个 case 对应 `research/<case-id>/` 下的一个
独立项目。

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

`validate` 检查 case schema、必需模板和目录。Patch、PoC、复现记录和报告直接由
研究人员维护在对应项目目录中。
