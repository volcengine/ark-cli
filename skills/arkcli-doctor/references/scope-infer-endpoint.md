# `arkcli doctor infer-endpoint`（单接入点诊断）

> 这是 [`arkcli-doctor`](../SKILL.md) 的 **infer-endpoint domain reference**。当用户/Agent 拿到一个 `ep-xxx` 接入点 ID，问『为什么报 429 / 慢 / 不可用 / 状态异常 / ModelAccessDenied / RateLimitExceeded / InvalidParameter / 用量多少 / 配额够不够 / 限流怎么处理』这类**单接入点级**问题时读这里。
>
> **CRITICAL — 开始前 MUST 先用 Read 工具读取**：
> - [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md)（认证闸门、命令选择顺序）
> - [`../SKILL.md`](../SKILL.md)（doctor 总入口与路径决策表）

## 它解决什么

`arkcli doctor infer-endpoint <ep-id>` 是一次性把"这个接入点现在到底什么情况"答出来的命令——状态 + 用量 + 错误率 + 错误码分布 + 配额压力，一屏看清。

| 用户场景                                    | Agent 跑什么                                                                                      |
|--------------------------------------------|---------------------------------------------------------------------------------------------------|
| 「ep-xxx 报 429 怎么办」                   | `arkcli doctor infer-endpoint ep-xxx` + 按错误码 reference 路由                                   |
| 「ep-xxx 报 ModelAccessDenied」            | 同上 + 加载 [`error-codes.md` §1.2](error-codes.md#12-modelaccessdenied)                          |
| 「ep-xxx 慢 / 用着越来越卡」               | `arkcli doctor infer-endpoint ep-xxx --window 1h`                                                |
| 「ep-xxx 状态异常 / pull 不下来 / 挂了」   | 同上（status 异常时直接出 status_reason，不进 VMP，<2s 返回）                                     |
| 「ep-xxx 一个请求为什么失败」              | 暂不支持单条请求级诊断——让用户提供错误码 / request_id / 完整错误 JSON，再走总入口                  |

## 边界：不做什么

- 不查 `sample_request_ids` / 不读 GPU 利用率、HPA 决策链等**平台内部数据**——这些是平台诊断的范畴，用户侧 doctor 拿不到。
- 不替代单条请求 trace（暂无 request 粒度反查）。
- 不替代用户 SLO 评价（"50ms 算不算慢"由用户自己设 target，doctor 给的是分位数事实）。

## 第一步：跑诊断

```bash
arkcli doctor infer-endpoint ep-xxx
# 自定义时间窗（默认 24h）：
arkcli doctor infer-endpoint ep-xxx --window 1h
# JSON 输出：
arkcli doctor infer-endpoint ep-xxx --format json
# 触发 fixer（必先 dry-run）：
arkcli doctor infer-endpoint ep-xxx --fix --dry-run
```

> 调用决策树（doctor 内部，了解即可）：
> ```
> endpoint exists?
>   ↓ no  → fail，让用户检查 ep-id
>   ↓ yes
> status 异常?
>   ↓ yes → 直接出 status_reason + 错误码 enrich，不进 VMP（<2s）
>   ↓ no
> VMP preflight 通过?
>   ↓ no  → blocking_dependency.vmp.* → 引导 account scope 修 VMP 三段
>   ↓ yes
> 查 VMP（错误码分布 + qps_peak + tpm_peak）→ 算错误率 + 配额压力 → 错误码 enrich + 输出
> ```

> 如果 VMP preflight 失败（doctor 输出 `error.blocking_dependency.vmp.*`），先去修账号级的 VMP 三段授权——由 `arkcli doctor account` 输出引导，配合 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) / [`../../arkcli-auth/SKILL.md`](../../arkcli-auth/SKILL.md) 处理。VMP 没通的时候 endpoint 性能段拿不到数据。

## 第二步：读输出（4 段）

### ① Endpoint 基础信息

| Check ID                          | 看什么                                          | 失败/异常时                                                                  |
|-----------------------------------|------------------------------------------------|------------------------------------------------------------------------------|
| `endpoint.exists`                 | endpoint 是否存在                              | 不存在 → 让用户核对 ep-id 或换 profile                                       |
| `endpoint.status`                 | 状态 + 异常原因 (`Running` / `Failed` / ...)   | `Failed` → 直接看 `status_reason`（如 `ImagePullBackOff`），通常是模型版本/镜像/资源问题 |
| `endpoint.model_info`             | 模型名 / 版本 / 模态                            | 模态决定下一步看什么指标（LLM vs 生图 vs 生视频）                            |

