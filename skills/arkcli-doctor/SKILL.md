---
name: arkcli-doctor
version: 1.0.0
description: "arkcli doctor 诊断总入口（一个 skill 收口四个 domain：错误码 / infer-endpoint / model / metrics）：把火山方舟（Ark）使用中的报错、资源状态、用量配额、性能指标统一收口到 `arkcli doctor` 命令家族（CLI 健康 / account / infer-endpoint / model / metrics / 错误码）。当用户/Agent 拿到一个 Ark 错误码（ModelAccessDenied / InvalidParameter / RateLimitExceeded / ContentRiskBlocked / OutputVideoSensitiveContentDetected 等），或拿到一段 Ark 错误 JSON、一个 ep-xxx 接入点 ID、一个模型名，问『为什么报错 / 怎么修 / 状态怎么样 / 用量多少 / 配额压力 / 是不是被冻结 / 权限够不够 / 接入点慢/挂了 / 模型整体怎么样 / 哪个 endpoint 把模型配额吃了 / 生视频拦截怎么修 / RPM 多少 / 错误率多少 / P99 时延多少 / Cache 命中率怎么样』时使用；用户说『arkcli doctor / 诊断 / 体检 / 健康检查 / 排查 / 自助修复 / 看一下指标 / 拉一下监控数』时也走这里。**模型名 + 异常/体检语义也归这里**：用户说『我用 <模型> 一直/经常/最近 失败 / 慢 / 超时 / 被限流 / 报错』『<模型> 失败率高 / 一批请求大半失败 / 这模型怎么了 / 现在还能用吗 / 是不是被限流 / 配额够吗』『看下 <模型> / 诊断下 <模型> / 体检 <模型> / <模型> 健康吗 / <模型> 整体什么情况』时，对应跑 `arkcli doctor model <name>`（即便用户没拼出错误码，跨任务统计性失败 / 持续性体感异常 / 配额视角 / 显式诊断动词 都走这里，不要被『生成视频/seedance/seedream』关键词带去 arkcli-gen）。生视频拦截类错误码（Input/Output…SensitiveContentDetected / PrivacyInformation / PolicyViolation）继续在本 skill 的 references 里覆盖。注意：单纯的 arkcli 安装/登录/profile/认证问题归 arkcli-shared / arkcli-auth；本 skill 只负责 `arkcli doctor` 命令家族。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli doctor --help"
---

# arkcli Doctor（诊断总入口）

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)（认证闸门、命令选择顺序、输出与安全规则）。**

## 它解决什么

`arkcli doctor` 是 Ark CLI 唯一的对外诊断收口。所有诊断与修复建议都从 `arkcli doctor [<scope>] [<id>]` 出，**默认即诊断**（只读体检），`--fix` 是诊断的附加模式（仍需用户二次确认）。Agent 拿到用户的报错 / 资源 ID / 含糊问题后，先用本 skill 决定走哪条路径，再 delegate 到对应 scope reference 或错误码 reference。

> 边界：**安装、版本、登录、profile、AK/SK 认证**类问题归 [`arkcli-shared`](../arkcli-shared/SKILL.md) / [`arkcli-auth`](../arkcli-auth/SKILL.md)；这里只覆盖 `arkcli doctor` 命令家族（业务侧诊断）。

## 三个 Domain（一个 skill 收口）

本 skill 一次性覆盖四个平行 domain。每个 domain 一份 reference，按用户场景路由：

| Domain                | 命令                                | Reference                                                                                       | 解决什么                                                                  |
|-----------------------|-------------------------------------|-------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|
| **错误码**            | `arkcli doctor error <code>`        | [`references/error-codes.md`](references/error-codes.md)                                       | 单错误码翻译成结构化诊断 + 修复（含生视频 5 个 subtype + 通用 Ark/公共码） |
| **infer-endpoint**    | `arkcli doctor infer-endpoint <id>` | [`references/scope-infer-endpoint.md`](references/scope-infer-endpoint.md)                     | 单接入点状态 / 用量 / 错误率 / 配额压力                                  |
| **model**             | `arkcli doctor model <name>`        | [`references/scope-model.md`](references/scope-model.md)                                       | 跨接入点的模型整体诊断（用量 / 配额 / top endpoint / 模态自适应指标）     |
| **metrics**           | `arkcli doctor metrics <id>`        | [`references/scope-metrics.md`](references/scope-metrics.md)                                   | 36 条具名 PromQL 查询（指标值直出，不打健康判定，给 LLM/SRE 拿数）        |

