# `arkcli doctor model`（单模型诊断）

> 这是 [`arkcli-doctor`](../SKILL.md) 的 **model domain reference**。当用户/Agent 拿到一个**模型名**（如 `doubao-pro-32k` / `doubao-seedance-1-0-pro`），问『这模型整体怎么样 / 限流是模型级还是接入点级 / 哪个 endpoint 把模型配额吃了 / 模型级 RPM/TPM 配额还有多少 / 生视频的排队压力 / 生视频生成耗时分布 / LLM 首 token 慢吗 / 模型在跨接入点维度上整体出错严重吗』时读这里。
>
> **CRITICAL — 开始前 MUST 先用 Read 工具读取**：
> - [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md)（认证闸门、命令选择顺序）
> - [`../SKILL.md`](../SKILL.md)（doctor 总入口与路径决策表）

## 它解决什么

`arkcli doctor model <model-name>` 是把"这个模型在我账号下整体什么情况"答出来的命令——modality、endpoint 总数与状态、用量、错误率、模型级配额压力、按模态自适应的延迟与生成时长指标，一屏看清。**关键定位**：跨接入点的聚合视角，不是单 endpoint。

| 用户场景                                              | Agent 跑什么                                                                 |
|------------------------------------------------------|------------------------------------------------------------------------------|
| 「`doubao-pro-32k` 整体怎么样」                       | `arkcli doctor model doubao-pro-32k`                                         |
| 「我的生视频模型卡了」                                | `arkcli doctor model <video-gen-name>`（重点看 queue_time + gen_time + 配额）|
| 「为什么模型报 `ContentRiskBlocked` / 大量 ModelAccessDenied」 | 同上 + 按 top error code 加载错误码 reference                       |
| 「seed-2.0 模型限流是多少 / 怎么提配额」              | `arkcli doctor model <name>` + 提配额引导                                    |
| 「哪个 endpoint 把这个模型的配额吃光了」              | `arkcli doctor model <name>` 看 top endpoint by usage                        |
| 「跨接入点对比错误情况」                              | 同上看 top endpoint by error rate                                           |

## 边界：不做什么

- 不替代单接入点级诊断（"`ep-xxx` 为什么 429"）—— 走 [`scope-infer-endpoint.md`](scope-infer-endpoint.md)
- 不替代单条请求 trace（暂无 request 粒度反查）
- 不读跨账号大盘 / 不跨账号横比
- 不替代用户 SLO 评价（"50ms 算不算慢"由用户自己设 target）
- 不做效果评估和改进建议（"我的模型效果如何 / 怎么改进"——本 reference 只看运行健康度）

## 第一步：跑诊断

```bash
arkcli doctor model <model-name>
# 自定义时间窗（默认 24h）：
arkcli doctor model <model-name> --window 1h
# JSON 输出：
arkcli doctor model <model-name> --format json
# 触发 fixer（必先 dry-run）：
arkcli doctor model <model-name> --fix --dry-run
```

> 调用决策树（doctor 内部）：
> ```
> model exists?
>   ↓ no  → fail，让用户检查模型名拼写或换 profile region
>   ↓ yes  → 拿到 modality（LLM / 生图 / 生视频 / ...）
> VMP preflight 通过?
>   ↓ no  → blocking_dependency.vmp.* → 引导 account scope 修 VMP 三段
>   ↓ yes
> 查 VMP（按 modality 选 PromQL）：
>   通用：error_count + count_total + qps_peak + tpm_peak (或 ipm_peak)
>   modality-specific：见下表
> → doctor 侧算 error_rate + 配额压力 + top endpoint → 错误码 enrich + 输出
> ```

> 如果 VMP preflight 失败（doctor 输出 `error.blocking_dependency.vmp.*`），先去修账号级的 VMP 三段授权——由 `arkcli doctor account` 输出引导，配合 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) / [`../../arkcli-auth/SKILL.md`](../../arkcli-auth/SKILL.md) 处理。VMP 没通的时候模型性能段拿不到数据。

## 第二步：读输出（6 段）

### ① 模型基础信息

| Check ID                | 看什么                               | 解读                                                             |
|-------------------------|--------------------------------------|------------------------------------------------------------------|
| `model.exists`          | 模型存在 + **modality 识别**          | modality 决定后面看哪些专属指标——必读，没有这个其它都是无用功     |
| `model.source`          | 平台基础模型 / 自传 / 精调            | 自传或精调（`cm-xxx`）走的代码路径与基础模型略有不同，问题排查时区分 |

> 不存在 → fail。让用户核对模型名（`arkcli models list` / `arkcli custommodel list`）或换 profile region。

