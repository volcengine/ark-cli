---
name: arkcli-plans
version: 1.2.0
description: "ARK 套餐管理(Agent Plan / Coding Plan,个人版 + 企业版):查询持有 / 购买 / 续费 / 模型清单 / 轮换 APIKey,**以及企业版席位的全部管理操作**:列出席位(`plans team seat-list`)、给员工分配席位(`plans team seat-assign`)、查谁绑了哪个 seat、轮换席位 APIKey(`plans team rotate-apikey`)。命中关键词:套餐 / 买 / 续费 / 列席位 / 看 seat 绑定 / 谁绑了哪个 seat / 给员工分席位 / 解绑 / 团队席位 admin 视图 / 轮换 APIKey。**用量类问题(还剩多少额度 / 用了几成 / 每个 seat 用了多少)走 [arkcli-usage](../arkcli-usage/SKILL.md)。**动词路由:**列 / 绑 / 分 / 轮换 / 解绑** → 这里;**用 / 消耗 / 多少** → arkcli-usage。另含 harness-status:只读查看本机 AI Agent(claude-code/opencode/openclaw/trae)上 Agent Plan 内置 MCP(豆包搜索/dataPro/OpenViking)装没装、key 就没就绪;命中:我的 MCP 装好了吗 / 查 MCP 安装状态 / 看 agent 上的 MCP / harness 状态。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli plans --help"
---

# arkcli plans

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与共享安全规则。**
**CRITICAL — 任何 `plans buy` / `plans renew` / `plans personal rotate-apikey` / `plans team seat-assign` / `plans team rotate-apikey` 在执行之前，务必先用 Read 工具读取对应 `references/*.md`，禁止盲目调用。**

## 🔒 协议闸门（plans buy / plans renew 强制流程）

`plans buy` / `plans renew` 是**计费写操作**，加 `--yes` 才真扣款。**不传 `--yes` 不传 `--dry-run` 时**, CLI 不下单, 而是返回:

```json
{
  "status": "agreement_required",
  "agreements": [{"title": "...", "url": "..."}, ...],
  "next_step": "arkcli plans buy ... --yes"
}
```

**Agent 必须严格按以下顺序执行:**

1. **第一次调用必须不带 `--yes`** — 拿到 `agreements` 数组
2. **逐条**把每条协议的 `title` 和 `url` 展示给最终用户(不能省略,不能合并)
3. 等待用户**明确表达**已阅读并同意全部协议(例如"我已阅读并同意"、"OK 我同意"等清晰表态);**不能**用模糊回答("好"、"嗯") 默认通过
4. 确认后才能加 `--yes` 重跑 `next_step` 字段里的命令(直接复制 next_step 即可,flag 已回显完整)

**违反这条流程 = 帮用户跳过法律合规步骤**。这跟 frontend 购买面板的 "我已阅读并同意《...》" checkbox 等价 — 必须真人看过才能勾。

`--dry-run` 是询价路径,不下单也不需要协议确认 — 但用户最终走 `--yes` 真下单时仍需先看协议。

## 业务定位

`arkcli plans` 管理两类 ARK 套餐：

- **Agent Plan**（agent-plan / agent-plan-team）：智能体调用套餐，按 tier 划档（small/medium/large/max）
- **Coding Plan**（coding-plan / coding-plan-team）：编程辅助套餐，按 tier 划档（lite/pro）

每类套餐都有 **个人版**（personal）与 **企业版 / 团队版**（team）两种形态：

| 形态 | 资源单位 | 典型操作 |
|---|---|---|
| personal | 个人订阅 (subscription) | 查看 / 购买 / 续费 / 轮换专属 APIKey |
| team | 企业席位 (seat) | 查看 / 购买 / 续费 / 列出席位 / 绑定子用户 / 轮换席位 APIKey |

> personal 一个账号下一份订阅；team 是按席位（SeatID）粒度管理，每个席位绑一个子用户。

> **本 skill 不含套餐用量 / 配额查询**（"用了多少 / 还剩多少 / 几号刷新 / 按模型拆分"）。`plans get` 只回答"我**持有**哪些套餐 + 状态(Effective/Running)"，不回答"用了多少"。配额视图在 [`../arkcli-usage/`](../arkcli-usage/SKILL.md) — 详见下方"快速决策"分流表。

