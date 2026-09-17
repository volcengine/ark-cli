# arkcli-config 最小评估用例

目标：验证本 skill 在"该唤起/不该唤起/排错/写操作 guard"四个维度上行为稳定。

## 1) 该唤起（Trigger）

输入（用户说法）：

- "为什么我的命令打到了错误环境/错误 base URL？"
- "我想切换默认 profile"
- "我怀疑 `--profile` / `ARK_PROFILE` 覆盖错了，帮我确认当前生效配置"
- "我要初始化一个 profile（base-url/region/api-key）"

期望行为：

- 常规准入使用 `arkcli auth status --format json`、`arkcli auth whoami --format json` 和 Resources Skill；宿主禁止额外认证探测时沿宿主协议，不自行探测。
- `profile show/list/keys list` 可能同步并回写 Key，不能当作无配置副作用的查询；用户明确要求 Profile 管理时先说明影响。
- 解释实际命中的配置优先级，Profile 选择包括 first-platform fallback；区分持久切面与临时 override。评分看真实命令、状态与证据，不要求回答复述固定字面量。
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

## 4) 无业务副作用的命令面检查

先检查帮助，不在真实登录目录中创建、切换、删除或刷新 Profile。集成用例由 Eval Kit
提供独立夹具、明确的状态目录隔离与恢复，不能只替换一个 HOME 就声称真实凭证不会被触及。

```bash
arkcli auth whoami --help
arkcli resources list --help
arkcli profile show --help
arkcli profile list --help
arkcli profile use --help
arkcli config reset --help
```

组合回归输入：“当前默认聊天模型是什么，现在能不能用？只查，不发送请求、不改任何配置或 Key。”
要求 Agent 实际加载 Config/Resources/shared，查询当前上下文与默认资源，不执行可能回写 Key
的 Profile 命令，也不发送 Chat/Gen。回答必须区分已查到的默认/兼容信息与尚未验证的实际调用、
权限、额度；不能凭 exit=0、Key Active 或 invocable 就声称全部可用。

## 5) 更新策略

输入（用户说法）：

- “以后自动更新 arkcli”
- “关闭 arkcli 的静默自动更新”

期望行为：

- 明确这是持久化配置写操作，分别执行 `arkcli config set update.mode automatic` 或 `arkcli config set update.mode disabled`。
- 说明公开模式只有 `automatic` 和 `disabled`；缺失配置及历史 `notify` 对外均按 `disabled` 处理。
- 当前六个生产 gate 均已开启；用户显式执行 `automatic` 时应为当前 exact install 建立 consent 并持久化配置，不得把仅写配置冒充为授权成功。
- `disabled` 关闭静默自动安装，但保留隐式版本检查、更新提示以及显式 `arkcli update` / `arkcli update --check`。
- 不得再建议旧 `notify` setter；历史 YAML 中的 `notify` 只作为兼容输入读取。

当前默认 enrollment 还必须满足：

- 只有真正全新且安装版本等于 registry `latest` 的 stable 全局 npm 安装默认进入 `fresh_pending`。
- 第一次成功的、符合安全条件的人工业务命令会在同一策略事务完成 grace 与 exact consent；重新校验 active 后该 invocation 即可调度。
- AI Skill、CI、非 TTY、Preview、config/update 与内部维护命令不消耗宽限、不调度。
- 重装后版本等于 registry `latest` 时不改 mode：已有 `automatic` 换发 exact receipt，已有 `disabled` / `notify` 保持关闭；不等时持久化 `disabled`。
- registry 失败不改 mode、不发新 receipt；`--ignore-scripts` 不完成对齐。恢复时先安装 latest，再显式设置 `automatic`。
- 长期锁定仍建议先执行 `arkcli config set update.mode disabled`，再安装精确版本；安装 `@latest` 不会自动重开 automatic。
- automatic 成功结果只在下一次人工业务命令的 stderr 显示一次，不得污染 stdout 或改变业务退出码。
- 不用通用 YAML 编辑或定时任务替代产品命令。
