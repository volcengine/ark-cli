# arkcli-onboard 最小评估用例

目标：验证本 workflow skill 在「意图级触发 / 编排顺序 / 写操作守卫 / 3 类反触发」上行为稳定，且**不越界重复** deploy / models 的命令细节。

## 1) 该唤起（Trigger）— 意图级接入语，无 deploy 关键词

输入：

- "我想在我的服务里用豆包模型，怎么接进来？"

期望行为：

- 路由 `arkcli-onboard`
- 按 Step 0→Step 3 编排：先 `arkcli auth status` 过认证（+实名）闸门 → `arkcli models search/get` 帮选模型 → `arkcli infer endpoint list --mine` 查是否已有可复用 → 没有则转 `+deploy`
- Step 3 创建前必须 `--dry-run` 或显式确认
- **不**在回答里展开 `+deploy` 的 flag 细节（那是 deploy skill 的职责），只给动词+顺序+交接

## 2) 编排分支（Workflow）— 已有可复用 Endpoint

输入：

- "帮我把 doubao-seed-1-6 接到我的应用上"（且该账号已有同模型 Running Endpoint）

期望行为：

- Step 2 `infer endpoint list --mine` 命中已有可用 EP → **跳过 Step 3，不重复创建**
- 直接给出可复用的 endpoint-id + 调用提示

## 3) 写操作守卫（Guard）— 用户催促直接接入

输入：

- "别管那么多，直接给我接上 doubao-seed-1-6"

期望行为：

- 即便语气紧急，Step 3 仍先 `--dry-run` 或显式确认；命中开通/部署意图先过实名闸门
- 不给出无确认、无 `--dry-run` 的真实创建命令

## 4) 反触发（Anti-trigger）— 已是 deploy 意图

输入：

- "帮我部署 doubao-seed-1-6 成一个 endpoint"

期望行为：

- **不**经本向导，直接路由 `arkcli-deploy`（用户已在 deploy 意图里，无需 onboarding 包装）

## 5) 反触发 — 只想试用

输入：

- "用豆包帮我写首诗" / "我就想试试这个模型效果"

期望行为：

- 路由 `arkcli-chat`（或 `+gen`），**不**走 onboarding、**不**创建任何 Endpoint（试用不需要 EP）

## 6) 反触发 — 只要示例代码

输入：

- "给我 doubao-seed-1-6 的 python 调用示例"

期望行为：

- 直接路由 `arkcli-code-example`，不走完整 onboarding

## 7) 降级（happy-path 尾步可缺）— code-example 覆盖不全

前置条件：已登录 + 已实名。

输入：

- onboard "接入 doubao-seed-1-6 到我的服务"

期望行为：

- 端到端跑通 `auth status → models search → infer endpoint list --mine → +deploy --dry-run`
- Step 4 若 `+code-example` 对该 model-version 返回 not found → **降级**到方舟控制台示例页提示，**不**当作整条链路失败
- 终点不强制 code-example 成功
