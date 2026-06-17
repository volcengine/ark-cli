# 用户身份解析（"我的 xxx" 语义）

> 本文从 `arkcli-shared` 拆出。触发场景：用户请求带"我的 / 我创建的 / 我部署的 / my <资源>"这类主语自指语义。属 arkcli-auth（身份）范畴，仅在该场景加载。

用户请求出现"我的 / 我创建的 / 我部署的 / my <资源>"这类**主语自指**语义时，
agent 必须把过滤限定到当前登录身份名下的资源，**不能**把"全部"列表当成"我的"列表返回。
**绝对禁止**直接解析 `~/.arkcli/.env` 里的 JWT。

> **本节只规定抽象决策框架。** 具体过滤入口（Tag 字符串拼装 / Filter 字段名 / 服务端
> 参数 / 是否客户端 jq 过滤）**不在此写**，由该资源所属 skill 的 `references/*.md`
> 自己声明。原因：**不存在跨资源通用的"我的"过滤公式** —— endpoint 走 `TagFilters`，
> API Key 火山走 `Filter.UserId`，
> usage 已经天然按账号 scope，机制完全不同。把 endpoint 专属的 `IAMUser/<uid>/<name>` Tag 公式
> 硬套到 API Key / model 等资源上一定会得到错误结果。

## 触发关键词（中英同义）

- "我的 endpoint / 我创建的接入点 / 我跑的 / 我部署的"
- "我的 API Key / 我的 key"
- "我的用量 / 我用了多少 token / 我消耗了多少"
- "my endpoint / my deployment / my usage / endpoints I created" 等英文同义

## 决策顺序（严格按此优先级）

1. **看该命令是否有内建 `--mine`（或同义服务端过滤 flag）**。有就直接用 ——
   一行结束、不需要 `whoami`、不需要 jq、不会因身份串拼错漏数据。
   判断方式：读该命令在所属 skill 的 `references/<cmd>.md`。
2. **没有 `--mine` 时，读该资源 skill 的 reference**，按 reference 声明的入口走。
   入口形态可能是请求侧 `Filter.<字段>`，可能是响应里 `Tags[]` 的客户端 jq 过滤，
   也可能是已天然按账号 scope 无需额外过滤。**不替资源猜入口。**
3. **reference 没说**：抽样一两页响应看实际过滤字段，并提示用户该 skill 缺文档。
   不要自己拍脑袋套别的资源的过滤公式 —— 跨资源套用一定会错。

## 资源-入口速查（详细规则在各 skill 自己的 reference 里）

| 资源 | 详见 | 入口形态 |
|---|---|---|
| infer endpoints | `arkcli-infer-endpoint/references/arkcli-infer-endpoint-list.md` | 内建 `--mine`（服务端 `TagFilter sys:ark:createdBy`） |
| API Keys | `arkcli-auth`（`arkcli auth apikey` 已内建 personal/all 切换） | 火山：服务端 `Filter.UserId` |
| Foundation Models | `arkcli-models` | 公共目录，无"我的"概念 |
| 自定义模型 | `arkcli-models/references/arkcli-models-list.md` | 当前**无**服务端 `--mine`，需 `--page-all` + 客户端字段过滤 |
| Usage / 用量 | `arkcli-usage` | 默认即当前账号 scope，无需额外过滤 |

新增资源时**在该 skill 自己的 reference 里**写明入口；不要把具体公式塞回本文。

## Shell 变量名禁忌（与具体过滤入口无关，全局适用）

如果某条 reference 让你在 shell 里临时存身份字段（如 `whoami` 输出），**禁止**使用
`UID` / `EUID` / `HOME` / `PATH` / `PWD` / `USER` 等 POSIX shell 内置变量名。
`UID` 在 zsh 下是 real user id，赋值会触发 `setuid()` syscall，普通用户报
`failed to change user ID: operation not permitted`；bash 下报 `readonly variable`。
**临时变量统一用小写**（`ark_uid` / `my_uid` / `arkcli_user`）。

最稳的避坑方式：**优先走第 1 条决策**（命令内建 flag），根本不让 user_id 经手 shell 变量。

## 认证态边角

- `whoami` 显示 `auth_method=aksk`，或 `sso_expired=true` 且缺 `user_id`
  → "我的 xxx" 无法满足，按当前 profile tenant 引导用户重新 SSO 登录（火山：`arkcli auth login volc-sso`），**不要用 AK/SK 模式重试**。
  使用 `--mine` 的命令在这种态会直接报 `--mine requires an SSO sub-user login`。
- `whoami` 显示 `is_root=true`（账号主身份）
  → root 身份在部分 API 响应里 `UserId` 可能是 `0`；具体 root 处理（直接拒绝 / 走
  AccountId / 双条件过滤等）由各资源 reference 自己声明，不给统一规则。
- 命中限流（`AccountFlowLimitExceeded`）
  → 调大 `--page-delay`（推荐 `800` ms），不要原地重试。

## 禁止

- 不要从 `~/.arkcli/.env` 读取 `VOLCENGINE_ID_TOKEN` 自己解 JWT；whoami 已经做了
- 不要让用户自己提供 `user_id`；CLI 已经能拿
- 不要把"全部资源"当成"我的资源"返回（特别是跨 project / 跨用户的列表）
- 不要在命令已支持 `--mine` 时仍走 whoami + jq —— 优先级是反的
- 不要把 endpoint 的 `IAMUser/<uid>/<name>` Tag 公式套到 API Key / usage / model 等
  其它资源 —— 那不是全局机制，**只在带 `sys:ark:createdBy` Tag 的资源上有效**