## 适用场景

- 查看当前账号下持有哪些套餐 → `plans get`
- 下单买套餐（含个人 / 团队）→ `plans buy`
- 续费已有套餐 → `plans renew`
- 看套餐支持的模型清单 → `plans model-list`
- 个人版轮换 APIKey → `plans personal rotate-apikey`
- 列出 / 筛选企业版席位 → `plans team seat-list`
- 把企业版席位绑给子用户 → `plans team seat-assign`
- 轮换企业版席位的 APIKey（自己 or 管理员批量）→ `plans team rotate-apikey`
- 查本机 AI Agent 上专属 Harness 能力装没装（只读）→ `plans harness-status`：按能力卡报豆包搜索 / 专业数据集 / Agent 记忆（MCP），外加全局**火山引擎Supabase**（CLI+Skill 装没装）。「我的 supabase / MCP 装好了吗」都走这里（Agent 记忆卡仅个人版 `agent-plan`；团队版报 `absent` 是预期、非漏装，详见其 reference）

## 快速决策

- 用户问 "我有什么套餐 / 我订阅了什么"：直接 `arkcli plans get`，零参数（**仅持有列表 + 状态，不含用量**）
- 用户问 **"我用了多少 / 还剩多少 / 几号刷新 / 套餐内 vs 套餐外"** —— **不在本 skill**，转 [`../arkcli-usage/`](../arkcli-usage/SKILL.md)：
  - "我的套餐还剩多少 / 几号刷新" → `arkcli usage plan` 或 `arkcli usage balance --type plan`
  - "我哪个模型用得最多 / 套餐内套餐外比例" → `arkcli usage plan-details`
  - "团队席位的用量" → `arkcli usage seats --product agent-plan-team --with-usage`
- 用户问 "Agent Plan / Coding Plan 支持什么模型"：`arkcli plans model-list --plan <plan>`
- 用户要 "买 / 续 套餐"：先看 [references/arkcli-plans-buy.md](references/arkcli-plans-buy.md) / [renew.md](references/arkcli-plans-renew.md)，**严格要求显式 `--plan`、`--type`、`--duration`、团队版还要 `--quantity`**，不要替用户做选择
- 用户要 "重置 / 轮换 APIKey"：分清个人版还是企业版，参考对应 reference；**写操作，原 APIKey 立即失效**，必须按 reference 走二次确认（除非用户明确要 `--yes` 跳过）
- 用户要 "**查席位 / 看团队 seat 绑定情况 / 谁绑了哪个 seat / 列出席位 / 哪些席位激活了 / 团队席位 admin 视图**" → `plans team seat-list --plan <agent-plan-team|coding-plan-team>`(**这是 seat 管理的默认入口**,管理视角列基础信息 + 绑定关系,**不带用量数字**;要看每个 seat 用了多少 token / 套餐百分比 → `arkcli usage seats --with-usage`)
- 用户要 "把席位分配给员工 / 给员工分配 seat / 解绑席位"→ `plans team seat-assign`,先准备好 `seat-id=user-id` 配对清单
- 用户**只知道员工用户名**不知道精确 UserID：先 `arkcli iam userid --username <prefix>` 反查到 `user_id`，再喂给 `seat-assign --bind seat-id=<user_id>`

## Agent 快速执行顺序

1. 先确认认证状态：`arkcli auth status`；缺失走 `../arkcli-auth/`
2. 读操作（`get` / `model-list` / `seat-list`）直接执行；只在用户问 "我的" 时考虑当前身份
3. **写操作（`buy` / `renew` / `rotate-apikey` / `seat-assign`）** 务必：
   - 先读对应 reference
   - 跟用户确认关键字段（plan / type / duration / SeatIDs / UserID 配对）
   - `rotate-apikey` 默认不传 `--yes`，让 CLI 走 [Y/N] 二次确认
4. 部分失败（per-item Success/Failed 数组）是合法返回；exit code 非零时**不要**直接当全失败，先看 stdout 里的 success_count / failed_count

## 写操作风险清单（必读）