## 命令家族

```
arkcli doctor                                  # CLI 健康（默认 scope，install/connect/config）
arkcli doctor infer-endpoint <ep-id>           # 单接入点诊断（→ references/scope-infer-endpoint.md）
arkcli doctor model <model-name>               # 模型维度诊断（→ references/scope-model.md）
arkcli doctor metrics <query-id>               # 具名 PromQL 查询（→ references/scope-metrics.md）
arkcli doctor error <error-code>               # 错误码查表（只读，→ references/error-codes.md）
arkcli doctor <scope> [<id>] --fix             # 触发修复（先 dry-run，需用户二次确认）
```

> 命令是**扁平 verb**：没有 `diagnose / fix / list` 子动词，`--fix` 只是 flag。`--window <duration>`（默认 24h）控制 VMP 时间窗。`--format pretty|json`（TTY 默认 pretty，pipe 默认 json）。

## 核心范式：从用户消息到答案

> [!WARNING]
> **第一步 MUST 做：扫一遍用户消息里有没有资源 ID（模型名 like `doubao-*`/`seed-*` 或 `ep-xxx`）**。**只要有资源 ID，就 MUST 走 `doctor <scope> <id>`**，不管同时有没有错误码。资源 ID 在场时直接 `doctor error <code>` 是**错误路径**——会丢掉错误率分布、top endpoint、配额压力等关键诊断信号。
>
> 反例自检（如果你打算这么干，停下来）：
> - ❌ 用户说「我的 doubao-seedance-1-0-pro 一直报 ContentRiskBlocked」 → 你跑 `arkcli doctor error ContentRiskBlocked`
> - ✅ 正解：跑 `arkcli doctor model doubao-seedance-1-0-pro`，再按返回的错误码分布加载 [`error-codes.md`](references/error-codes.md) 的 subtype 段
>
> - ❌ 用户说「ep-xxx 报 ModelAccessDenied」 → 你跑 `arkcli doctor error ModelAccessDenied`
> - ✅ 正解：跑 `arkcli doctor infer-endpoint ep-xxx`，再加载 `error-codes.md` 的 model_access_denied 段

```
 用户消息（自然语言 / 错误 JSON / "看看 ep-xxx" / 错误码 / request_id）
        │
        ▼
 ① 抽取关键值：error_code / error_message / resource_id (ep-xxx / model-name) / request_id
        │
        ▼
 ② 按下表决定 Path（**有 resource_id 一律 Path 1 / Path 3**，不要降级到 Path 2）
        │
        ▼
 ③ 跑相应 doctor 命令 + 加载对应 reference
        │
        ▼
 ④ 摘 findings + recommended_fixes 给用户（不复述完整 JSON）
        │
        ▼
 ⑤ 修复命令一律先给用户看、等确认；--fix 不替用户先按
```

## 路径决策表

|              | **有错误码**                                                                 | **无错误码**                                                                 |
|--------------|------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| **有资源 ID**| **Path 1（最理想）**：跑 `doctor <scope> <id>` + 加载 [`error-codes.md`](references/error-codes.md) 对应 subtype | **Path 3**：跑 `doctor <scope> <id>` 看整体（状态/用量/错误率/配额）         |
| **无资源 ID**| **Path 2**：直接 `doctor error <code>` 查表 → 加载 [`error-codes.md`](references/error-codes.md)；不跑 scope | **Path 4/5**：只有 request_id 或啥都没有 → 追问用户要资源 ID 或具体错误信息 |

> 暂不支持 request_id 反查能力（涉及鉴权与平台内部数据）。Path 4/5 不要瞎猜，直接追问。

> [!IMPORTANT]
> **Path 1 vs Path 2 不要走错**：只要用户消息里既给了**错误码**又给了**资源 ID**（模型名 / `ep-xxx`），**MUST** 走 Path 1（先 `doctor <scope> <id>` 拿到该资源在该错误码上的真实分布 / 占比 / top endpoint，再加载 `error-codes.md` 对应 subtype 解读修复）。**不要**直接 `doctor error <code>` 当成 Path 2 处理——那会丢失模型上下文（错误率分布、top endpoint、配额压力等关键诊断信号）。
>
> Path 1 触发条件示例：
> - 「我的 doubao-seedance-1-0-pro 一直报 ContentRiskBlocked」 → `doctor model doubao-seedance-1-0-pro`，**不是** `doctor error ContentRiskBlocked`
> - 「ep-xxx 报 ModelAccessDenied」 → `doctor infer-endpoint ep-xxx`，**不是** `doctor error ModelAccessDenied`
> - 只有用户**只给错误码 / 错误 JSON**、没给资源 ID 时（Path 2），才直接查表。

