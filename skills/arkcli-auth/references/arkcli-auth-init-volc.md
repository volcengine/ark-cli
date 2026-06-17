# arkcli init-volc — 无交互环境变量引导

`arkcli init-volc` 从 `VOLC_INIT_*` 环境变量无交互地生成一个火山 **platform** 类型 profile 并设为 default。用于**云开发机 / CI** 等已注入凭证的场景:外部把 STS / 账号 / API Key 通过环境变量注入后,一条命令即让 arkcli 处于 ready 状态,后续 arkcli / OpenCode 调用无需 SSO 登录或交互选择。

## 何时用 / 何时不用

- ✅ 云开发机启动后, 由注入方在写好 `VOLC_INIT_*` 后自动调用, 让机器内的 OpenCode / arkcli 开箱即用
- ✅ CI / 自动化里已有火山 STS + ARK API Key, 想跳过交互式 SSO 引导
- ❌ **不要**用于本地终端用户首次引导 —— 那走 `arkcli auth login`(SSO Gate 1+2)

## 命令

```bash
arkcli init-volc            # 消费 VOLC_INIT_*, 零交互, 落 platform profile + 设 default
arkcli init-volc --dry-run  # 只预览将创建的 profile, 不写盘
```

- 无参数(`NoArgs`), 全程**零交互**;缺必填环境变量 → 报错退出(非零 exit)
- 幂等:同账号重跑覆盖同名 profile
- 不发任何网络请求(STS / API Key / 身份事实都是注入的, 直接落盘)

## 环境变量契约(`VOLC_INIT_*`)

| 环境变量 | 必填 | 含义 | 缺省 |
|---|---|---|---|
| `VOLC_INIT_STS_ACCESS_KEY` | ✅ | 火山 STS AccessKeyId | — |
| `VOLC_INIT_STS_SECRET_KEY` | ✅ | 火山 STS SecretAccessKey | — |
| `VOLC_INIT_STS_SESSION_TOKEN` | ✅ | 火山 STS SessionToken | — |
| `VOLC_INIT_ACCOUNT_ID` | ✅ | 火山主账号数字 id | — |
| `VOLC_INIT_API_KEY` | ✅ | ARK 数据面 API Key | — |
| `VOLC_INIT_REGION` | 推荐 | region code, 如 cn-beijing | cn-beijing |
| `VOLC_INIT_USER_ID` | 推荐 | IAM 子用户数字 id;root 留空 | 空 |
| `VOLC_INIT_IS_ROOT` | 推荐 | 是否主账号 root:true / false | false |
| `VOLC_INIT_USER_NAME` | 可选 | 用户名(owner_trn 展示) | 空 |
| `VOLC_INIT_PROJECT_NAME` | 可选 | project;空 = 账号全部资源 | account-wide |
| `VOLC_INIT_STS_EXPIRE` | 可选 | STS 过期 epoch 秒 | now+12h |
| `VOLC_INIT_TRN` | 可选 | 真实 IAM trn;给了优先, is_root 从 `:root` 后缀推 | 按 IS_ROOT 构造 |

## 落地行为

- **STS** → 写 `.env` + identity store, 供控制面 V4 签名(`models` / `resources`)
- **API Key** → profile 数据面 key, 供 `+chat` / `+gen`
- **account_id / user_id / is_root** → 持久化为当前 active 身份, 后续命令 cfg 直接读到
- 产出 profile 名:`platform_<region>_<project|accountwide>`, 设为 default

## 输出

```json
{"ready":true,"profile":"platform_cn-beijing_accountwide","type":"platform","region":"cn-beijing","project":"账号全部资源","identity_key":"volc-2100000001","is_root":false,"has_api_key":true}
```

不回显任何凭证明文。

## Agent 路由提示

- 用户说"云开发机引导 / 无交互初始化 / 已注入 VOLC_INIT 怎么让 arkcli ready" → 本命令
- 用户说"我要登录 / 首次配置 / 选 project / 选 key" → `arkcli auth login`(见 [`arkcli-auth-login.md`](arkcli-auth-login.md))
