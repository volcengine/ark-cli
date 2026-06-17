# Profile Resources Default 偏好（含漂移检测与跨模态）

> 本文从 `arkcli-shared` 正文拆出。仅在跑 `+chat` / `+gen` / `+deploy` 且涉及 `--model` / `--endpoint` / 默认资源选择时加载。常规读任务不需要读本文。

profile.yaml 在 `resources.<modality>.default` 字段存用户对每个 modality (text / image / video) 默认资源 ID 的偏好。这是 +chat / +gen / +deploy 行为对齐的基础。

- **platform profile**：默认值是用户已部署的 endpoint id (`ep-xxx`)
- **agent-plan profile**：默认值是火山预置 model id（如 `doubao-seed-2-0-pro-260215` / `doubao-seedream-5.0-lite` / `doubao-seedance-2.0`）
- **coding-plan profile**：text modality 用 `ark-code-latest` 这类套餐内文本模型；image / video 借道 platform 数据面 — 字段存的是用户在 platform 上 `+deploy` 出来的 endpoint id（`ep-xxx`），arkcli 会在 `+gen` 时自动把数据面切到 `/api/v3`。不再要求切 profile
- **团队版 `agent-plan-team` / `coding-plan-team` profile**：数据面行为**等同对应个人版** —— agent-plan-team 同 agent-plan（自带生文/生图/生视频预置模型，chat/understand/gen 全本地）；coding-plan-team 同 coding-plan（text 用套餐内文本模型，image/video 借道 platform endpoint）。base_url / 模型派发完全一致，区别仅在凭证来自团队席位（`GetSeatInfo` 的明文 ApiKey）

每次帮用户跑 `+chat` / `+gen`：

- 用户没传 `--model` → CLI 自动 fallback 到 `profile.resources.<modality>.default`
- 用户传了 `--model <X>`、且 `X` ≠ 当前 default → 见下文"漂移检测"
- 默认值为空时 CLI 已 fail-fast，错误体里会给 `arkcli resources list --modality <m>` 和 `arkcli profile set-default --modality <m> <id>` 两条引导，agent 直接照做

## Default 漂移检测与 promote nudge

在帮用户成功跑 `arkcli +chat` / `+gen` / `+deploy` 且用户**显式**传了 `--model X`（或 `--endpoint X`、platform 场景）时：

1. 读 profile.yaml 的 `resources.<modality>.default`
   - `+chat` → modality = text
   - `+gen --modality image` → modality = image
   - `+gen --modality video` → modality = video
   - `+deploy` 由用户传的 `--set-default <modality>` 自带（用户已经显式说了）
2. 比较 `X` 跟当前 default
3. **命令成功之后**（数据面返回 200 / `+deploy` 返回 ep id），如果 `X` ≠ 当前 default：
   > "你这次用了 `X`，但当前 `<modality>` default 是 `Y`。要把 `X` 设为新 default 吗？"
4. 用户同意 → 跑 `arkcli profile set-default --modality <m> <X>`（已经包含 inline 校验，不需要 agent 二次校验）
5. 用户拒绝 → 不动 profile

**何时不 nudge**：
- 命令失败 → 不 nudge（用户误调用不该被推 promote）
- 用户没传 `--model` 走的就是 default → 不 nudge
- 同一 session 已经 nudge 过同一个 `(modality, id)` → 不再问（避免 nag）
- 用户明确说"别再问 X" → 写到 CLAUDE.md memory 持久化偏好

## Cross-modality 处理 (S10 后)

S10 改造后, arkcli 不再因为 active profile 是 `coding-plan` 就阻断 `+gen --modality image|video`. 行为如下:

- **active=platform**: 直接走 platform 数据面 `/api/v3` + endpoint id, 跟原行为一致
- **active=agent-plan**: 走 agent-plan 数据面 + 预置模型 (image=`doubao-seedream-5.0-lite`, video=`doubao-seedance-2.0`)
- **active=coding-plan + modality=image|video**: arkcli 自动借道 platform 数据面 `/api/v3` + 用户在 platform 控制面 `+deploy` 出的 endpoint id (cfg 临时 clone, 不污染 factory cache). 用户体感不需要切 profile
- **active=agent-plan-team / coding-plan-team（团队版）**: 跟对应个人版完全一致 —— agent-plan-team 同 agent-plan（image/video 用预置模型）；coding-plan-team 同 coding-plan（image/video 借道 platform `/api/v3` + endpoint id）

资源发现 / default 设置链路也对齐:
- `arkcli resources list --modality image` 在 coding-plan profile 下会列出同账号下 platform 已 deploy 的 endpoint
- `arkcli profile set-default --modality image <ep-id>` 在 coding-plan profile 下也能 verify + 写入 (走同款 ListEndpoints)

`+gen --modality image` 在 `agent-plan` profile 下默认会用 `doubao-seedream-5.0-lite`, `--modality video` 默认 `doubao-seedance-2.0`. 用户跑 `arkcli resources list --modality image` / `video` 可看完整可用清单.
