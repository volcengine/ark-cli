# RL 插件

实现或审查 rollout、grader、自定义 pipeline、tracing，以及把推理 Agent 改造成训练 Agent 时读取本文件。

## 核心概念

- Rollout plugin：执行模型、工具或 Agent 交互，并产出轨迹。
- Grader plugin：对一条或多条轨迹评分，返回 reward 和可选指标。
- Pipeline wrapper：把 rollout 和 grader 接入 `custom_rl_pipeline`，并可配置环境变量和权重。

选择装饰器前，检查当前 `ark_sdk.resources.pipeline_plugin`。较新的 SDK 可能暴露 `rollout`、`single_rollout`、`group_rollout`、`single_grader`、`group_grader` 和多个 pipeline class，但具体用法以生成模板为准。

## Rollout 职责

Rollout 应该：

- 按需读取 `sample.messages`、`sample.tools` 和 `sample.extra`
- 构造有边界的模型、工具或 Agent 交互
- 调用外部 API 或工具时设置超时和重试策略
- 发生 tool call 时把工具结果回填到 messages
- 设置最大步数并确保循环可退出
- 在样本需要重试、丢弃、失败或向 grader 传递额外信息时，返回 SDK 约定的状态或结果

训练感知客户端很关键。生成模板通常会包装客户端，让 SDK 能跟踪 completion 状态和轨迹：

```python
client = Ark(base_url=proxy.url, api_key=proxy.jwt_token)
wrap_inplace_trace(client)
wrap_client_inplace(client, proxy=proxy)
```

异步代码应使用异步 client 和模板里的异步包装函数。除非模板明确这么做，否则不要在 `async def` rollout 里调用阻塞式同步 client。

## Grader 职责

Grader 应该：

- 校验评分逻辑需要的 sample 字段
- 对收到的每条 trajectory 给出分数
- 返回形状符合 SDK 预期的 reward 列表
- 在 `metrics` 中补充有排障价值的自定义指标
- 对畸形轨迹做确定性处理
- 避免过于宽松的纯字符串打分，除非任务确实适合这样评估

Reward 设计建议：

- 尽量使用稠密、可解释的 reward，而不是过于稀疏的全对/全错判断。
- 对工具参数格式错误、缺失最终答案、循环不退出、隐藏 fallback 等失败模式明确扣分。
- 防止 reward hacking：不要因为回答提到期望文本就直接给高分，必须判断是否真实满足任务。
- 过程 reward 只有在与最终质量一致时才使用。

## 推理 Agent 到训练 Agent 的改造

改造推理 Agent 时：

- 把静默 fallback 改成明确扣分或明确状态
- 让错误工具参数对 grader 可见
- 限制最大步数和重试次数
- 记录足够状态以复现失败
- 在并发下测试下游 API 的延迟、限流和失败行为

## Runtime 与并发

装饰器可能支持实例规格、超时、并发、replica 或 request concurrency 等 runtime 设置。这些字段会随 SDK 变化，因此应复制当前模板的写法。

正式训练前，用接近计划 batch size 的并发做批量测试。插件过慢时，即使模型配置正确，也会浪费训练资源。

## 可观测性 Hook

可用时使用当前模板的 tracing 工具：

- 在 rollout/grader 函数上使用 `@trace_monitor`
- 对模型 client 使用 `wrap_inplace_trace(...)` 或异步等价函数
- 对工具/API 辅助函数使用 `@observe()`
- 用 logger 记录高层决策、工具调用、校验失败和 reward 汇总

不要记录密钥、原始鉴权 header、API key 或完整敏感 payload。
