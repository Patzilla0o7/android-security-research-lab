# Output

Output 保存运行过程中产生的数据。

例如：

Build Log

logcat

Bugreport

Screenshot

Tombstone

Output 不属于长期资产。

可以随时清理。

设备证据统一保存到：

```text
output/evidence/<workspace>/<case-id>/<UTC timestamp>/
```

每个目录包含 `manifest.json` 和 `SHA256SUMS`。清理或归档前应确认 case ID、设备
serial、采集状态和校验值；重要证据应转移到受控的长期存储。
