# 认证态边界与错误恢复（控制面 STS/AK-SK + API Key 数据面恢复）

> 本文从 `arkcli-shared` 拆出。仅在以下两类**鉴权态相关**场景加载：① 当前是 AK/SK 态、要判断能调什么；② 数据面（Runtime）调用撞到 401/403/AccessDenied 类错误要恢复。常规读写任务不需要读本文。

## 控制面用 STS 还是长效 AK/SK（agent 决策）

> **先记结论：绝大多数情况都用不到长效 AK/SK。** 只要 SSO 登录过（`auth login` 现在唯一可用的登录方式就是 SSO），控制面就一直走 **SSO 派生的 STS**，长效 AK/SK 即便配过也是旁路、不参与签名。

控制面（火山 OpenTOP）签名只有两条路，**在建客户端时按"有没有 SSO IDToken"二选一定死**，运行时不再改：

```
有 cfg.IDToken（= SSO 登录过，含已过期但可 refresh）
   → STS 路：arkcli 自签 V4，凭证是 SSO 派生的临时 STS（带 X-Security-Token 头）
   → STS 过期 → 自动 refresh → refresh 不了才报错让你重登；【绝不回落到 AK/SK】

无 cfg.IDToken  且  配过 AK/SK（cfg.AK && cfg.SK）
   → AK/SK 路：volcengine-go-sdk 用【静态长效 AK/SK】直接 V4 签名（无 session token、无 X-Security-Token 头）
```

要点（都经代码核实，别再写"AK/SK 派生 STS"——它不派生，是静态直签）：

- **SSO 与长效 AK/SK 互斥，SSO 优先**：有 IDToken 就只用 STS，AK/SK 完全不碰（`NewArkClient`：`if cfg.IDToken != "" { return ... }`，根本不构造静态 AK/SK 客户端）。
- **长效 AK/SK 真正被用，当且仅当**：`IDToken` 为空（从没 SSO / 已 logout）+ 配过 AK/SK。现状下 `auth login` 的 AK/SK 通道已关，AK/SK 只能由 `arkcli config init --access-key/--secret-key` 注入——所以这条路**几乎只出现在"专门配了 AK/SK 且从没 SSO"的小众场景**。
- **STS ≠ 可降级成长效 AK/SK**：STS 必带 `X-Security-Token` 头，长效 AK/SK 不带；两者走不同代码路（自签 V4 vs SDK），不是等价请求。**别把 STS 里的 AK/SK 抠出来当长效用**——丢了 session token 头，server 判未授权。


### 各平面怎么走

- **控制面**（列模型 / endpoint / API Key / `+deploy` / infer endpoint 写操作）：火山按上面的 STS 或长效 AK/SK 二选一。
- **数据面**（`+chat` / `+gen` / `+understand` / embeddings）：永远用 ARK API Key（`Authorization: Bearer <api-key>`），跟控制面 STS / AK-SK 之争无关。
- **"我的 xxx"身份过滤**：长效 AK/SK 态拿不到 IAM 子用户身份（无 IDToken），自指过滤受限 —— 见 [`identity-resolution.md`](identity-resolution.md)。

### Agent 判定规则

1. 默认认为控制面走 SSO/STS（最常见态）。先看 `auth status` 的 `auth_method` 再决策，不要先发命令看 fallback。
2. 看到 `requires Volc SSO STS` 报错 → STS 缺失 / refresh 失败，转 [`../SKILL.md`](../SKILL.md) 引导 `arkcli auth login volc-sso`，**不要原地重试、也不要指望它回落长效 AK/SK**（不会）。
4. 只有当 `auth_method=aksk`（确实无 SSO、专配了长效 AK/SK）时才按长效 AK/SK 态思考；这是小众场景，别默认往这上面套。
5. 不要把 `arkcli api` 当绕过认证约束的后门——它走同一套客户端，认证态不对照样被拒。

## API Key 模式的错误恢复（agent 决策）

数据面（Runtime）调用——`+chat`、`+gen`、embeddings、image / video 生成等——使用 ARK API Key 鉴权（`Authorization: Bearer <api-key>`）。当 agent 在数据面调用中**遇到鉴权类或权限类错误**时，按下面规则恢复，不要原地重试。

### 触发特征

任一即满足，判定为 API Key 失效 / 缺失 / 不匹配 / 权限不足：

- HTTP 401 / 403
- 错误前缀含 `arkruntime.<action>:`（来自 SDK，标识 Runtime 数据面调用）
- 错误码：`"code":"AccessDenied"`、`"code":"InvalidApiKey"`、`"code":"Unauthorized"`、`"code":"AuthenticationError"`
- 错误文案含 `do not have access` / `Unauthorized` / `Authentication` / `api key` / `expired`
- arkcli 内部报 `runtime auth failed` / `runtime API key not configured`

典型样例：

```
error: arkruntime.create_responses: Error code: 403 - {"code":"AccessDenied","message":"The request failed because you do not have access to the requested resource. Request id: ..."}
```

### Agent 处置流程

1. **不要重试原命令** —— 鉴权 / 权限失败重试只会被同样的错误打回。
2. 用 `arkcli auth status` 看当前 `api_key` 字段；若缺失或明显失效，进入下一步。
3. **引导用户跑 `arkcli auth apikey`**（交互式，会列出该账号下所有 API Key 让用户选并持久化到 `~/.arkcli/.env`）。这条命令必须由用户亲自执行，agent 不要试图非交互执行。
4. **边界场景**：
   - **用户一个 API Key 都没有** → `arkcli auth apikey` 列表为空 → 引导用户去 console 创建：`https://console.volcengine.com/ark/region:ark+<region>/apiKey`（`<region>` 跟随当前生效 region，例如 `cn-beijing`）；创建完回来再次执行 `arkcli auth apikey` 选刚创建的那把
   - **用户已有 API Key 但持续报 `AccessDenied` 类权限错误** → 当前这把 key 缺少目标资源（模型 / endpoint）的访问权限；单纯重选可能还是同样错误。引导用户去 console 给当前 key 加权限，或新建一把带正确权限的 key，再回来 `arkcli auth apikey` 切过去
5. 选好 key 后，**回到用户原始任务继续执行**，不要停在 `auth apikey` 的结果上。

### 与控制面的区分

- 控制面接口（model / endpoint 管理 / usage 等）的鉴权错误，参考上面"控制面用 STS 还是长效 AK/SK"和 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 的"认证闸门"；不要把控制面错误按本节流程引导到 `auth apikey`。
- 区分线索：报错前缀是否是 `arkruntime.<action>:`，或路径是否是 OpenAI 兼容的 `/chat/completions`、`/embeddings`、`/images/generations`、`/responses` 等数据面端点。
