# profile set-default 详细参考

> **前置**：先读 [`../SKILL.md`](../SKILL.md)。

`profile set-default` 显式表达"在某 modality 下用什么资源作 default"的偏好，写到 `profile.yaml` 的 `Resources.<modality>.Default`。`+chat` / `+gen` 命令在 `--model` 缺省时按这里 fallback。

## 命令模板

```bash
# 设 text default（chat 用）
arkcli profile set-default --modality text doubao-seed-2-0-pro-260215

# 设 image default（gen --modality image 用）
arkcli profile set-default --modality image doubao-seedream-5.0-lite

# 设 video default（gen --modality video 用）
arkcli profile set-default --modality video doubao-seedance-2.0

# 显式指定 profile（多账号场景）
arkcli profile set-default --profile platform_cn-beijing_default ep-xxx

# 跳过 inline verify（控制面暂不可达但 id 已知有效）
arkcli profile set-default --modality text --skip-verify doubao-seed-2-0-pro-260215
```

## 关键 inline verify 行为

默认情况下 set-default 会先调控制面校验 `<id>` 在当前 profile 的可用列表里：

| profile.Type | 校验来源 |
|--------------|---------|
| `platform` | `ListEndpoints` (PageAll) + modality filter — endpoint 自带 modality 必须跟 `--modality` 匹配, text endpoint 不能写到 image/video default. unknown modality (custom endpoint) 放行避免错杀 |
| `agent-plan` | text: `ListAgentPlanLatestModel`<br>image: `AgentPlanImageModels` 硬编<br>video: `AgentPlanVideoModels` 硬编 |
| `coding-plan` | text: `ListArkCodeLatestModel` (要 SSO AccountID)<br>image/video: 借道 platform `ListEndpoints` + modality filter (S10, commit f69be53), 校验跟 platform 同款 |

**校验用 target profile 的身份**（codex P0-A 修正）：`--profile X` 时通过 `rt.RebuildFactoryForProfile(X)` 重建 invoker，避免 active=A 的 ListEndpoints 结果被当成 B 的可选范围。

## 跟 `+chat / +gen` 的衔接

```
+chat / +gen --model 未传:
  ↓
  resolveModelDefault(active profile name)  ← P0-B: 按 active profile 解析, 不是 DefaultProfile
  ↓
  读 profile.Resources.<modality>.Default
  ↓
  S9 runtime alias: plan 类 + text + default=="auto" → 替换为 "ark-code-latest"
  ↓
  发数据面请求
```

如果 `Resources.<modality>.Default` 为空，会报：
```
--model 未传, profile "platform_cn-beijing_default" text default 也未设.
  请显式 --model <id>, 或先跑
    arkcli resources list --modality text              # 看可用 ID
    arkcli profile set-default --modality text <id>    # 设默认
```

Agent 看到此 hint 时，按 hint 跑两条命令即可，不要自己猜默认值。

## 输出形态

stdout：
```json
{
  "profile": "platform_cn-beijing_default",
  "modality": "text",
  "new_default": "doubao-seed-2-0-pro-260215",
  "verified": true
}
```

stderr：
```
[arkcli] profile "platform_cn-beijing_default" text default 已设为 "doubao-seed-2-0-pro-260215"
```

`verified=false` 表示加了 `--skip-verify`。
