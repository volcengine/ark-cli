# 常用全局 flags

> 本文从 `arkcli-shared` 正文拆出。构造复杂调用、或排查 project-name / base-url / region 覆盖时查阅。各命令自身的 flag 以 `arkcli <domain> <verb> --help` 为准。

| Flag | 作用 |
|------|------|
| `--profile` | 指定活动 profile |
| `--api-key` | 覆盖 ARK API key |
| `--project-name` | 覆盖 ARK Project Name（全局唯一入口；优先级：flag > `ARK_PROJECT_NAME` > `profile.project` (profile.yaml) > identity store apikey 关联 project > `.env`（火山 `VOLCENGINE_ARK_PROJECT_NAME`，老用户兼容）） |
| `--base-url` | 覆盖 base URL |
| `--region` | 覆盖 region |
| `--transform` | 对输出做路径提取 |
| `--page-all` | 自动翻页 |
| `--page-limit` | 限制翻页次数 |
| `--page-delay` | 翻页间隔（ms） |
| `--dry-run` | 预览请求 |
| `--debug` | 输出调试信息到 stderr |
