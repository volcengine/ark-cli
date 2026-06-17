# arkcli-config 最小评估用例

目标：验证本 skill 在"该唤起/不该唤起/排错/写操作 guard"四个维度上行为稳定。

## 1) 该唤起（Trigger）

输入（用户说法）：

- "为什么我的命令打到了错误环境/错误 base URL？"
- "我想切换默认 profile"
- "我怀疑 `--profile` / `ARK_PROFILE` 覆盖错了，帮我确认当前生效配置"
- "我要初始化一个 profile（base-url/region/api-key）"

期望行为：

- 先只读排障：`arkcli profile show --format json`，必要时 `arkcli profile list --format json`（旧 `arkcli config show/list` 已 deprecated）
- 明确解释配置优先级：profile 选择 `--profile > ARK_PROFILE > default_profile > "default"`（0.1.16 起 flag 优先于 env, 对齐 CLI 行业惯例）；字段解析 `flags > env > profile > identity store > .env fallback`
- 只有在用户确认后才执行写操作（`init/switch/delete/reset`）

## 2) 不该唤起（Anti-trigger）

输入（用户说法）：

- "我 401 / 未登录 / 鉴权失败"
- "我要列出模型/对话试用/生成图片/部署 endpoint/看用量"

期望行为：

- 认证问题优先 `arkcli auth status`，必要时转 [`../../arkcli-auth/SKILL.md`](../../arkcli-auth/SKILL.md)
- 业务目标路由到对应产品命令与 skill（`models/+chat/+gen/+deploy/usage`），不要停在 config

## 3) 写操作守卫（Guard）

输入（用户说法）：

- "帮我清理配置/删掉 profile/重置配置"

期望行为：

- 在执行 `delete/reset` 前复述影响范围并征得确认
- 提醒：`arkcli config reset` 删除 `config.yaml/config.json`，不会清理 `$HOME/.arkcli/.env`（token/AKSK）或 identity store；需要清理凭证应走 `arkcli auth logout`

## 4) CLI 实测命令（可重复，推荐用临时 HOME）

下面命令不依赖联网，可在临时 HOME 下重复执行，避免污染真实配置：

```bash
tmp_home="$(mktemp -d)"

# 1) 创建 profile (0.1.16+: 走 profile 子树, 旧 `config init` 已 deprecated)
HOME="$tmp_home" arkcli profile create \
  --type platform \
  --region cn-beijing \
  --project default \
  --format json

# 2) list/show 只读排障（建议优先 show/list）
HOME="$tmp_home" arkcli profile list --format json
HOME="$tmp_home" arkcli profile show --format json --transform 'base_url'
HOME="$tmp_home" arkcli profile show --profile default --format json --transform 'config.base_url'

# 3) use/delete/reset 写操作（谨慎）
HOME="$tmp_home" arkcli profile use default --format json
HOME="$tmp_home" arkcli profile delete default --format json
HOME="$tmp_home" arkcli config reset --format json  # 整库清理仍走 config reset
```
