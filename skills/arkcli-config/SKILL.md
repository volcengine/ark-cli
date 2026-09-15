---
name: arkcli-config
version: 1.2.1
description: "arkcli 本地配置管理。处理 profile 配置归因、update.mode 的 automatic/disabled 策略、config reset 与历史 yaml 排障；profile 类操作优先使用 `arkcli profile <subcmd>`。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli config --help"
---

# arkcli config

**⚠️ 0.1.16 起命令树迁移**：`arkcli config init/list/show/switch/delete` 已标记 deprecated，请改用 `arkcli profile create/list/show/use/delete`（行为一致 + 新增 profile 切面字段 type/project/owner_trn/available_api_keys）。`arkcli config reset` 保留（清整个本地状态，超出单 profile）。0.2.x 删除全部已 deprecated 子命令。

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、命令选择顺序与共享安全规则**
**CRITICAL — 配置类写操作前必须确认用户意图。`profile show/list/keys list` 可能同步远端 Key 并回写本地库存或默认 Key，不能当作无配置副作用的准入探测。**

## 唤起信号（When To Trigger）

- 用户提到：`profile`、`base-url`、`api-key` 覆盖混乱，或仍在使用已删除的全局 `--region/--project-name`
- 用户现象：同一条业务命令“打到了错误环境/错误 base URL/错误账号”，怀疑是配置来源不一致
- 用户目标：初始化/更新 profile、切换默认 profile、重置本地配置文件
- 用户目标：启用、关闭或恢复 arkcli 的自动更新/更新提示策略

## 反唤起信号（When NOT To Trigger）

- 用户要解决的是“未登录/401/鉴权失败/没有权限”：先走 `arkcli auth status`，必要时转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- 用户目标是业务功能（models/+chat/+gen/+deploy/usage 等）：不要停留在 config，排除配置后回到原任务

## 适用场景

- 初始化或更新 profile
- 查看当前解析后的配置
- 切换默认 profile
- 删除单个 profile 或重置整个本地配置
- 排查为什么业务命令连到了错误环境、错误账号或错误 base URL

## Agent 快速执行顺序

1. 先遵守 shared 的认证及宿主边界。普通 CLI 的当前身份/Profile 诊断用 `arkcli auth status --format json`、`arkcli auth whoami --format json`；宿主禁止额外认证探测时，使用宿主已给的上下文，不自行探测。
2. Chat/Gen 的默认模型和路由用 `arkcli resources list --modality <text|image|video> --format json`，按 Resources Skill 区分发现、兼容性和真实调用验证。仅查看默认模型不等于授权发送收费请求。
3. 切换默认 profile：`arkcli profile use <name>`
4. 创建 profile：`arkcli profile create --type=...`（详见 arkcli-profile 子树，本 skill 不重复）
5. `arkcli config reset` 是破坏性操作（清整个本地配置文件），必须确认用户意图
6. 更新策略只使用 allowlist 命令：`arkcli config set update.mode automatic|disabled`
7. 用户明确需要 Profile 详情、清单或 Key 管理时才考虑 `arkcli profile show` / `arkcli profile list`；执行前说明可能同步并回写 Key。用户要求“不改配置/Key”时不执行它们，也不改用旧 `config show/list` 绕过限制。

## 配置归因（优先级）

`arkcli` 的解析优先级（从高到低）：

- Profile 选择: `--profile` flag > `ARK_PROFILE` 环境变量 > `config.yaml` 的 `default_profile` > 第一个 `type=platform` 的 profile > `"default"` sentinel
- API Key: `--api-key` > `ARK_API_KEY` > profile > 同 identity store > 产品 `.env` 兼容值
- Base URL: `--base-url` > `ARK_BASE_URL` > profile 自定义值 > 按 Profile 的 Region/Type 派生 > 默认值
- Region/Project: 由所选 Profile 持有；`ARK_REGION` / `ARK_PROJECT_NAME` 不再覆盖运行时

`--api-key/--base-url` 仅供数据面命令使用。无 Profile 时必须成对提供；覆盖 Base URL 时也必须显式提供 API Key。控制面/本地命令显式收到这两个 flag 会 fail-fast。

排障时要明确”是谁覆盖了谁”，不要直接猜测。

> **0.1.16 修正**：`--profile` 现在优先于 `ARK_PROFILE`（对齐 CLI 行业惯例 flag > env > config）。0.1.15 及更早版本是 env > flag，与本规则相反。

### 证据与结论

说明本轮实际检查了什么、哪些来源覆盖仍未验证，不要求复述固定评测词句。`auth whoami` 的 Profile 字段是持久切面摘要，不能据此声称临时 API Key/Base URL override 已验证或数据面调用一定成功。缺少证据时明确保留未知，不为补字段执行可能回写配置的命令。路径展示使用 `$HOME/...`，不读取凭证文件。

## 核心规则

- 本 skill 负责解释“为什么命令打到了错误环境/错误账号/错误 base URL”，并兼容历史 `config init/list/show/switch/delete` 命令排障语义
- 常规准入沿 `auth status/whoami` → `resources`，不把 Profile 管理命令当无副作用查询
- profile 写操作（create / use / set-default / delete）走 `arkcli profile` 子树
- `arkcli config reset` 负责全量清理；`arkcli config set update.mode <mode>` 只写更新策略。不要用它写任意 YAML 路径。

## Guard Checklist（必做）