### ② 接入点分布

| Check ID                                | 看什么                                                                   |
|-----------------------------------------|--------------------------------------------------------------------------|
| `model.endpoint_total`                  | 当前模型下接入点总量                                                     |
| `model.endpoint_state_distribution`     | 状态分布占比：`Running` / `Failed` / `Pending` / ...                     |
| `model.top_endpoint_by_usage`           | Top 5 接入点按调用次数排（用量"吃料"分析）                              |
| `model.top_endpoint_by_error_rate`      | Top 5 接入点按错误率排（错误集中分析）                                  |

**用 top endpoint 干嘛**：

- **用量集中**：一个 endpoint 占了模型用量 70%+ → 配额压力主要由它产生，提配额前看是否能客户端分流
- **错误集中**：一个 endpoint 错误率远高于其它 → 不是模型问题，是该 endpoint 的配置 / 上游问题，跳到 [`scope-infer-endpoint.md`](scope-infer-endpoint.md) 单查那个 ep
- **均匀分布**：所有 endpoint 错误率都高 → 模型级问题（账号未开通 / 模型配额撞顶 / 平台侧故障）

### ③ 用量统计

| Check ID                  | 看什么                                          |
|---------------------------|------------------------------------------------|
| `model.usage`             | 调用次数 + token 用量（input / output 分开），1h 与 24h 双窗口 |

仅展示，不告警。但与配额压力（§⑤）、错误率（§④）联动判断"是不是真的需要提配额"。

### ④ 错误率与错误码分布

| Check ID                  | 看什么                                                                     |
|---------------------------|---------------------------------------------------------------------------|
| `model.error_rate`        | 错误率（按模态分流，与 `quota_pressure` 同款思路；默认阈值 5% warn）。LLM (`base_model` 维度) / 生图 (`base_model` 维度，注意 `model_service_type` 是服务类型 i2i/t2v/r2v 等不是模型名) → `overall.rate_percent`；生视频 → `request.rate_percent`（CreateContentGenTask 调用，`ark_model` 维度）+ `task.rate_percent`（异步任务执行，`foundation_model_name` 维度，从成功率取 `1 - rate`）|
| `model.errors`            | 错误码分布（按模态分流，与 `error_rate` 同款思路）。LLM/生图 → `errors.overall.by_error_code[]`（`code` + `http_status_code` 拆桶；LLM 用 `code` 标签，生图用 `volc_error_code`，统一成 `code` 字段）；生视频 → `errors.task.by_error_code[]`（仅任务执行失败，无 `http_status_code` 维度，tmp.md 没给请求侧分布 PromQL） |

主导错误码 → 加载对应错误码 reference（[`error-codes.md`](error-codes.md)）。常见路由：