> **status fail 时不要继续读后面的指标**——endpoint 都没起来，VMP 上自然没流量数据。直接给用户 status_reason 的解释 + 修复建议（重新 deploy / 换模型版本 / 提工单）。

### ② Performance（VMP 时间窗内）

| Check ID                  | 看什么                                          | 怎么判断                                                                      |
|---------------------------|------------------------------------------------|-------------------------------------------------------------------------------|
| `endpoint.usage`          | 调用次数 / token 用量（输入 + 输出分开）        | 仅展示，不告警                                                                |
| `endpoint.error_rate`     | 错误率（按模态分流，与 `quota_pressure` 同款思路）| 默认阈值 5%，≥5% 标 ⚠。LLM/生图 → `overall.rate_percent`；生视频 → `request.rate_percent`（前端调用错误率）+ `task.rate_percent`（异步执行错误率，doctor 侧从成功率取 `1 - rate`）|
| `endpoint.errors`         | 错误码分布（按模态分流，与 `error_rate` 同款思路）。LLM/生图 → `errors.overall.by_error_code[]`（`code` + `http_status_code` 拆桶；LLM 用 `code` 标签，生图用 `volc_error_code`，统一成 `code` 字段）；生视频 → `errors.task.by_error_code[]`（仅任务执行失败，无 `http_status_code` 维度）；主导错误码 → 加载对应错误码 reference |

**错误码路由**——按 doctor 总入口的 references 表：

