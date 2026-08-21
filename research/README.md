# Research

Research 保存所有漏洞研究的长期资产和结构化索引。

一个漏洞对应一个 Project。

包括：

- 已知 CVE
- 自发现漏洞
- Patch
- Root Cause
- PoC
- EXP
- 调试记录

Research 是实验室最重要的数据。

使用 `lab research new <case-id> --title <title>` 创建标准项目；一个漏洞一个目录，
不要多个漏洞共用一个目录。大型原始证据不直接提交到 Git，通过
`artifacts/evidence.json` 保存路径和 SHA-256 关联。

完整工作流见 [Research 项目](../docs/research.md)。