| 主导错误码                                     | 加载                                                                                                                |
|-----------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `ModelAccessDenied` / `ModelNotOpen` / `AccessDenied` | [`error-codes.md` §1.2](error-codes.md#12-modelaccessdenied)                                                  |
| `RateLimitExceeded.*` / `ModelAccount{Rpm,Tpm,Ipm}…` | [`error-codes.md` §1.3](error-codes.md#13-ratelimitexceeded)                                                  |
| 生视频 `Output*SensitiveContentDetected*`       | [`error-codes.md` §1.1.4 / §1.1.5](error-codes.md#11-生视频特有seedance)                                            |
| 其它                                          | `arkcli doctor error <code>` 查表，按返回的 `subtype` 跳节                                                          |

### ⑤ 模型级配额压力

| Check ID                          | 看什么                                                                              |
|-----------------------------------|------------------------------------------------------------------------------------|
| `model.account_quota_pressure`    | 账号下该模型的 RPM / TPM / IPM 实际峰值 vs 配额上限                                  |

**阈值**（与 [`../SKILL.md`](../SKILL.md) "配额压力阈值"段一致）：

| 占比      | severity | 表现                                |
|-----------|----------|-------------------------------------|
| `≥95%`    | fail     | 已经/即将限流，立即提配额           |
| `≥80%`    | warn     | 高压，建议提配额或调流量分布         |
| `≥50%`    | info     | 健康但有空间                         |
| `<50%`    | pass     | 充裕                                 |

**模型级 vs 接入点级判定**（与 §② top endpoint 联动）：

```
模型级配额压力高（≥80%）?
 ├─ 是 ──► 哪些 endpoint 吃料?
 │         ├─ 一两个 endpoint 占大头 → 客户端 sharding，把流量分给其它 endpoint；分不开再提配额
 │         └─ 流量均匀 ──► 真的需要提配额
 └─ 否 ──► 限流是单 endpoint 自己的事，跳到 doctor infer-endpoint 看那个 ep
```

### ⑥ 模态专属指标（按 § ① modality 自适应）

doctor 拿到 modality 后追加专属 check 段。**首先看 § ① 的 modality 字段**，然后只关注那个模态的指标：

| Modality   | 主指标                                                          | 看什么                                                  | 用于                                          |
|------------|----------------------------------------------------------------|---------------------------------------------------------|-----------------------------------------------|
| **LLM**    | `model.ttft.{p50, p99}`                                         | 首 token 延迟分位                                       | 流式对话 UX 核心                              |
|            | `model.tpot.{p50, p99}`                                         | 每 token 延迟分位                                       | 长文输出体验                                  |
| **生图**   | `model.gen_time.{avg, p99}`                                     | 平均生成时间 / p99（无 TTFT/TPOT 概念）                 | 生图主指标——慢就是这个数大                    |
| **生视频** | `model.queue_time.{avg, p99}`                                   | 平均排队时长 / p99                                       | **生视频独有**：GPU/TPU 紧张时排队是主延迟    |
|            | `model.gen_time.{avg, p99}`                                     | 排队后实际生成耗时                                       | 与 queue_time 拼起来才是端到端                |
|            | `model.video_duration_distribution`                             | 生成时长分布（5s / 10s / 30s 等）                       | 计费 + 业务行为画像                           |

**怎么解读**：

- **LLM `ttft.p99` 慢**：通常与 QPS 同步上升 = 拥塞；与 QPS 无关的尖峰 = 上游 / 下游某次抖动；持续偏高 = 模型 / 版本变化
- **LLM `tpot.p99` 慢**：长文体验差，用户感觉"打字越来越慢"
- **生图 `gen_time.p99` 慢**：生成本身慢；看是否所有 endpoint 都慢（模型问题）还是个别 endpoint（resource starve）
- **生视频 `queue_time.p99` 高 + `gen_time.p99` 正常**：GPU 紧张，排队是主延迟。提配额或换时段
- **生视频 `gen_time.p99` 高但 `queue_time` 正常**：模型 / 输入复杂度变了（如分辨率提高 / 视频时长变长）
- **生视频 `video_duration_distribution`**：业务结构画像。30s 视频占比高 → 计费 + GPU 占用都重；可与用户讨论是否切短

> doctor 输出里如果某段缺失（比如 ttft / tpot 当前未提供），如实告诉用户该指标暂不可用，**不要假装能给数字**。

## 第三步：聚合 vs 原始时序（跨 endpoint 维度）

doctor 默认输出聚合统计值；底层是 VMP Prometheus 时序流，必要时拿原始时序做更深分析。**模型级**的原始时序与单 endpoint 的有差别——它聚合了所有 endpoint 的流量，能告诉你**跨 endpoint 的时间格局**。

> 拿原始时序：`arkcli doctor model <name> --series`（实际 flag 名以 doctor 实现定稿为准）。返回 `series.<metric>[]` 字段，每条 `{ts, value, labels}`，`labels` 含 `endpoint_id` 维度。

### 决策规则

| 用户问什么                                              | 用哪种            | 原因                                                         |
|--------------------------------------------------------|-------------------|--------------------------------------------------------------|
| 「模型整体怎么样」                                     | 聚合（默认）      | 一个数判断"要不要进一步查"                                  |
| 「错误是哪个时段集中爆发的」                            | 原始时序          | 聚合把尖峰平均掉了                                          |
| 「`<endpoint>` 是从什么时候开始吃模型配额的」          | 原始时序（按 endpoint 维度） | 看具体 endpoint 流量曲线 vs 总量                          |
| 「模型升级 / 切换流量后效果变化」                       | 原始时序          | 与变更时间点对齐找因果                                       |
| 「业务高峰时段是什么时候 / 是否周期性」                 | 原始时序          | 看曲线形状判断是否周期、是否需要提前扩容                    |

### 跨 endpoint 时序模式 → 推断

| 模式                                                                   | 推断                                                                         |
|------------------------------------------------------------------------|------------------------------------------------------------------------------|
| 一个 endpoint 流量稳步上升、其它平稳                                  | 业务方有新功能上线带量；如果模型级配额因此撞顶，提前规划                     |
| 多个 endpoint 同时段错误率上升                                          | 模型 / 平台侧问题（不是 endpoint 配置问题）                                  |
| 一个 endpoint 错误率周期性高、其它正常                                  | 该 endpoint 的客户端有定时任务或异常重试                                    |
| `queue_time` 跨所有 endpoint 同步上升                                   | GPU 池资源紧张（所有 endpoint 共用）；提配额或换时段                        |
| `tpm_peak` 长期 50-70% 偶尔顶 95%                                       | 业务有脉冲；副本数够但单副本会撞顶                                          |

> 给用户的输出里**不要直接贴一长串原始时序数据**。多模态 agent 应该消化时序后给一句结论性描述（"错误集中在 14:05-14:08，主要由 endpoint `ep-xxx` 贡献"），并附前后聚合对比。

## 常见 Finding → 修复路径

| Finding                                                                  | severity | 修复方向                                                                                |
|--------------------------------------------------------------------------|----------|------------------------------------------------------------------------------------------|
| 模型不存在 / `model.exists = fail`                                       | fail     | 让用户核对模型名 / profile region                                                       |
| 模型级 RPM/TPM/IPM ≥95%                                                  | fail     | 见 [`error-codes.md` §1.3](error-codes.md#13-ratelimitexceeded) 修复 A：提配额          |
| 模型级配额 80-95% + 流量集中在 1-2 个 endpoint                            | warn     | 客户端 sharding 分流；不必立即提配额                                                    |
| 主导错误码 `ModelAccessDenied`                                           | fail     | 加载 [`error-codes.md` §1.2](error-codes.md#12-modelaccessdenied)：账号未开通该模型 / 子账号缺权限 |
| 主导错误码 `ContentRiskBlocked` / `Output*SensitiveContentDetected*`     | warn     | 加载 [`error-codes.md` §1.1.x](error-codes.md#11-生视频特有seedance)（生视频专属）       |
| 一个 endpoint 错误率远高于其它（>10x）                                    | warn     | 跳 [`scope-infer-endpoint.md`](scope-infer-endpoint.md) 单查那个 ep                     |
| 生视频 `queue_time.p99` 高 + `gen_time.p99` 正常                          | warn     | GPU 紧张：提配额或换时段；客户端可加重试退避                                            |
| 生视频 `gen_time.p99` 持续偏高 + 业务无变化                                | info     | 模型 / 版本变化或输入复杂度上升；与业务方核对近期 prompt / 视频时长是否调整              |
| LLM `ttft.p99` 持续偏高                                                  | info     | 当前未提供延迟指标时不要假装；如指标提供，看是否与 QPS 同步                              |

## 给用户看的话术

读完输出后按以下顺序：

1. **整体结论一句话**：modality + 总用量 + 错误率水平 + 配额压力等级
2. **关键问题**：摘最关键的 finding（fail 优先；多 fail 按 error_code Top 排），并附**一条**已加载好的 error reference 里的修复方案
3. **跨 endpoint 上下文**：top endpoint by usage / by error rate 给一行——是不是某个 endpoint 主导了模型级问题
4. **下一步建议命令**：执行类（重新 deploy / `+gen` 重试）的命令**先给用户看、等确认**，不替按

## 安全与边界

- **只读优先**：`arkcli doctor model` 默认只读，不消耗推理 token、不重新发请求
- **`--fix` 大多是 A 型**：注册的多是 skill-driven fixer（如提配额工单引导），doctor 进程不直接调 IAM / 改控制面
- **越权边界**：只读用户 AK/SK 能拿到的 Ark API + VMP 数据；不读平台内部 trace / 调度数据；不跨账号横比

## 何时 _不_ 用本 reference

- 单接入点报错 / 慢 / 状态异常 → [`scope-infer-endpoint.md`](scope-infer-endpoint.md)
- 账号级问题（实名 / 欠费 / IAM 策略 / VMP 三段）→ `arkcli doctor account` + [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md)
- 想看用量明细（不带"为什么 / 异常"判断）→ [`../../arkcli-usage/SKILL.md`](../../arkcli-usage/SKILL.md)
- 想看模型基础元信息（参数 / 价格 / 支持参数）→ [`../../arkcli-models/SKILL.md`](../../arkcli-models/SKILL.md)（公共模型）
- 想给模型提升配额（明确意图 + 知道要多少）→ 直接走控制台工单流程；本 reference 只诊断 + 引导

## 参考

- [`../SKILL.md`](../SKILL.md) — 总入口与路径决策
- [`scope-infer-endpoint.md`](scope-infer-endpoint.md) — 单接入点诊断（联动看具体 ep）
- [`error-codes.md`](error-codes.md) — 错误码总册（含 §1.2 ModelAccessDenied / §1.3 RateLimitExceeded / §1.1 生视频）
- [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) — 共享执行协议
- [`../../arkcli-models/SKILL.md`](../../arkcli-models/SKILL.md) — 模型元信息查询
