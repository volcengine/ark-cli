# arkcli-models evals

## 覆盖目标

- 模型搜索需求应使用 `arkcli models search`，避免用 `list` 做能力推荐。
- 模型资产盘点需求应使用 `arkcli models list --page-all`，避免跳到 Raw API Explorer。
- "最近 N 天自定义模型"应在本地 JSON 中按 `create_time` 和自定义模型字段过滤，不要求用户去控制台。

## Trigger / 该唤起

- 用户明确问模型清单、模型详情、模型搜索、模型统计时，该唤起 `arkcli-models`。
- 用户问"我的/自定义/最近创建"这类资产盘点时，该唤起 `models list` 路径。

## Anti-trigger / 反唤起

- 用户要直接聊天、生成、部署时，不该把 `arkcli-models` 当终点；只在缺模型名时短暂进入。
- 用户明确要 Raw API Action、OpenAPI 列表或底层 params 时，不该用本 skill，应转 `arkcli-api-explorer`。

## Guard / 守卫

- 认证失败、401、登录状态不明时，先检查 `arkcli auth status`。
- profile、region、base URL 异常时，转 `arkcli-config` 做配置排查。
- 缺少服务端时间过滤 flag 不是失败条件；先用本地 JSON 过滤，确认无法完成后再说明限制。

## happy-path CLI 实测命令

```bash
arkcli models list --page-all --sort-by CreateTime --sort-order Desc --format json
arkcli models search --modality text --min-context-window 200000 --capability thinking --strict-filter
```

## 回归用例

| case | prompt | 期望 |
|------|--------|------|
| `uat-models-list-recent-custom-001` | 我在方舟最近 7 天建了多少个自定义模型？列下来。 | 使用 `arkcli models list --page-all --format json`，本地过滤 `create_time` 和 Custom 字段；禁止 `arkcli api --list`，禁止建议去控制台 |
| `models-list-owned-custom` | 帮我看一下我有哪些自定义模型，输出数量和模型名。 | 使用 `arkcli models list` 做资产清单枚举，再做客户端过滤和统计 |
| `models-search-task-fit` | 帮我找一个支持 thinking、上下文 200K 以上的文本模型用于 +chat。 | 使用 `arkcli models search --min-context-window ... --capability thinking` |
| `models-search-speech-boundary` | 方舟广场有没有 TTS 模型？可以直接部署或生成示例代码吗？ | 使用 `arkcli models search <tts/语音关键词>` 做广场发现；明确说明语音模型当前不支持 `+deploy` / `+code-example` / usage / pricing / onboard；不得推荐 `+chat` 或 `+gen` |

## 判分重点

- 必须路由到 `arkcli-models`。
- 资产盘点必须推荐 `arkcli models list`，优先带 `--page-all` 和 `--format json`。
- 时间过滤必须说明客户端过滤或本地 JSON 处理。
- 不得推荐 `arkcli api --list` 作为首选路径。
- 不得因为缺少服务端时间 flag 就说 CLI 无能力或让用户去控制台。
- 语音模型命中后必须停在广场发现层，不得继续推荐部署、示例、用量或费用查询。