| 错误码                                    | 加载 reference                                                                                          |
|-------------------------------------------|---------------------------------------------------------------------------------------------------------|
| `ModelAccessDenied`                       | [`error-codes.md` §1.2](error-codes.md#12-modelaccessdenied)                                            |
| `RateLimitExceeded` / 429                 | [`error-codes.md` §1.3](error-codes.md#13-ratelimitexceeded)                                            |
| 生视频 `Output*SensitiveContentDetected*` 系列 | [`error-codes.md` §1.1](error-codes.md#11-生视频特有seedance) 5 个完整 subtype                        |
| 其它                                      | `arkcli doctor error <code>` 查表，按返回的 `subtype` 跳节                                              |

### 按模态看哪些性能指标

读完 § ① 拿到 `endpoint.model_info` 的 modality 后，**只看该模态相关的延迟 / 生成时长指标**——不同模态的 UX 主体验不同，看错的指标会得出错误结论。

| Modality   | 主指标                                          | 看什么                                                           | 用于                                                          |
|------------|------------------------------------------------|------------------------------------------------------------------|---------------------------------------------------------------|
| **LLM**    | `endpoint.ttft.{p50, p99}`                      | 首 token 延迟分位                                                | 流式对话 UX 核心（"打字开始等多久"）                          |
|            | `endpoint.tpot.{p50, p99}`                      | 每 token 延迟分位                                                | 长文输出体验（"打字越来越慢"）                                |
| **生图**   | `endpoint.gen_time.{avg, p99}`                  | 平均生成时间 / p99（无 TTFT/TPOT 概念）                          | 生图主指标——慢就是这个数大                                  |
| **生视频** | `endpoint.queue_time.{avg, p99}`                | 平均排队时长 / p99                                                | **生视频独有**：GPU/TPU 紧张时排队是主延迟来源                |
|            | `endpoint.gen_time.{avg, p99}`                  | 排队后实际生成耗时                                                | 与 queue_time 拼起来才是端到端                                |
|            | `endpoint.video_duration_distribution`          | 生成时长分布（5s / 10s / 30s 等）                                 | 计费 + 业务行为画像                                            |

> doctor 输出里如果某段缺失（比如 `ttft` / `tpot` 当前未提供），如实告诉用户该指标暂不可用，**不要假装能给数字**。看不到延迟分位时，用户问"为啥慢"只能从错误率 + 配额压力两个角度说。

> 跨接入点看模型整体（"我的 doubao-pro-32k 整体怎么样"）→ 跳 [`scope-model.md`](scope-model.md)：模型级配额 / top endpoint 分析 / 跨 endpoint 时序模式都在那个 reference。

### ③ 配额压力（接入点级 vs 模型级）

| Check ID                       | 看什么                                          | 怎么判断                                                                              |
|--------------------------------|------------------------------------------------|---------------------------------------------------------------------------------------|
| `endpoint.quota_pressure`      | 接入点 RPM/TPM 实际峰值 / 上限                  | 阈值见 [`../SKILL.md`](../SKILL.md) "配额压力阈值"段                                   |

**判断逻辑**：
- 接入点级压力高（≥80%）但模型级低 → **业务流量分布不均**，建议加副本或调流量分配，不必提配额
- 接入点级 + 模型级都高 → **真的需要提配额**，给提配额工单 URL（控制台路径，按 modality 不同）
- 接入点级低、模型级高 → 多半是其它 endpoint 把模型配额吃了，去 `arkcli doctor model <name>` 看跨接入点情况

### ④ Findings + Recommended fixes

doctor 把上面三段聚合成 finding 列表 + 建议修复。给用户时摘**最关键的 1-3 条**，**不要复述完整 JSON**。

## 聚合统计 vs 原始时序：什么时候用哪种

doctor 默认输出的是**聚合统计值**（一个数概括一段时间，如 `error_rate.overall.rate_percent=3.3%`、`ttft.p99=850ms`、`tpm_peak=14.2M`）——足够回答"健康吗 / 配额够吗"。

但底层数据来自 VMP 的 Prometheus 时序流，必要时可以拿**原始时序**（按 `step` 取点，通常 1min 粒度）做更深的分析。聚合是平均，**会平均掉突发与抖动**——单凭聚合值无法回答"是不是某一刻爆发的"。

> **怎么拿原始时序**：用 `--series` flag（实际 flag 名以 doctor 实现定稿为准；语义上是"返回时间序列而非聚合点"）。返回的 JSON 在 `series.<metric>[]` 字段里，每条 `{ts, value, labels}`。

### 决策规则

| 用户问什么                                       | 用哪种            | 原因                                                                       |
|--------------------------------------------------|-------------------|----------------------------------------------------------------------------|
| 「现在/最近怎么样 / 健康吗」                     | 聚合（默认）      | 一个数足以判断"要不要进一步查"                                              |
| 「错误是某个时间突然爆发的吗」                   | 原始时序          | 聚合把尖峰平均掉了；时序能看到时间分布                                      |
| 「流量是匀速还是脉冲」                           | 原始时序          | 看曲线形状判断是否要削峰填谷 / 加副本                                       |
| 「慢的时候是哪几个时刻 / 持续多久」              | 原始时序（含分位）| p99 时序告诉你是持续慢还是周期性慢                                          |
| 「配额压力是稳定高还是偶发顶撞」                 | 原始时序          | `tpm_peak` 一个数看不出"长期 70% 还是只有 1 分钟撞 99%"                     |
| 「发工单要附数据」                               | 聚合 + 原始时序   | 工单要结论也要证据：聚合证明问题，时序证明时间分布                           |
| 「我接入点突然变慢/出错前后做了什么」            | 原始时序          | 与变更/部署时间点对齐，找因果                                                |

> 给用户的输出里**不要直接贴一长串原始时序数据**。多模态 agent 应该消化时序后给一句结论性描述（"错误集中在 14:05-14:08 的 3 分钟里，之后归零"），并附前后聚合对比。

### 每类指标，原始时序能告诉你什么

| 指标                                | 时序模式 → 推断                                                                                                            |
|-------------------------------------|---------------------------------------------------------------------------------------------------------------------------|
| `request.qps` / RPM                 | 平稳 = 业务稳定；尖峰 = 突发流量；阶梯上升 = 业务增长（提前规划提配额）；周期性 = cron / 定时任务                            |
| `error_count` by `error_code`       | 集中在某 1-5 分钟 = 临时故障（可能是平台侧 incident）；均匀分布 = 常态问题，按错误码 enrich 修；与 QPS 同步 = 过载触发      |
| `error_rate`                        | 错误率随 QPS 上升而上升 = 过载 / 配额；错误率与 QPS 无关 = 配置/权限类（如 ModelAccessDenied 跟流量没关系）                |
| `ttft.p99` / `tpot.p99`             | 与 QPS 同步上升 = 拥塞；与 QPS 无关的尖峰 = 上游/下游某次抖动；持续偏高 = 模型/版本变化（**当前未提供**）                            |
| `gen_time.p99` / `queue_time.p99`   | `queue_time` 拉高 = GPU 紧张（提配额或换时段）；`gen_time` 拉高但 `queue` 不变 = 模型 / 输入复杂度变了                      |
| `tpm_peak` / `qps_peak`             | 长期 50-70% 偶尔顶 95% = 业务有脉冲，副本数够但单副本会撞顶；持续 90%+ = 真的需要提配额                                    |

### 原始时序分析的 anti-pattern

- **只看一两个点**：时序至少看一个完整周期（业务上是 1h / 1d，按场景）
- **用绝对值不看上下文**：「错误数 100」是好是坏取决于总 QPS，错误率才是横向可比的
- **把瞬时 spike 当趋势**：5 分钟一次的尖峰可能只是定时任务或客户端 cron
- **不对齐时间窗**：error_rate vs QPS 比对必须用同一时间窗、同一 step

### 当前局限（要诚实告诉用户）

- **当前未提供 `latency.ttft.p99` / `tpot.p99`**。用户问"为什么慢"时只能从错误率 + 配额压力两个角度说，**不要假装能给延迟分位**。
- **暂不支持 `request_id` 反查**。用户拿一条具体请求来问，先追问错误码或完整错误 JSON，再走错误码路由。
- **原始时序数据是用户 AK/SK 能看到的范围**，不要尝试关联平台内部 trace（k8s pod 状态等）——那是平台诊断的事。

## 常见 Finding → 修复路径

| Finding                                              | severity | 修复方向                                                                          |
|------------------------------------------------------|----------|------------------------------------------------------------------------------------|
| 状态 `Failed: ImagePullBackOff`                      | fail     | 重新 deploy / 换模型版本；持续失败提工单                                          |
| `ModelAccessDenied` 占比高                           | fail     | 加载 [`error-codes.md` §1.2](error-codes.md#12-modelaccessdenied)：账号未开通该模型 / 子账号缺权限 |
| `RateLimitExceeded` 占比高 + 接入点级 TPM ≥95%       | fail     | 提配额（控制台）+ QoS 调整                                                         |
| 总错误率 ≥5% 但无主导错误码                          | warn     | 错误码分布尾部、定位不到单一根因——给用户做问卷或 sample 错误日志                  |
| 接入点级 TPM 80-95% / 模型级 < 50%                   | warn     | 流量分布不均；不必提配额                                                           |

## 给用户看的话术

读完输出后按以下顺序：

1. **状态结论一句话**：endpoint Running / Failed / 受影响范围。
2. **关键问题**：摘最关键的 finding（fail 优先；多 fail 按 error_code Top 排），并附**一条**已加载好的 error reference 里的修复方案。
3. **配额上下文**：如果配额压力 warn，告诉用户是接入点级还是模型级，建议是"调分布"还是"提配额"。
4. **下一步建议命令**：执行类（重新 deploy / `+gen` 重试）的命令**先给用户看、等确认**，不替按。

## 安全与边界

- **只读优先**：`arkcli doctor infer-endpoint` 默认只读，不消耗推理 token、不重新发请求。
- **`--fix` 大多是 A 型**：当前注册的几乎都是 skill-driven fixer（如 `fix-model-access-grant` 引导用户控制台开通），doctor 进程不直接调 IAM / 改控制面。
- **越权边界**：只读用户 AK/SK 能拿到的 Ark API + VMP 数据；不读平台内部 trace/调度数据。

## 何时 _不_ 用本 reference

- 想**创建** endpoint → 走部署 skill（一键）或 [`../../arkcli-infer-endpoint/SKILL.md`](../../arkcli-infer-endpoint/SKILL.md)（细控）
- 想**列 / 启 / 停 / 更新** endpoint → [`../../arkcli-infer-endpoint/SKILL.md`](../../arkcli-infer-endpoint/SKILL.md)
- 跨接入点看模型整体（"我的 doubao-pro-32k 整体怎么样"）→ [`scope-model.md`](scope-model.md)
- 账号级问题（实名 / 欠费 / IAM 策略 / VMP 三段）→ `arkcli doctor account` + [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md)
- 想看用量明细（不带"为什么/异常"判断）→ [`../../arkcli-usage/SKILL.md`](../../arkcli-usage/SKILL.md)

## 参考

- [`../SKILL.md`](../SKILL.md) — 总入口与路径决策
- [`error-codes.md`](error-codes.md) — 错误码总册（含 §1.2 ModelAccessDenied / §1.3 RateLimitExceeded / §1.1 生视频）
- [`scope-model.md`](scope-model.md) — 单模型诊断（联动看跨接入点）
- [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) — 共享执行协议
- [`../../arkcli-infer-endpoint/SKILL.md`](../../arkcli-infer-endpoint/SKILL.md) — endpoint CRUD（doctor 不重复实现）