| 命令 | 风险 | 必做项 |
|---|---|---|
| `plans buy` | **计费**，`IsAutoPay=true` 自动扣款 | 必须先不带 `--yes` 走协议闸门 → 把 agreements 展示给用户 → 用户明确同意后加 `--yes`。详见 [协议闸门流程](#-协议闸门plans-buy--plans-renew-强制流程) |
| `plans renew` | **计费**，自动扣款 | 同 buy 协议闸门;团队版必传 `--seat-ids` |
| `plans personal rotate-apikey` | **原 APIKey 立即失效** | 默认走 [Y/N] 确认；提醒用户同步替换 harness 配置 |
| `plans team seat-assign` | 修改席位绑定关系 | 显式 `--bind seat-id=user-id`，自动调 IAM 反查 UserName |
| `plans team rotate-apikey` | **原 APIKey 立即失效** | self-rotate 默认 agent-plan-team；admin batch 通过 `--seat-ids` |

## 命令一览

| 命令 | 类型 | 说明 |
|------|------|------|
| [`plans get`](references/arkcli-plans-get.md) | 读 | 列出当前账号持有的套餐（个人版 + 团队版聚合） |
| [`plans buy`](references/arkcli-plans-buy.md) | 写（计费） | 下单购买套餐 |
| [`plans renew`](references/arkcli-plans-renew.md) | 写（计费） | 续费已有套餐 |
| [`plans model-list`](references/arkcli-plans-model-list.md) | 读 | 列套餐支持的模型 + 当前选中的 ark-latest-model |
| [`plans personal rotate-apikey`](references/arkcli-plans-personal-rotate-apikey.md) | 写（毁坏） | 轮换 Agent Plan 个人版 APIKey |
| [`plans team seat-list`](references/arkcli-plans-team-seat-list.md) | 读 | 列出企业版席位 + 多维度筛选 |
| [`plans team seat-assign`](references/arkcli-plans-team-seat-assign.md) | 写 | 批量绑定企业版席位到子用户 |
| [`plans team rotate-apikey`](references/arkcli-plans-team-rotate-apikey.md) | 写（毁坏） | 轮换企业版席位 APIKey（self / admin batch） |
| [`plans harness-status`](references/arkcli-plans-harness-status.md) | 读 | 查本机 AI Agent(claude-code/opencode/openclaw/trae)上 Agent Plan 内置 MCP 的安装状态(只读、免登录) |

## 常见降级

- 鉴权错误：转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- 区域 / project 不对：转 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)
- **想看套餐用量 / 配额（used / total / percent / reset_at / 按模型拆）**：转 [`../arkcli-usage/SKILL.md`](../arkcli-usage/SKILL.md) — 本 skill 只回答"持有什么"，不回答"用了多少"
- **想看账单 / 结算金额**：转 [`../arkcli-billing/SKILL.md`](../arkcli-billing/SKILL.md) — `plans` 不出钱
- 想生成模型调用样例：转 [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md)
- 试用某个模型不打算正式部署：转 [`../arkcli-chat/SKILL.md`](../arkcli-chat/SKILL.md) / [`../arkcli-gen/SKILL.md`](../arkcli-gen/SKILL.md)
- 没现成产品命令时回退 [`../arkcli-api-explorer/SKILL.md`](../arkcli-api-explorer/SKILL.md)
- 想**安装 / 注入 / 移除** MCP（不是查看状态）：转 [`../arkcli-helper/SKILL.md`](../arkcli-helper/SKILL.md) — `plans harness-status` 只读查看，不改配置

## 参考

- [arkcli-shared](../arkcli-shared/SKILL.md) -- 认证 / 全局参数 / 输出规则
- [arkcli-usage](../arkcli-usage/SKILL.md) -- 套餐**用量 / 配额**视图（`usage plan` / `usage plan-details` / `usage balance` / `usage seats --with-usage`），跟本 skill 的"持有 / 购买 / 席位管理"互补
- [arkcli-billing](../arkcli-billing/SKILL.md) -- 套餐**结算金额**（火山计费中心拆账，T+1）
- [arkcli-deploy](../arkcli-deploy/SKILL.md) -- 套餐买好后用 `+deploy` 部署 endpoint
- [arkcli-helper](../arkcli-helper/SKILL.md) -- 给 AI Agent **注入 / 移除** Agent Plan 内置 MCP（`plans harness-status` 只读查看，注入改动走这里）