- 只读意图：不执行可能回写 Key 的 `profile show/list/keys list`；脱敏不等于无副作用，命令名为 show/list 也不等于只读
- 写操作确认：`profile use/delete` 与 `config reset` 执行前必须复述影响范围并征得确认
- 全量清理提醒：`arkcli config reset` 删除 `config.yaml/config.json`，不会清理 `$HOME/.arkcli/.env`（token）或 identity store；要清理凭证请用 `arkcli auth logout`

## 与其他 skill 的串联

- profile 创建 / 切换 / 资源 default 设置 / API Key 管理 → 转 `arkcli-profile`（推荐）或继续看本 skill 的兼容映射
- 业务命令失败且怀疑是 `--profile`、`--base-url`、API Key 来源或 Profile 的 Region/Project 不一致时，先转到这里
- 配置问题排除后，再回到原始业务 skill 继续
- 配置仍无法解释问题时，再考虑转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md) 或 [`../arkcli-api-explorer/SKILL.md`](../arkcli-api-explorer/SKILL.md)

## 命令一览

| 命令 | 说明 | 状态 |
|------|------|------|
| `arkcli config reset` | 删除整个本地配置文件（保留） | ✅ 活跃 |
| `arkcli config set update.mode automatic` | 显式为当前 exact install 启用 automatic；手工重装后也用它恢复授权 | ✅ 活跃 |
| `arkcli config set update.mode disabled` | 关闭静默自动安装，保留隐式版本检查、更新提示和手工 update | ✅ 活跃 |
| `arkcli profile show [--profile <name>]` | 查看 Profile；可能同步并回写 Key | ✅ Profile 管理，替代 `config show` |
| `arkcli profile list` | 列出 Profile；可能同步并回写各 Profile 的 Key | ✅ Profile 管理，替代 `config list` |
| `arkcli profile use <name>` | 切换默认 profile | ✅ 替代 `config switch` |
| `arkcli profile delete <name>` | 删除单个 profile | ✅ 替代 `config delete` |
| `arkcli profile create --type=...` | 创建 profile（替代旧 `config init`） | ✅ 替代 `config init` |
| `arkcli config init/list/show/switch/delete` | 旧子命令，0.2.x 移除 | ⚠️ deprecated |

公开模式只有 `automatic` 和 `disabled`。`disabled` 保留隐式版本检查和更新提示，但绝不静默安装；显式 `arkcli update` 和 `arkcli update --check` 也始终可用。缺失 `update.mode` 和历史配置中的 `notify` 对外都按 `disabled` 处理，历史配置中的 `notify` 继续按 `disabled` 兼容读取，但不得再建议用户设置 `notify`。

Windows、macOS、Linux 的 fail-closed transaction 与六个产品/平台生产 gate 均已开启。普通 npm postinstall 绝不直接创建 active mutation consent，也不立即更新；只有能证明“此前没有产品状态目录”的 stable 全局 npm 新安装才创建绑定 exact install 的惰性 pending evidence。首次运行和环境变量本身不能绕过后续宽限与 exact consent。

### 新安装 enrollment

只有能证明“此前没有产品状态目录”的 stable 全局 npm 新安装才默认写入 `automatic`。postinstall 只创建绑定当前 exact install 的惰性 pending evidence，不创建 active consent：

1. 第一次成功的人工业务命令在 stderr 告知 automatic 已开启及关闭命令；本次不调度更新，只完成宽限。
2. 第二次成功的人工业务命令只激活 exact-install consent；本次仍不调度更新。
3. 第三次及后续人工业务命令才可能调度 automatic patch 更新。

由 npm `postinstall` 启动的全部 CLI 进程（包括其中的 `+connect`）以及用户手工执行的 `+connect` 都不消耗 enrollment、不做隐式版本检查，也不调度 automatic。AI Skill、CI、非 TTY、Client Preview、`config`、`update` 和内部维护命令同样不消耗 enrollment，也不调度 automatic。更新成功后，下一次成功的人工业务命令只在 stderr 显示一次 `旧版本 -> 新版本` 结果，不修改 stdout 或业务退出码。

当前生产 automatic 按明确产品策略关闭 24/48/72 小时与 10%/50%/100% cohort admission；原 rollout 实现与回归测试仍保留，在线 Probe 仍要求两次独立观测。普通 automatic 仍须满足 exact consent、实时 registry target/SRI/tarball 校验、reservation、退避和全部 staged-apply 安全边界。

手工 npm 重装、降级、安装身份变化或 `--ignore-scripts` 安装后，如果当前 exact install 没有对应 pending/consent，automatic 必须暂停，不得沿用旧安装授权。恢复时由用户明确执行 `arkcli config set update.mode automatic`；长期锁定版本使用：

```bash
arkcli config set update.mode disabled
npm i @volcengine/ark-cli@<exact-version> -g --registry https://registry.npmjs.org
```

新机器首次安装历史版本时，先对安装命令设置 `ARKCLI_NO_UPDATE_NOTIFIER=1`，安装后再写入持久 `disabled`。`disabled` 位于 `$HOME/.arkcli/config.yaml`，npm 重装不能覆盖。`arkcli config reset` 会尽力撤销 exact consent 后再清配置，不会把用户自动放回 automatic。

## 参考

- [`references/arkcli-config-init.md`](references/arkcli-config-init.md)
- [`references/arkcli-config-profile.md`](references/arkcli-config-profile.md)
- [`references/evals.md`](references/evals.md)