> [!NOTE]
> **不要为 `doctor model` / `doctor infer-endpoint` 做兜底前置调用**：这两条命令内部已经先跑 `model.exists` / `endpoint.exists`，覆盖了"模型/接入点是否存在"。**不要**在调用前先跑 `arkcli models list` / `arkcli infer list` / `arkcli auth status` 等做存在性 / 身份兜底——doctor 命令失败时会自己报清楚原因（404 / VMP precheck / 鉴权），按它的输出处理即可，多余兜底徒增 token 消耗。

## 路由到 scope reference

按 scope 找 reference：

| scope                 | reference                                                                                       | 触发的用户语义                                                              |
|-----------------------|-------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| `account`             | 暂无独立 reference；账号级体检（identity / compliance / permissions / ecosystem-VMP）由 `arkcli doctor account` 输出消费，并在前置依赖闸门时回看 [`arkcli-shared`](../arkcli-shared/SKILL.md) / [`arkcli-auth`](../arkcli-auth/SKILL.md) | 没权限 / 账号被冻结 / 子账号没 Ark 权限 / VMP 跨服务授权 / 实名 / 欠费       |
| `infer-endpoint`      | [`references/scope-infer-endpoint.md`](references/scope-infer-endpoint.md)                     | `ep-xxx` 报 429 / 慢 / 状态异常 / 报某错误码 / 想看用量与配额压力           |
| `model`               | [`references/scope-model.md`](references/scope-model.md)                                       | 跨接入点看模型整体；模型级配额压力；按模态自适应的延迟与生成时长指标          |
| 默认（CLI 健康）      | 暂无独立 reference；直接读 `arkcli doctor` 输出 + 走 `arkcli-shared` 故障分流                     | 刚装 arkcli 能用吗 / 突然跑不通                                              |

## 路由到错误码 reference

所有错误码诊断细则集中在一个文件：[`references/error-codes.md`](references/error-codes.md)。`arkcli doctor error <code>` 返回的 JSON 里 `reference` 字段统一指向该文件，`subtype` 字段告诉你跳到哪个小节。

### 完整 subtype 路由

