---
name: arkcli-plans
version: 1.2.4
description: "ARK 套餐管理(Agent Plan / Coding Plan,个人版 + 企业版):查询持有 / 购买 / 续费 / 模型清单 / 轮换 APIKey,**以及企业版席位的全部管理操作**:列出席位(`plans team seat-list`)、给员工分配席位(`plans team seat-assign`)、查谁绑了哪个 seat、轮换席位 APIKey(`plans team rotate-apikey`)。也负责购买参数的逐步澄清，以及 `payment_failed` + Order ID 后回到已有订单补单/取消，禁止重新 buy。命中关键词:套餐 / 买 / 续费 / 列席位 / 看 seat 绑定 / 谁绑了哪个 seat / 给员工分席位 / 解绑 / 团队席位 admin 视图 / 轮换 APIKey。反触发：Token 额度包/资源包不是 Agent Plan/Coding Plan，应走 arkcli-profile 判断 platform 切面。**用量类问题(还剩多少额度 / 用了几成 / 每个 seat 用了多少)走 [arkcli-usage](../arkcli-usage/SKILL.md)。**动词路由:**列 / 绑 / 分 / 轮换 / 解绑** → 这里;**用 / 消耗 / 多少** → arkcli-usage。另含 harness-status:只读查看本机 AI Agent(claude-code/codex/opencode/openclaw/trae)上 Agent Plan 内置 MCP(豆包搜索/dataPro/OpenViking)装没装、key 就没就绪;命中:我的 MCP 装好了吗 / 查 MCP 安装状态 / 看 agent 上的 MCP / harness 状态。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli plans --help"
---

# arkcli plans

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与共享安全规则。**
**CRITICAL — 任何 `plans buy` / `plans renew` / `plans personal rotate-apikey` / `plans team seat-assign` / `plans team rotate-apikey` 在执行之前，务必先用 Read 工具读取对应 `references/*.md`，禁止盲目调用。**
**CRITICAL — Token 额度包 / Token 资源包是 platform 按量计费产品，不是本 skill 管理的 Agent Plan / Coding Plan。不得因「套餐」或价格字样调用 `plans get/buy/renew`；转 [`arkcli-profile`](../arkcli-profile/SKILL.md) 解释 `platform` + `/api/v3` 归属。**

## Agent Plan 控制台的超额后付费入口

