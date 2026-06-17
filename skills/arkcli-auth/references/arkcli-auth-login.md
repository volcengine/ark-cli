# auth login

## 推荐顺序

交互式（推荐）：

```bash
arkcli auth login
```

明确指定方式时：

```bash
arkcli auth login volc-sso                                      # 火山 SSO (浏览器)
arkcli auth login --no-browser                                  # 火山 SSO 无浏览器 (TTY 交互 / 非 TTY 出 Phase 1)
arkcli auth login --no-browser --code <授权码>                  # 无浏览器 Phase 2 (agent/沙箱喂码完成登录)
```

> **0.1.16 暂关 AK/SK 登录通道**：`--access-key/--secret-key` flag 已注释、交互菜单也移除了 "AK/SK" 选项。CI / 自动化 / agent / 沙箱场景请改用 `arkcli auth login --no-browser`（见下方「无浏览器两段式」）。**注意 `--no-browser` / `--code` 是 `auth login` 根命令的本地 flag, 不是 `volc-sso` 子命令的 flag；写成 `auth login volc-sso --no-browser` 会触发 `unknown flag --no-browser`。** AK/SK 通道恢复将在后续版本同步更新本文档。

## 无浏览器两段式跨进程 flow（agent / 沙箱 / CI 必读）

**根因**：老 `--no-browser` 是单进程交互式 —— 打印 URL 后当场 `阻塞读 stdin` 等你粘授权码。agent / 沙箱里每条命令是独立短命进程，stdin 不是真人终端（常是 `/dev/null`），`读 stdin 立刻 EOF`，进程在拿到授权码前就崩（`读取授权码失败: EOF`）。解法是把单进程拆成**两段跨进程命令**，中间用一次性落盘 session 接力 PKCE（对齐 lark-cli `auth login --no-wait` / `--device-code` split-flow）。

> ⚠️ **两段必须在同一运行环境执行**：Phase 1 把 `code_verifier`+`state` 落到 `~/.arkcli/.sso-pending.json`，Phase 2 靠读这个文件接力。**两条命令必须共享同一 `HOME` / 同一持久化卷**（同一容器 / 同一开发机）。若无状态调度把两步派到不同 worker、不同容器或不同 `HOME`，Phase 2 找不到 pending 会报「没有待完成的…会话」。

**CLI 按 stdin 是否 TTY 自动分流，agent 无需判断**：

```text
arkcli auth login --no-browser [--code X]
  ├─ 有 --code X ──────────► Phase 2: 读盘换 token + 激活（须与 Phase 1 同一 HOME）
  └─ 无 --code
       ├─ 非 TTY(agent/沙箱) ► Phase 1: 出 URL + 落盘 PKCE/state + 立即退出（不读 stdin）
       └─ 是 TTY(真人) ──────► 一条命令交互式（打印 URL → 当场粘贴），行为同旧版
```

> `--code` **必须与 `--no-browser` 一起用**（它只是无浏览器 Phase 2 的喂码入口）；单独写 `auth login --code X` 会报 `--code 仅在 --no-browser 模式下有效`。

**agent 执行步骤**：

1. **Phase 1** —— 跑 `arkcli auth login --no-browser`。stdout 是结构化 JSON：
   ```json
   {"stage":"authorize_pending","method":"sso_no_browser",
    "authorize_url":"https://signin.volcengine.com/authorize/oauth/authorize?...",
    "next_command":"arkcli auth login --no-browser --code <授权码>","expires_in_sec":600}
   ```
   把 `authorize_url` **原样**转发给用户（不要改写/编解码），请他在任意设备浏览器完成 SSO，复制页面显示的 **base64 授权码**回来。命令已退出（exit 0），不会挂住。
2. **Phase 2** —— 拿到授权码后跑 `arkcli auth login --no-browser --code <授权码>`（`--code` 必须连 `--no-browser`，且与 Phase 1 同一 `HOME`）。CLI 读 Phase 1 落盘的 `~/.arkcli/.sso-pending.json`（含 PKCE `code_verifier` + `state`）换 token、激活身份、删盘。成功后输出 `{"auth_method":"sso_no_browser","path":"..."}`。

**出错恢复（先看 pending 还在不在，再决定重跑 Phase 1 还是重试 Phase 2）**：

| 报错 | pending 状态 | 怎么办 |
|------|-------------|--------|
| 会话过期（超 `expires_in_sec`，TTL 默认 10min） | 已被清 | **重跑 Phase 1** 拿新 URL |
| `没有待完成的…会话` / pending 不存在（含 Phase 1、2 不在同一 HOME） | 无 | **重跑 Phase 1**（并确保两段同一 HOME） |
| `base64 解码失败` / `state 不匹配(CSRF)` / token 交换失败 | **仍在盘** | **TTL 内直接重试 Phase 2**：纠正后重跑 `--code`——贴对当前 `authorize_url` 的码；token 交换瞬时失败可让用户从**同一 authorize_url** 重新授权拿新码（PKCE/state 仍有效，**无需**重跑 Phase 1） |

> `state 不匹配(CSRF)` 最常见原因是粘了**上一轮 / 别处** authorize_url 的码，不是攻击；pending 未删，粘当前这轮的码重试 Phase 2 即可。真疑似被注入时再手动重跑 Phase 1 换全新会话。

**其它**：
- 凭证可移植仍可用：有浏览器的机器登录后把 `~/.arkcli/` 拷进沙箱复用，与两段式互补。

## 说明

- `volc-sso` 适合火山方舟交互式开发
- `--no-browser` 是真正的 cross-device 流（对齐 volcengine-cli `ve login --remote`）：
  - CLI 不监听任何端口
  - 用户在 **任意设备**（手机/远程电脑/平板）浏览器打开 authorize URL
  - 完成登录后，浏览器页面显示一段 **base64 字符串**（不是 URL，不是 6 位数字）
  - 把这段 base64 串粘贴回 CLI 提示符 `授权码:` 处即可
- 登录成功后凭证写入 identity store，并绑定或切换到对应 tenant 的 profile（同 tenant 重登更新当前 profile 的 identity_key；跨 tenant 登录创建以 tenant 命名的新 profile 并设为 default）
- 如需手动重新拉取或切换 ARK API Key，运行 `arkcli auth apikey`
- SSO 首次登录会自动拉取 API Key 和 Project Name
- 临时覆盖 project：`--project-name <name>` 或 `export ARK_PROJECT_NAME=<name>`

## --no-browser 跟浏览器流的 OAuth2 差异（agent 排障用）

| 维度 | `volc-sso`（浏览器） | `--no-browser`（cross-device） |
|------|-----|-----|
| client_id | `trn:signin:::devtools/same-device` | `trn:signin:::devtools/cross-device` |
| redirect_uri | `http://127.0.0.1:<port>/oauth/callback` | `https://signin.volcengine.com/authorize/oauth/authorize` |
| CLI 本地端口 | 需要（callback server） | 不需要 |
| 用户粘贴内容 | — (自动 callback) | base64 串 |
| refresh token | 必须用 same-device client_id refresh | 必须用 cross-device client_id refresh |

`.env` / identity store 都持久化了 `client_id`（`VOLCENGINE_OAUTH_CLIENT_ID` / `IdentityToken.client_id`），后续 `TryRefreshVolcToken` 会用对的 client_id 续期。**用户/agent 不应手动改这个字段**，否则 refresh 会被 server 拒（RFC 6749 §6）。
