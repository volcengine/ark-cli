---
name: caching-thinking
description: arkcli +chat 的 Caching / Thinking / ExpireAt 用法 reference, 配合 --store + chat get 实现可验证的多轮缓存与思考开关。
---

# +chat Caching / Thinking / ExpireAt

这三组 flag 让 `+chat` 真正能控制服务端缓存、推理深度和持久化生命周期, 同时把状态回显到响应里方便 agent 验证。

## 何时使用

- **Caching** —— 多轮对话相同 system / 工具定义重复, 想跑 prompt cache 省钱; 或反过来要禁缓存测纯净行为。
- **Thinking** —— 模型支持 reasoning 但本轮不想让它思考 (省 token); 或想强制让它思考。
- **ExpireAt** —— 想给 `--store` 过的 response 加个过期, 不让它无限存储。

## flag 速查

| flag | 类型 | 取值 / 形态 |
|---|---|---|
| `--caching` | string | `enabled` / `disabled` (空表示用服务端默认) |
| `--cache-prefix` | bool | 与 `--caching` 联用; 启用 prefix-cache (autotest 的 `prefix_cache_test` 在用)。**与 `--max-output-tokens` 互斥** —— 服务端约束 `caching.prefix is not supported when max_output_tokens is set`，arkcli 会在客户端就拦截这个组合 |
| `--thinking` | string | `auto` / `enabled` / `disabled` (空表示用服务端默认) |
| `--expire-at` | int (epoch sec) | 配合 `--store` 使用; 不传则用服务端默认 GC 策略 |

## 典型组合

### 开缓存的多轮对话

```bash
# 第一轮: 启用缓存 + store
RID=$(arkcli +chat "草莓什么颜色" --model ep-xxx \
  --caching enabled --store \
  --format json | jq -r .id)

# 第二轮: 同样启用缓存, 接续 RID
arkcli +chat "苹果呢?" --model ep-xxx \
  --caching enabled --store \
  --previous-response-id "$RID"

# 核对缓存配置（不是命中证明）: chat get 拿回响应
arkcli chat get "$RID" --format json | jq .caching
# → {"type":"enabled"}
```

### 关思考压短输出

```bash
arkcli +chat "用 30 字介绍 LLM" --model ep-xxx \
  --thinking disabled --max-output-tokens 100
# 没有 thinking phase, 直接出 response
```

### 持久化 + 过期

```bash
TS=$(($(date +%s) + 3600))   # 1 小时后过期
arkcli +chat "记住我叫小明" --model ep-xxx \
  --store --expire-at "$TS"

# 校验
arkcli chat get "$RID" --format json | jq '{store, expire_at}'
# → {"store":true, "expire_at":1746676800}
```

## 输出形态

非流式时, `+chat` 与 `chat get` 都返回扩展过的 `ResponsesResult`:

```json
{
  "id": "resp_...",
  "object": "response",
  "status": "completed",
  "created_at": 1746673200,
  "model": "...",
  "content": "...",
  "reasoning_content": "...",
  "usage": { ... },
  "store": true,
  "expire_at": 1746676800,
  "previous_response_id": "...",
  "caching": {"type": "enabled", "prefix": null},
  "thinking": {"type": "disabled"},
  "reasoning": {"effort": "medium"}
}
```

流式需要机读证据时使用 `--stream --include-events`，核对成功终态中实际返回的
response 字段；普通 `--stream` 是人类可读文本，不是上面的扁平 JSON。
未回显的字段保持未知，不为补字段再次发起推理。详见 [stream-events.md](stream-events.md)。

## 验证清单 (autotest 对应)

| autotest 用例 | 是否解锁 |
|---|---|
| `Test_ResponseCreate_ExpireAtAndCaching` | ✅ |
| `Test_ResponseCreate_MaxOutputTokens` (含 Thinking disabled) | ✅ |
| `Test_ResponseCreateAndGet_PersistedFields` | ✅ (配合 PR-1 的 chat get) |
| `cache/cache_test.go` 8 个用例 | ✅ |
| `cache/prefix_cache_test.go` | ✅ (用 --cache-prefix) |
| `partial/partial_mode_test.go` | ✅ (Thinking + Caching 组合) |
| `Test_Stream_ExpireAtAndCaching` | 使用 --include-events 检查实际终态 |
| `cache/cache_stream_test.go` 5 个用例 | 检查实际终态，配置本身不证明命中 |

## 常见错误

| 现象 | 原因 |
|---|---|
| `unsupported caching.type "X"` | 取值不是 `enabled`/`disabled` |
| `unsupported thinking.type "X"` | 取值不是 `auto`/`enabled`/`disabled` |
| `--cache-prefix` 不生效 | 没同时传 `--caching enabled`; 服务端默认拒绝单独 prefix |
| `--expire-at` 不生效 | 没传 `--store`; 服务端只对 stored response 应用 GC |
| 设置后 `chat get` 拿不到回显 | 该 response 没 `--store` 或已过期 |

## 与 reasoning-effort 联动

`--reasoning-effort` 与 `--thinking` 是两个独立维度：

- 明确要求关闭思考时使用 `--thinking disabled`；`--reasoning-effort minimal` 只是强度参数，不能替代关闭开关。
- CLI 将两者分别传为 Responses 请求的 `thinking.type` 和 `reasoning.effort`。交付 SDK/HTTP 示例时保留用户明确的值，不把 effort 写进 thinking，也不凭经验自动补 high。
- 参数组合是否支持、最终是否采用由目标模型和服务端决定；保留实际错误与回显，不保证所有模型都接受相同组合。
- `caching.type=enabled` 只说明配置；只有实际 usage 中的缓存命中统计才能证明命中。缺少统计时标为未知，不能声称已省下具体 token 或费用。
- 比较延迟时记录模型版本、输入、参数、缓存状态和实际耗时；未实测时不宣称 disabled/minimal/high 有跨模型固定排序，也不为验证猜测擅自增加付费请求。