用户问订阅页为何没有「超额后付费管理」按钮时，这是控制台套餐问题，不是 `arkcli helper` 的本机 TTY 矩阵问题。先用 `arkcli plans get` 核对当前身份是否持有生效的 Agent Plan 个人版或团队版；控制面查询失败时如实说明订阅状态未知，不把凭证错误当成未订阅。再读取官方[超额后付费管理](https://ark.volcengine.com/region:cn-beijing/docs/agent-plan-personal-over-quota-postpaid-management)正文核对当前页面路径和前提，不用 Helper 状态代替控制台页面事实。

官方个人版文档将入口放在[开通管理的 Agent Plan 页面](https://ark.volcengine.com/region:cn-beijing/openManagement?LLM=%7B%7D&advancedActiveKey=agentPlan)「套餐信息」区域；进入后可分别管理模型与 Harness。用户所说的“订阅页”可能就是这个 Agent Plan 页面，不能断言“订阅页没有/不应有按钮”；先给文档的精确页面与区域，再核对用户实际所在页面。模型超额后付费要求先在「开通管理」开通模型服务，开启超额后付费前还需实名认证。Harness 也可在该页对应卡片管理，关闭该 Harness 的套餐抵扣会同时关闭其超额后付费。先问清个人版或企业版、模型服务开通状态；团队版额外核对管理员对席位的授权规则，不把个人版前提直接套到企业版。CLI 不能观察 Web 按钮是否实际渲染，不能仅凭文档断言某个前提就是这次缺按钮的根因；需要用户提供页面状态或由控制台排查。只作指引，不替用户开启真实计费开关。

## Agent Plan 模型消失与下线通知

用户明确说 Agent Plan 中某个旧模型不见了，或问下线短信后是否要重设时，先读 Agent Plan 官方[模型下线公告](https://ark.volcengine.com/region:cn-beijing/docs/agent-plan-personal-model-deprecation-announcement)中**该模型所在行及较新的后续迁移行**，直接交付公告的停止新用户、正式停服时间与建议迁移模型；不要横查 Coding Plan 的模型概览或在多篇无关文档中耗尽本轮。公告是产品时间表，不是当前账号实时查询；用户问“我现在还可用吗”时再用 `arkcli plans model-list --plan agent-plan --format json` 核对本账号清单，失败就明确实时状态未知，不把 STS 错误当作模型已不可用，也不因模型广场仍显示 Published 就推翻套餐下线公告。

问“需要重新设置吗”时按配置类型区分：使用 `auto`/`ark-code-latest` 等自动路由者一般不需要改工具中的固定模型名；工具里写死旧 Model Name 者需在停服前改为**当前仍受支持**的目标并验证；本人自建且绑定旧模型的 `ep-...` 另需核对并迁移 Endpoint。不能从订阅存在或公告本身推断用户属于哪种配置；无法读到本人配置/资源时给条件化操作，不替用户改配置或创建接入点。若公告先推荐的迁移模型也进入后续下线期，以该公告更新的迁移链和当前支持列表为准，不把旧建议当永久可用。

## 购买意图的逐步收参与失败重入

购买信息缺失时按固定顺序逐个收敛，不把所有问题堆在同一轮：

1. 先问套餐家族：`Agent Plan` 还是 `Coding Plan`。
2. 再问版本与档位：只列已选家族的个人版档位及一个企业/团队版入口。个人版选择同时确定 `--plan` 与 `--type`；企业入口只确定团队 `--plan`。
3. 选择企业/团队版后，再单问同家族档位 `--type`，然后单问席位数 `--quantity`。Agent Plan 档位为 `small/medium/large/max`，Coding Plan 为 `lite/pro`。
4. 最后问时长 `--duration`，收齐后进入协议确认。

档位选项必须说明当前价格、典型场景和关键限制，按 [buy 的档位说明](references/arkcli-plans-buy.md#档位选项的业务说明) 核对来源，不只显示裸档位名；未知价格/权益明确标未确认。时长可提供 1 / 3 / 6 / 12 月快捷选项，但仍接受 1–12 内的其他合法整数。

同时缺多项时，当前轮只问上述顺序中第一个未知决策。宿主有结构化选择能力时优先用，但通用 Skill 不写死工具名；没有时用简短选项文本退化。参数已全部明确且用户只要命令时，只给不带 `--yes` 的第一次协议闸门命令，不执行。

“当前轮只问一个决策”也限制当前轮可展示的信息：只列出该层的选项和必要差异，不得在说明、括号、示例或下一步预告中泄露后续层的选项名、合法值或 flag。比如套餐家族尚未确定时，回复中只能出现 `Agent Plan` 与 `Coding Plan` 及二者的简短区别，不能同时列出个人/团队、`small/medium/large/max`、`lite/pro`、时长或席位数。用户回答当前层后，下一轮再展示下一层。

`payment_failed` 且返回 Order ID 表示订单已经落库。必须复述该 Order ID，引导用户到 console 对这一单补付或取消；不重新收参，不执行 `plans buy` / `plans renew`，也不用 billing 出账记录代替已有订单的支付重入点。

`ProductStockNotEnough` 表示库存不足，不是参数错误；不重试、不重问已确认参数、不换档或拆单，按 [buy 的失败语义](references/arkcli-plans-buy.md#失败语义) 交付对应套餐控制台处理入口。transport error 同样保留本轮参数，不能重新开始购买向导。

续费不重新选档位：先确认 `--plan` 与续费月数，团队版从当前席位清单选择 `--seat-ids`；沿用原订阅/席位档位，不传购买专用的 `--type` 或 `--quantity`。用户已说清的参数不重复追问。仅询价走 `pricing plans` 或对应命令的 `--estimate`，不下单。

协议闸门返回后，把本次套餐、档位、时长、席位范围、实付价 `total_amount_cny` 与原价（如有）一起交付；价格和权益以当前返回为准，不用历史示例价承诺优惠、视频资格或有效期。超时/连接中断不能证明“未下单、未扣款”：保留错误和已有订单/席位 ID，先核对订单状态，不自动重放计费写操作。

## 🔒 协议闸门（plans buy / plans renew 强制流程）

`plans buy` / `plans renew` 是**计费写操作**，加 `--yes` 才真扣款。**不传 `--yes` 不传 `--estimate` 时**, CLI 不下单, 而是返回:

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

`--estimate` 是在线询价路径,不下单也不需要协议确认；它不是 Client Preview。用户最终走 `--yes` 真下单时仍需先看协议。

## 业务定位

`arkcli plans` 管理两类 ARK 套餐：

- **Agent Plan**（agent-plan / agent-plan-team）：智能体调用套餐，按 tier 划档（small/medium/large/max）
- **Coding Plan**（coding-plan / coding-plan-team）：编程辅助套餐，按 tier 划档（lite/pro）

两类套餐都有个人订阅与企业/团队席位，但 **Key 来源不能按 personal/team 一刀切**：

| Profile 类型 | 资源单位 | Key 来源与轮换入口 | OpenAI 兼容数据面 Base URL 路径 |
|---|---|---|---|
| `agent-plan` 个人版 | 个人订阅 | `Scene=RealAgentPlanPersonal` 专属 Key；仅此类型可用 `plans personal rotate-apikey` | `/api/plan/v3` |
| `coding-plan` 个人版 | 个人订阅 | 与 `platform` 共用普通 API Key 池；在普通 Key 管理入口处理，**没有 Coding Plan 个人版专属 Key** | `/api/coding/v3` |
| `agent-plan-team` | 企业席位 | 当前 Running 席位的 Key；用 `plans team rotate-apikey` 管理 | `/api/plan/v3` |
| `coding-plan-team` | 企业席位 | 当前 Running 席位的 Key；用 `plans team rotate-apikey` 管理 | `/api/coding/v3` |

> personal 一个账号下一份订阅；team 是按席位（SeatID）粒度管理，每个席位绑一个子用户。

> **本 skill 不含套餐用量 / 配额查询**（"用了多少 / 还剩多少 / 几号刷新 / 按模型拆分"）。`plans get` 只回答"我**持有**哪些套餐 + 状态(Effective/Running)"，不回答"用了多少"。配额视图在 [`../arkcli-usage/`](../arkcli-usage/SKILL.md) — 详见下方"快速决策"分流表。

## 套餐资源与接入配置的证据边界

- `plans model-list` 列的是套餐文本/路由模型，**不是图片、视频资源的完整目录**。用户问“我的 Agent Plan 能否生图/有哪些图片模型”时，先从本轮 `arkcli auth status --format json` 的 `profiles_summary[]` 按 `type=agent-plan` / `agent-plan-team` 找到实际 `name`，再执行 `arkcli resources list --profile <完整 name> --modality image --format json`，读取 `items[].id`、`resource_kind`、`invocable` 和 `is_default`。`--profile` 要填完整 Profile 名，不是 `agent-plan` 这样的类型字面量；多个同类型候选需核对身份/项目，不能猜选。不能因 `plans model-list` 没有 Seedream 就回答“不支持生图”，也不把示例数量当当前权益。
- “Coding Plan 下的接入点”要区分订阅、文本路由与 `ep-xxx` 推理接入点：`ark-code-latest` 是套餐模型路由别名，**不是一个 `ep-xxx` Endpoint**。用对应 Coding Plan Profile 的 `resources list --modality text --format json` 核对 `items[].resource_kind` / `id` 后回答是否存在 `ep-xxx`；不能用 `plans get` 查到订阅就推翻“没有接入点”的候选结论。Coding Plan 的图片/视频另走 platform Endpoint 池与后付费 Key，见 [`arkcli-resources`](../arkcli-resources/SKILL.md)。
- 给 SDK、Claude Code、OpenCode 等配置 Plan 时，先确认 Profile 类型：Agent Plan 数据面为 `/api/plan/v3`，Coding Plan 为 `/api/coding/v3`；普通 platform 才是 `/api/v3`。个人 Agent Plan 专属 Key、个人 Coding Plan 使用的普通池 Key、团队席位 Key 不可混用；只从本轮身份/Key 元数据与对应 [Profile 类型矩阵](../arkcli-profile/SKILL.md)确认，不把一个套餐的 Key 指向另一条计费路径。
- `plans model-list` 返回的展示名是 `model_name`，Plan 数据面及 Agent 配置中的 `model` **优先填 `output_name`**（如实际返回的 `auto` 或无日期尾缀名称），仅在该字段为空时按 [model-list 字段口径](references/arkcli-plans-model-list.md) 回退；带版本的 `model_id` 主要用于精确查询/路由目标控制面写入，不能见到 `selected_model_id` 就直接抄进客户端环境变量。用户贴配置模板时逐个数清占位字段，不补造未出现的槽位。
- 用户问“我的 Plan Key 在哪/怎么取”时先转 [`arkcli-auth` 的 Key 查询与交付](../arkcli-auth/references/api-key-query.md) 定位目标；`apikey.get_raw`/`profile apikey get` 有条件提供读取能力，**不能声称 CLI 只能轮换、不能读取已有 Key**。导出受 Access 与宿主限制，失败应如实报告；查询失败绝不自动触发破坏性的轮换，也不在聊天中回显完整 Key。

## 适用场景

- 查看当前账号下持有哪些套餐 → `plans get`
- 下单买套餐（含个人 / 团队）→ `plans buy`
- 续费已有套餐 → `plans renew`
- 看套餐支持的模型清单 → `plans model-list`
- 切换 ark-code-latest 的路由目标（Auto 智能调度或具体影子模型）→ `plans model-apply`
- Agent Plan 个人版轮换专属 APIKey → `plans personal rotate-apikey`；Coding Plan 个人版普通 Key 不走此命令
- 列出 / 筛选企业版席位 → `plans team seat-list`
- 把企业版席位绑给子用户 → `plans team seat-assign`
- 轮换企业版席位的 APIKey（自己 or 管理员批量）→ `plans team rotate-apikey`
- 查本机 AI Agent 上专属 Harness 能力装没装（只读）→ `plans harness-status`：按能力卡报豆包搜索 / 专业数据集 / Agent 记忆（MCP），外加全局**火山引擎Supabase**（CLI+Skill 装没装）。「我的 supabase / MCP 装好了吗」都走这里（Agent 记忆卡仅个人版 `agent-plan`；团队版报 `absent` 是预期、非漏装，详见其 reference）

## 反唤起信号

- 只问套餐已用/剩余额度、刷新时间或席位用量 → [`arkcli-usage`](../arkcli-usage/SKILL.md)，不是 `plans get`。
- 要读取/注入已有 API Key 的值而非轮换 → [`arkcli-auth`](../arkcli-auth/SKILL.md) 的 Key 查询流程；不要为只读需求调用 `plans personal/team rotate-apikey`。
- 要创建或管理普通 `ep-xxx` 接入点 → [`arkcli-deploy`](../arkcli-deploy/SKILL.md) / [`arkcli-infer-endpoint`](../arkcli-infer-endpoint/SKILL.md)；Plan 持有或 `ark-code-latest` 路由信息不能代替 Endpoint 查询。

## 快速决策

- 用户问 "我有什么套餐 / 我订阅了什么"：直接 `arkcli plans get`，零参数（**仅持有列表 + 状态，不含用量**）
- 用户问“何时到期 / 还剩几天”：先读 [get 的到期时间规则](references/arkcli-plans-get.md#到期时间--还剩多少天)；仅在能安全捕获并脱敏 debug 的环境核实对应上游时间字段，按北京时间（UTC+8）交付。不能直接回显 debug，也不能把 Usage 的额度刷新时间当订阅到期时间。
- 用户问 **"我用了多少 / 还剩多少 / 几号刷新 / 套餐内 vs 套餐外"** —— **不在本 skill**，转 [`../arkcli-usage/`](../arkcli-usage/SKILL.md)：
  - "我的套餐还剩多少 / 几号刷新" → `arkcli usage plan` 或 `arkcli usage balance --type plan`
  - "我哪个模型用得最多 / 套餐内套餐外比例" → `arkcli usage plan-details`
  - "团队席位的用量" → `arkcli usage seats --product agent-plan-team --with-usage`
- 用户问 "Agent Plan / Coding Plan 支持什么模型"：`arkcli plans model-list --plan <plan>`
- 用户要 "切换 ark-code-latest 底层模型 / 改成 Auto 智能调度 / 锁定某个模型"：`arkcli plans model-apply --plan <plan> --model <model-id|output-name|auto>`（**写操作**，与控制台联动；可选项先 `plans model-list` 确认）
- 用户要“买套餐”：读 [buy.md](references/arkcli-plans-buy.md)，收齐 `--plan`、`--type`、`--duration`，团队版另收 `--quantity`，不要替用户做选择
- 用户要“续套餐”：读 [renew.md](references/arkcli-plans-renew.md)，确认 `--plan`、续费月数，团队版另选已有 `--seat-ids`；保持原档位，不传 `--type` / `--quantity`
- 使用 plan 时被拦 `plan_agreement_required`（三方渠道订单首用签署）：见 [references/arkcli-plans-agreement-signing.md](references/arkcli-plans-agreement-signing.md)；**禁止**擅自设 `ARKCLI_ALLOW_HEADLESS_PLAN_AGREEMENT` 替用户签署，引导用户在交互式终端完成签署
- 使用 plan 时返回 `plan_agreement_status_unavailable`：签署状态尚未确认，检查网络与登录状态后重试；`--yes` 和 `ARKCLI_ALLOW_HEADLESS_PLAN_AGREEMENT` 都不能绕过该错误
- 用户要 "重置 / 轮换 APIKey"：分清个人版还是企业版，参考对应 reference；**写操作，原 APIKey 立即失效**，必须按 reference 走二次确认（除非用户明确要 `--yes` 跳过）
- 用户要 "**查席位 / 看团队 seat 绑定情况 / 谁绑了哪个 seat / 列出席位 / 哪些席位激活了 / 团队席位 admin 视图**" → `plans team seat-list --plan <agent-plan-team|coding-plan-team>`(**这是 seat 管理的默认入口**,管理视角列基础信息 + 绑定关系,**不带用量数字**;要看每个 seat 用了多少 token / 套餐百分比 → `arkcli usage seats --with-usage`)
- 用户要 "把席位分配给员工 / 给员工分配 seat / 解绑席位"→ `plans team seat-assign`,先准备好 `seat-id=user-id` 配对清单
- 用户**只知道员工用户名**不知道精确 UserID：先 `arkcli iam userid --username <prefix>` 反查到 `user_id`，再喂给 `seat-assign --bind seat-id=<user_id>`

## Agent 快速执行顺序

1. 先确认认证状态：`arkcli auth status`；缺失走 `../arkcli-auth/`
2. 读操作（`get` / `model-list` / `seat-list`）直接执行；只在用户问 "我的" 时考虑当前身份
3. **写操作（`buy` / `renew` / `rotate-apikey` / `seat-assign` / `model-apply`）** 务必：
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
| `plans model-apply` | 改变套餐请求路由指向（立即生效、与控制台联动） | 非交互必须显式 `--model`；先用 `plans model-list` 确认可选目标；团队版无席位会报错 |
| `plans team seat-assign` | 修改席位绑定关系 | 显式 `--bind seat-id=user-id`，自动调 IAM 反查 UserName |
| `plans team rotate-apikey` | **原 APIKey 立即失效** | self-rotate 默认 agent-plan-team；admin batch 通过 `--seat-ids` |

## 命令一览

| 命令 | 类型 | 说明 |
|------|------|------|
| [`plans get`](references/arkcli-plans-get.md) | 读 | 列出当前账号持有的套餐（个人版 + 团队版聚合） |
| [`plans buy`](references/arkcli-plans-buy.md) | 写（计费） | 下单购买套餐 |
| [`plans renew`](references/arkcli-plans-renew.md) | 写（计费） | 续费已有套餐 |
| [`plans model-list`](references/arkcli-plans-model-list.md) | 读 | 列套餐支持的模型；展示用 `model_name`，Plan 配置优先用 `output_name`，当前路由看 `selected_model_id` / `selected` |
| [`plans model-apply`](references/arkcli-plans-model-apply.md) | 写 | 设置 ark-code-latest 路由目标（auto 或具体影子模型） |
| [`plans personal rotate-apikey`](references/arkcli-plans-personal-rotate-apikey.md) | 写（毁坏） | 轮换 Agent Plan 个人版 APIKey |
| [`plans team seat-list`](references/arkcli-plans-team-seat-list.md) | 读 | 列出企业版席位 + 多维度筛选 |
| [`plans team seat-assign`](references/arkcli-plans-team-seat-assign.md) | 写 | 批量绑定企业版席位到子用户 |
| [`plans team rotate-apikey`](references/arkcli-plans-team-rotate-apikey.md) | 写（毁坏） | 轮换企业版席位 APIKey（self / admin batch） |
| [`plans harness-status`](references/arkcli-plans-harness-status.md) | 读 | 查本机 AI Agent(claude-code/codex/opencode/openclaw/trae)上 Agent Plan 内置 MCP 的安装状态(只读、免登录) |

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
