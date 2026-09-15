# 三方渠道订单首用数据协议签署

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

火山官网在抖店/移动商城等三方渠道售卖 **Agent Plan / Coding Plan 个人版**时，下单环节无法签署数据授权协议。后端在 `ListSubscribeTrade` 响应里通过 `NeedCheckAgreement=true` 标记「三方来源新购订单且尚未签署」，arkcli 会在**使用 plan 资源前**拦截并引导签署。签署状态**完全以服务端为准**：CLI 不保留任何本地签署缓存或 mock 旁路，每次使用都在线确认。

## 触发时机

仅当**同时**满足以下条件时执行在线签署状态检查：

- 本次执行走**个人版** plan 数据面（agent-plan / coding-plan profile；团队版与 platform 不触发）；
- 当前具备可绑定账号的控制面身份（SSO/控制面凭证；纯数据面 API Key 直接 fail-closed，见下文）；
- 随后 `ListSubscribeTrade` 返回 `NeedCheckAgreement=false` 时直接放行；返回 `true` 时再调 `GetAccountAgreement` 查账号级签署记录（`AgreementType=AgreeAgentPlanService` / `AgreeCodingPlanService`），`IsSigned=true` 放行、`false` 才弹签署引导；若控制面 invoker、认证刷新、网络或服务端不可用，则 fail-closed，不执行受保护的 Plan 请求。

命中触发的是**使用入口**，不是购买入口：

- `arkcli +chat` / `+gen` / `+understand`（解析到个人版 plan 数据面时）
- `arkcli chat get/delete/list-input-items` 与 `gen get/list/delete`（当前数据面为个人版 plan 时）
- `arkcli api <arkruntime-action>`（Raw API Explorer 指向个人版 plan 数据面时）
- `arkcli helper` / `arkcli helper configure` / `helper mcp`（配置或安装个人版 plan 能力前，包括 CUA）
- `arkcli profile apikey get`（导出个人版 plan 明文 key 前）
- `arkcli plans personal rotate-apikey` 与个人版 `plans model-apply`（变更个人版 plan 凭证或路由前）
- Raw API Explorer 中会变更个人版 plan 或泄露其明文凭证的控制面 action（包括个人版模型路由、Plan key、MCP key 和 OpenViking key）

与 `plans buy/renew` 的 `agreement_required` 协议闸门是两个独立机制：buy/renew 拦的是**下单**，本闸门拦的是**三方渠道订单的首次使用**。

## 交互流程（TTY）

一屏式签署屏（对齐 PRD mock 与 Web 控制台）：

```
◆ 数据授权协议
请注意，使用Agent Plan个人版前需要进行相关协议的签署，在初次使用时请阅读并同意以下协议：
  1. 《火山引擎数据授权协议》 <url>
  2. 《方舟平台专用条款》 <url>
  ...
按 1–N 打开对应协议 · Space 勾选/取消 · H 勾选/取消 Harness · Enter 继续 · Esc 退出
[ ] 已阅读并同意上述 N 份协议（必选，不勾选无法进入使用）
[ ] 已阅读并同意《Harness权益说明和产品专用条款》（可选，仅影响 Harness 抵扣）
```

- 数字键 1–N 在浏览器打开对应协议（停留在本屏，底部给瞬态反馈；URL 同屏展示可复制）；
- 勾选项 ≤ 2 个，**不做焦点移动**：`Space` 切必选（数据协议组）项，`H` 切可选（Harness）项（仅 Agent Plan 有）；
- 勾选框**默认全部不勾**（对齐 Web 控制台，需用户主动勾选）；Agent Plan 个人版 = 6 份数据协议（必选）+ 1 份 Harness 协议（可选），Coding Plan 个人版 = 5 份数据协议（必选，无 Harness 项）；
- Enter 提交时必选未勾会行内红字报错且不退出；Esc / Ctrl-C 放弃签署会**阻断使用**并返回取消错误；
- 提交后调 `SignAccountAgreement {AgreementType}` 完成签署，服务端未回 `IsSigned=true` 按失败处理；CLI 不写任何本地签署记录，下次使用仍由服务端 `GetAccountAgreement` 判定；
- 仅当用户在屏上主动勾选 Harness（Agent Plan）时，签署成功后才 best-effort 调 `UpdateAccountOverdraftSwitch` 开通抵扣；失败只给 warning（「协议已签署，但 Harness 抵扣开启失败」），不回滚签署。

## 非交互环境（CI / 管道 / agent）

严格模式，`--yes` 在非 TTY 下**不放行**：

- 无逃生 env：硬拒，返回 `plan_agreement_required` 结构化错误，引导到交互式终端签署；
- 显式设置 `ARKCLI_ALLOW_HEADLESS_PLAN_AGREEMENT=1`：stderr 审计行后调 `SignAccountAgreement` 签署**必选数据协议**并放行，**不开通 Harness 抵扣**（给真·无人值守自动化，与其它 `ARKCLI_ALLOW_HEADLESS_*` 逃生环境变量同一模式）。

TTY 下传 `--yes` 视为真人已预先阅读并同意必选数据协议，跳过交互直接签署；同样**只签必选协议、不开通 Harness**（对齐 Web 默认不勾），Harness 只能由用户在交互屏主动勾选。

Agent 引导原则（与 buy 闸门一致）：**禁止**在未告知用户的情况下擅自设置 `ARKCLI_ALLOW_HEADLESS_PLAN_AGREEMENT` 替用户签署；正确做法是引导用户在交互式终端跑一次使用命令完成签署。

## 纯 API Key 调用

显式 `--api-key` + Agent Plan/Coding Plan `--base-url` 会形成无控制面身份的 stateless 调用；已有个人版 Plan profile 在 SSO 失效后只剩 API Key 也属于同一类。CLI 无法仅凭数据面 API Key 查询账号的协议签署状态，因此真实执行会 fail-closed，返回 `plan_agreement_identity_required`。引导用户先登录并选择对应的个人版 Plan profile；不要通过改写 Base URL 或直接调用数据面来绕过。Platform 数据面的临时 API Key 调用不受此规则影响。

## 签署状态暂不可用 / 签署失败

CLI 必须在线确认状态（无本地缓存可旁路）。控制面 invoker 创建失败，或 `ListSubscribeTrade` 探测、`GetAccountAgreement` 查询因认证刷新、网络、限流、服务端错误而失败时，返回 exit code 4 的 `plan_agreement_status_unavailable`，保留底层错误供诊断，并在任何受保护业务请求发出前终止。用户已确认签署但 `SignAccountAgreement` 失败或服务端未置 `IsSigned=true` 时，返回 exit code 4 的 `plan_agreement_sign_failed`（数据授权协议签署失败，请稍后重试），同样阻断使用。检查网络和登录状态后重试；`--yes` 与 `ARKCLI_ALLOW_HEADLESS_PLAN_AGREEMENT` 都不能绕过状态确认。

`plans get/model-list`、usage、登录/profile 管理、购买和续费仍可执行：这些是只读查询或恢复/签署所需的控制面路径，不消费个人版 Plan 数据面，也不导出个人版 Plan 明文凭证。团队版不属于本次三方渠道个人版协议范围。