| 业务领域       | subtype                  | 跳节                                                                                                  |
|----------------|--------------------------|------------------------------------------------------------------------------------------------------|
| 方舟错误码     | `input_real_face`        | [`error-codes.md` §1.1.1](references/error-codes.md#111-input_real_face)                              |
| 方舟错误码     | `input_copyright`        | [`error-codes.md` §1.1.2](references/error-codes.md#112-input_copyright)                              |
| 方舟错误码     | `input_content_safety`   | [`error-codes.md` §1.1.3](references/error-codes.md#113-input_content_safety)                         |
| 方舟错误码     | `output_video_copyright` | [`error-codes.md` §1.1.4](references/error-codes.md#114-output_video_copyright)                       |
| 方舟错误码     | `output_video_safety`    | [`error-codes.md` §1.1.5](references/error-codes.md#115-output_video_safety)                          |
| 方舟错误码     | `model_access_denied`    | [`error-codes.md` §1.2](references/error-codes.md#12-modelaccessdenied)                               |
| 方舟错误码     | `rate_limit_exceeded`    | [`error-codes.md` §1.3](references/error-codes.md#13-ratelimitexceeded)                               |

> 其余 Volc 原生码已在 `error-codes.md` 各 section 索引表里覆盖（含 HTTP / 释义 / 处理方式）。`error-codes.md` 顶部目录是首要导航入口。
>
> doctor 的**前置依赖闸门**（VMP / TLS / TOS 跨服务授权）不是 API 错误码，由 `arkcli doctor account` 输出 + [`arkcli-shared`](../arkcli-shared/SKILL.md) / [`arkcli-auth`](../arkcli-auth/SKILL.md) 联合处理。
>
> 新错误码贡献流程：[`CONTRIBUTING.md`](CONTRIBUTING.md)。

## doctor 输出 schema（按需消费）

doctor 命令家族输出**两套** JSON schema——按命令分。

### A. 各 scope 命令（`arkcli doctor` / `account` / `infer-endpoint` / `model`）

`arkcli doctor <scope> <id> --format json` 返回一份结构化 JSON。**不要做严格 schema 校验**——字段会随版本增减（如 TTFT/TPOT/replicas、request_id 反查），按字段名按需取：

| 顶层字段                     | 含义                                                                          |
|------------------------------|------------------------------------------------------------------------------|
| `scope`                      | `cli` / `account` / `model` / `infer-endpoint`                                |
| `subject`                    | 资源 ID（model name / endpoint id），CLI 与 account 留空                      |
| `window`                     | VMP 时间窗（如 `24h`）                                                        |
| `checks[]`                   | 各 check 项的 raw 结果：`{ id, status: pass/warn/fail/skip, value, message }` |
| `findings[]`                 | 聚合后的问题清单：`{ severity, error_code?, root_cause, evidence }`           |
| `recommended_fixes[]`        | 修复建议：`{ id, kind: skill/url/command, target, reversibility }`            |
| `error.blocking_dependency?` | 前置依赖未通过（如 VMP 三段授权失败），命令终止 + 引导 fix                    |

### B. `arkcli doctor error <code>` 单错误码查表

这条命令是**只读查表**——不需登录、不消耗推理 token——把一个错误码翻译成结构化诊断。返回字段：

| 字段              | 含义                                                                                            |
|-------------------|------------------------------------------------------------------------------------------------|
| `category`        | 顶层类别（如 `policy` / `permission` / `quota` / `infrastructure`）                              |
| `subtype`         | 类别下的子类型（如 `input_real_face` / `model_access_denied` / `rate_limit_exceeded`）            |
| `code`            | 命中的火山真实错误码（原值回显，用于确认输入原码）                                                |
| `root_cause`      | 根因，用户友好的一句话                                                                           |
| `hint`            | 修复指引（可跑的命令，或需用户做的动作）                                                          |
| `rules`           | 触发信号候选；**多个 = 复合错误码，需消歧**（按 reference 里的判定步骤择一）                       |
| `needs_backend`   | 依赖但尚未上线的能力（如 `deface`）——**不要假装能自动修**                                        |
| `skill`           | 固定 `arkcli-doctor`（就是本 skill）                                                              |
| `reference`       | 统一是 `error-codes`（指向 [`references/error-codes.md`](references/error-codes.md)）；用 `subtype` / `code` 定位具体小节 |

**给用户看时**：摘 `findings` + `recommended_fixes`（scope 命令）或 `root_cause` + reference 段里的修复方案（error 命令），**不要复述完整 JSON**；建议执行需用户二次确认。

### 聚合 vs 原始时序

doctor 默认输出的是**聚合统计值**（如 `error_rate`、`ttft.p99`、`tpm_peak`）——一个数概括一段时间，足够判断"健康吗 / 配额够吗"。

底层数据来自 VMP 的 Prometheus 时序流，必要时可拿**原始时序**（按 step 取点，通常 1min 粒度，通过 `--series` 类 flag 触发；具体 flag 以 doctor 实现为准）做更深分析——聚合**会平均掉突发与抖动**，回答不了"是不是某一刻爆发的 / 是稳定高还是脉冲撞顶"。

| 用户问什么                          | 用哪种            |
|------------------------------------|-------------------|
| "现在怎么样 / 健康吗"              | 聚合（默认）      |
| "什么时候开始 / 持续多久 / 突发还是稳态" | 原始时序        |

详细决策规则与每类指标的时序分析方法见 [`references/scope-infer-endpoint.md`](references/scope-infer-endpoint.md) 的"聚合统计 vs 原始时序"段。

## 安全与边界

- **只读优先**：`arkcli doctor` 默认只读体检，不消耗推理 token、不改任何资源。
- **`--fix` 必先 dry-run**：fixer 分两型——A 型（skill-driven，本质是路由到 reference 让 Agent 跟着步骤走，doctor 进程不 mutate）；B 型（code-driven，少量场景如凭证注入）。两型都必须先看 dry-run、用户确认后再 apply。
- **越权边界**：所有外部调用都用用户 AK/SK 签名（Ark API / VMP / IAM）；不读平台内部数据（k8s / 调度器等）；不跨账号横比。
- **不替用户烧 token**：诊断本身不烧 token，但修复方案如果是重新 `+gen` / `+chat`，先把命令给用户看、等确认。
- **不假装能修没上线的能力**：`needs_backend` 字段非空（如 `deface`）时如实告诉用户该能力尚未上线，给可行替代。

## VMP 前置检测与 `--auto-bind`

`arkcli doctor metrics` / `model` / `infer-endpoint` 三条命令都依赖 VMP（托管 Prometheus）数据，调命令前会按顺序自动检测三件事：

1. **VMP 订阅**（`GetSubscription`）—— 账号是否已开通 VMP；
2. **跨服务授权 SLR**（IAM `CheckServiceLinkedRole`，`ServiceName=ark`）—— 是否已为 ark 建过服务关联角色；
3. **Telemetry 绑定**（Ark `ListTelemetryConfigs`）—— 是否已把某个 VMP workspace 绑到 ark observability。

任何一步没通过，命令直接报 `*output.ExitError` 并附 `LinkOpenMgmtCloudProduct` 跳转链接（云产品开通页 + cloudProduct 抽屉），不会替用户开通订阅或建 SLR——这两步必须在控制台勾「同意条款」。

只有"前两步通过、第三步缺绑"这一种情形可以让 CLI 自动收尾：

```bash
# 默认行为：未绑定时直接报错并提示加 --auto-bind
arkcli doctor metrics request.qpm

# 自动建 ark_default workspace（vmp.standard.15d / 防误删）+ 调
# CreateTelemetryConfig 绑到 ark observability
arkcli doctor metrics request.qpm --auto-bind
arkcli doctor model <model-name>          --auto-bind
arkcli doctor infer-endpoint <endpoint-id> --auto-bind
```

行为细节：
- 账号下已有任意 workspace → 复用第一个；没有 → 建 `ark_default`；
- `metrics` 子命令显式传 `--workspace-id <uuid>` 时跳过 precheck，尊重多账号 / 联调场景。

什么时候建议加 `--auto-bind`：
- agent 自动化路径（确认账号已订阅 + 已建 SLR），让命令幂等地把 telemetry 绑好；
- 用户已在控制台勾过条款、只差最后一步建 + 绑 workspace。

什么时候**不**加：
- 账号还没订阅 VMP / 没建 SLR——加了也没用，CLI 仍只会给跳转链接；先去控制台。

## 配额压力阈值（默认）

doctor 对 RPM/TPM 配额压力的阈值是写死的（用户后续可配置）：

| 占比      | severity | 表现                                |
|-----------|----------|-------------------------------------|
| `≥95%`    | fail     | 已经/即将限流，立即提配额           |
| `≥80%`    | warn     | 高压，建议提配额或调流量分布         |
| `≥50%`    | info     | 健康但有空间                         |
| `<50%`    | pass     | 充裕                                 |

错误率默认阈值 `5%` 以上 warn，具体 scope 的 reference 可在自己的细则里覆盖。

## 何时 _不_ 用本 skill

- 用户问 arkcli 怎么装 / 怎么登录 / profile 怎么切 → [`arkcli-shared`](../arkcli-shared/SKILL.md) / [`arkcli-auth`](../arkcli-auth/SKILL.md)
- 用户想**部署**新 endpoint → [`arkcli-deploy`](../arkcli-deploy/SKILL.md)（不是诊断）
- 用户想**列/启停/更新** endpoint → [`arkcli-infer-endpoint`](../arkcli-infer-endpoint/SKILL.md)
- 用户想看**用量明细**（不带『为什么/异常』判断） → [`arkcli-usage`](../arkcli-usage/SKILL.md)
- 用户想看**模型基础元信息**（参数 / 价格 / 支持参数） → [`arkcli-models`](../arkcli-models/SKILL.md)

## 参考

- [`references/error-codes.md`](references/error-codes.md) — 错误码总册（生视频 + 通用 Ark + 公共码）
- [`references/scope-infer-endpoint.md`](references/scope-infer-endpoint.md) — 单接入点诊断细则
- [`references/scope-model.md`](references/scope-model.md) — 单模型诊断细则
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — 加错误码 / scope check / fixer 的贡献入口
- [arkcli-shared](../arkcli-shared/SKILL.md) — 认证 / 命令选择 / 安全规则
- [arkcli-infer-endpoint](../arkcli-infer-endpoint/SKILL.md) — endpoint CRUD（doctor 不重复实现）
- [arkcli-models](../arkcli-models/SKILL.md) — 公共基础模型查询

