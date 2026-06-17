# Skills Layout

`skills/` 采用 domain-driven 的组织方式，每个 skill 对应一个稳定的能力域，并参考 `../cli/skills` 的三层组织方式：

1. `arkcli-shared`：统一的执行协议、认证闸门摘要、命令路由（深度细节下沉到 `arkcli-shared/references/*.md` 与 `arkcli-auth/references/*.md`，按需加载）
2. `arkcli-<capability>/SKILL.md`：能力域的快速决策、核心规则、命令索引
3. `arkcli-<capability>/references/*.md`：具体命令的参数、示例、错误处理，按需读取

```text
skills/
  arkcli-shared/
    SKILL.md
  arkcli-<capability>/
    SKILL.md
    references/
      arkcli-<capability>-<verb>.md
      evals.md
```

当前顶层 skill：

- `arkcli-shared`：共享执行协议正文（认证闸门摘要、命令路由、输出规则、安全规则、二次确认协议）；全局 flags / profile 默认与漂移 / 故障分流等查表类内容在 `arkcli-shared/references/`，身份解析 / AK-SK 边界 / API Key 恢复 / 实名闸门在 `arkcli-auth/references/`
- `arkcli-auth`：认证管理（login / logout / status）
- `arkcli-config`：profile 与本地配置
- `arkcli-models`：模型查询（list / get / search）
- `arkcli-custommodel`：自定义模型仓库管理（list / get / upload / update / delete / quantize）
- `arkcli-chat`：快速对话 / 推理（+chat，Responses API，支持多模态 @file）
- `arkcli-understand`：多模态理解工作流（+understand，12 个任务型 sub-skill × 4 模态，复用 Responses API 引擎）
- `arkcli-gen`：统一生成工作流（+gen，自动等待完成）
- `arkcli-deploy`：一键创建 Endpoint（+deploy）
- `arkcli-code-example`：为模型或接入点生成多语言调用示例代码（+code-example）
- `arkcli-infer-endpoint`：推理接入点管理（create / get / list / start / stop）
- `arkcli-plans`：套餐管理（get / buy / renew / model-list / personal rotate-apikey / team seat-list / team seat-assign / team rotate-apikey）
- `arkcli-usage`：推理用量统计（usage stats）
- `arkcli-train-finetune`：精调任务管理（create / list / status / stop / resume）
- `arkcli-billing`：账单查询（billing list，结算金额 / Token 计费,T+1 财务口径）
- `arkcli-connect`：将 skills 安装到本地 AI Agent（+connect）
- `arkcli-api-explorer`：Raw API 兜底

强制规则：

- 顶层 skill 名必须是稳定能力域，不是 OpenAPI Action 名
- 目录名必须和 `SKILL.md` frontmatter 的 `name` 一致
- 子命令说明、flags、示例统一放在所属 skill 的 `references/`
- reference 命名规则：
  - 具体命令参考：`arkcli-<domain>-<verb>.md`（例如 `arkcli-models-list.md`）
  - 非命令型指南允许固定名：`guide.md` / `evals.md`（用于路由规则、守卫清单、评估用例等）
- 仓库里只能有一个共享 skill：`arkcli-shared`
- 所有非共享 skill 都必须先要求读取 `../arkcli-shared/SKILL.md`
- `arkcli-shared` 必须承担统一的 agent 执行协议入口：认证闸门（摘要 + 指向 arkcli-auth）、命令选择顺序、降级路径入口、写操作安全边界与二次确认；稀有路径 / 查表类细节下沉到 `references/`，正文只保留"几乎每个任务都命中"的规则
- 各业务 skill 只补本能力特有的“快速决策 / Agent 快速执行顺序 / 核心规则”，不要重复抄共享规则
- 若某个能力下已有 `references/*.md`，调用具体命令前必须先读对应 reference，再执行命令
- 业务 skill 需要把用户原始目标串起来：例如先登录、查模型、再回到原任务，而不是停在中间步骤

## 最小评估 / 回归（强制）

每个非共享 skill 必须在 `references/evals.md` 中提供最小评估用例，用于持续验证可用性与回归：

- `trigger`：该唤起时能唤起，并给出正确的命令序列与检查点
- `anti-trigger`：不该唤起时能拒绝进入并正确分流到其他产品命令/skill
- `guard`：认证/配置/高风险写操作时能先做闸门检查或要求确认
- `happy-path`：至少 1 条端到端可重复的命令链路（若需要联网/登录，必须标注前置条件）

评估关注点不是“文案质量”，而是 skill 的：可路由性、可执行性、可诊断性、可回归验证性。

## 评测资产边界

`skills/` 目录只放会随 skill 一起交付、并可能被用户安装到 Agent 的内容：

- `SKILL.md`
- `references/*.md`
- 人类可读的最小回归说明（例如 `references/evals.md`）

不要把仅供 `skill-creator` / CI / benchmark 使用的机器评测资产直接塞进 `skills/`，例如：

- `grading_rules.json`
- `evals_runtime.json`
- benchmark profile / runtime adapter 私有配置

这类资产应放到仓库测试侧（推荐 `tests/skills/<skill>/`），避免污染最终 skill 交付物。

## 路由与守卫（建议，兜底类强制）

每个 skill 的 `SKILL.md` 建议包含（兜底类如 `arkcli-api-explorer` 必须包含）：

- 唤起信号（When To Trigger）
- 反唤起信号（When NOT To Trigger）
- 降级路径（auth/config/other domain）
- Guard Checklist（认证闸门、配置归因、风险确认、噪声控制）
