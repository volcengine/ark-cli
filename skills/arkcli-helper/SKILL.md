---
name: arkcli-helper
version: 1.1.5
description: "arkcli helper：为 Claude Code / Codex / OpenCode / OpenClaw / Hermes / ZCode 配置火山方舟 Plan 或 Platform Endpoint 的 model/provider，或给支持的 Agent（含 MCP-only Trae）配置 Harness 工具。真人 TTY 运行 `arkcli helper` 可选择模型和 Harness 能力；`helper mcp` 默认保持整组 Agent Plan MCP 行为，传 `--capability datapro|web-search|agent-memory` 只配单项，传 `--capability cua` 只给目标 Agent 安装云电脑 Skill；连 model/provider 一起配置用 `helper configure`，查状态用 `helper list`，移除用 `helper reset`。Agent Plan / Team 还支持专业数据集、豆包搜索、Agent 记忆、AI Native 应用开发底座、Agent 进化与 routing.v1 强制检索路由；上下文出现 ARKCLI ROUTING ENFORCEMENT / routing.v1，或用户要求在豆包搜索、DataPro、OpenViking、Supabase 之间强制分流时也使用本 Skill：调用这四类受管 Provider 前先提交 Route Plan，再调用唯一授权 Provider；原生 WebSearch/WebFetch/WebExtract 位于路由管理范围之外。Platform 只允许本人创建、Running、已验证为文本输出的 Endpoint，不配置 MCP、Supabase 或强制路由。豆包搜索配置成功后默认关闭目标客户端原生 WebSearch；传 `--keep-native-websearch` 可保留，并可用 `helper reset` 完整恢复 ArkCLI 配置。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli helper mcp --help"
---

# arkcli helper —— 给本机 AI Agent 配置 Plan / Platform Endpoint / 注入 MCP

**前置:** 先用 Read 读 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) 获取共享安全规则与认证闸门。

**CRITICAL — `arkcli helper` 整个命令域不支持 `--dry-run`。** Helper
面向 TTY/本地 harness 配置，不是 API request preview。任何 helper
命令带 `--dry-run` 都应 fail-fast 为 unknown flag；Agent 不得生成该
组合，也不得把它当作写操作保护。真实 `configure/reset/mcp/supabase`
仍必须先展示目标与路径并取得明确确认。

## 命令选择（先选最具体的子命令）

- 普通 Helper 准入从 `arkcli helper list` 和用户指定的目标开始；不要为补齐上下文自动执行 `profile show/list/keys list`。这些 Profile 命令可能在线同步并回写 Key，只有用户显式要求 Profile 管理时才进入，并先说明影响。
- 用户要查看或配置 Agent 的 model/provider、Plan、Platform Endpoint，或核对 `--with-mcp` / `--with-supabase` / `--with-routing` 等非交互选项时，必须选择 `arkcli helper configure`；只查看用法也要运行 `arkcli helper configure --help`，不能退化成父级 `arkcli helper --help`。
- 用户只要 Harness 工具、不改 model/provider 时，选择 `arkcli helper mcp [target]`。不传 `--capability` 配整组 Agent Plan MCP；只要专业数据集、豆包搜索或 Agent 记忆时分别传 `--capability datapro|web-search|agent-memory`；只要云电脑时传 `--capability cua`。选择豆包搜索后默认关闭目标客户端的原生 `WebSearch/web_search`，用户要求保留时加 `--keep-native-websearch`。发生关闭时立即输出作用域、配置文件和完整 reset 命令。只查看用法时运行 `arkcli helper mcp --help`。
- Helper 配置 Agent Plan 时只能使用该 profile 的专属/席位 Key，不能用普通 API Key 池兜底。个人版专属记录存在但非 Active（如 `Restricted`）时，真人 TTY 会先说明轮转将立即使旧 Key 失效并询问；确认后只轮转一次、最多读 3 次状态，第一轮 Active 即停止。无专属记录、用户拒绝或非 TTY 时不轮转；失败引导必须去 Agent Plan 使用配置页，不能去普通 API Key 页面或 `auth apikey`。
- `arkcli-auth` 只处理实际认证阻塞。用户仅要求查看 helper 的 `--help` 时，不要转去登录、`init-volc` 或其他认证命令。

## 适用场景 / 唤起信号

- 用户要给本机 Claude Code / Codex / OpenCode / OpenClaw / Hermes / ZCode 配置 Plan 或 Platform Endpoint 的 model/provider。
- 用户要给支持的目标 Agent 配置豆包搜索、专业数据集或 OpenViking MCP，只安装云电脑 CUA Skill，或配置 byted-supabase。
- 用户询问 `arkcli helper` 的安装态、配置态、文件落点、重载方式或 Agent Plan Harness 抵扣/超额后付费矩阵。

## 反唤起信号 / 不该唤起

- 只查询/购买/续费套餐或管理团队席位时转 `arkcli-plans`。
- 只搜索基础模型、查询用量或价格时分别转 `arkcli-models`、`arkcli-usage`、`arkcli-pricing`。
- 把 arkcli 内嵌 Skills 安装到 Agent 时转 `arkcli-connect`；不要把“安装 arkcli skill”误判为配置 Harness。
- 仅排查登录、401 或 profile 覆盖问题时转 `arkcli-auth` / `arkcli-config`，不要反复重写 Agent 配置。

## TTY Harness 能力矩阵（Agent Plan / Agent Plan Team）

真人直接运行 `arkcli helper` 时，向导顺序固定为：选择 Agent Plan profile → 选择模型 → 选择 AI Agent → 写入 model/provider → 进入 Harness 能力选择页。每行同时显示：

1. 本次是否重新配置（`[✓]` 已选、`[ ]` 未选、`[-]` 当前不可配置）；
2. 套餐抵扣是否开启；
3. 超额后付费是否开启；
4. 当前所选 AI Agent 的本地配置状态。

本次选择与当前状态互相独立：默认勾选全部可配置能力；已显示“已配置”的能力进入下一步后仍必须重新执行配置。仅远端能力、当前 Agent 不支持以及团队版不提供的 Agent 记忆显示 `[-]`，不能选择。用户可取消任意能力；全部取消后选择“进入配置”会被阻止，也可以显式选择“跳过能力配置”。

已知本地适配是精确映射，不按服务端名称做模糊猜测：

| Harness 能力 | 本地交付形态 | “已配置”的判定 |
|---|---|---|
| 专业数据集 | MCP | `dataPro-search` 已注入且使用真实 Plan Key |
| 豆包搜索 | MCP + Skill | 搜索 MCP 使用真实 Plan Key、`byted-web-search` Skill 已安装，且 Skill 根目录 `.env` 中存在可用的 `WEB_SEARCH_API_KEY` |
| Agent 记忆 | MCP | OpenViking control-plane 已注入且使用真实 Plan Key |
| AI Native 应用开发底座 | CLI + Skill | byted-supabase CLI、Skill、`ark_login` Agent Plan 登录态均就绪 |
| Agent 进化 | CLI + Skill | 官方 Evolve CLI、`evolve-setup` Skill、workspace 配置与凭证均就绪 |

服务端目录是远端权益状态的事实源：上面五个已知能力始终显示；当前套餐未返回映射时标明无远端权益、不开启远端切换，但仍可查看或补充本地配置。服务端新增但 arkcli 尚未适配的 Harness 显示在已知五项之后，标为“仅远端能力”，只允许查看/切换服务端权益，不提供猜测性的本地配置动作。

- 个人版直接管理账号开关。
- 团队版普通成员管理当前席位；管理员同时有席位时默认当前席位，也可选择企业账号范围。账号总闸关闭时，席位开关显示未开启并在详情中解释上级限制。
- 开启超额后付费属于真实计费写操作：CLI 必须展示确认文案，默认选“取消”。服务端下发协议/购买链接时一并展示。不得用 `--dry-run` 代替确认。
- Agent 进化只检测官方安装布局；当前不会自动执行未固定版本的远程 `latest/install.sh`，详情页提供官方命令让真人手动执行后刷新。
- 豆包搜索 Skill 下载期间持续显示 Spinner；真人可按 `Ctrl+C` 立即取消当前下载并只跳过豆包搜索剩余配置。此时已注入的搜索 MCP 保留，不写 Skill 凭证、不修改原生 WebSearch，并继续配置本次已勾选的 Agent 记忆等后续能力。该按键只属于 TTY 矩阵，不改变非交互命令语义。
- 强制路由只在能力配置完成后确认。Helper 从磁盘重新检测真实 Provider，动态列出可授权能力；零 Provider 时只能返回继续配置或不启用退出，不能安装空授权路由。
- `helper configure/mcp/supabase/list/reset` 的非交互契约不因矩阵改变；Agent 自动化仍应选最具体的子命令，不得尝试操纵 TTY 菜单。

## Agent Plan 强制检索路由（配置与运行时协议）

本 Skill 负责分类与调用协议；真正的强制性来自 Helper 安装到 Harness 的 `UserPromptSubmit` / `PreToolUse` gate。它为每轮注入新的 `turn_id`，并在受管 Provider 执行前阻断错误或未授权 Provider。原生 WebSearch、WebFetch、WebExtract 不属于受管 Provider，不进入该 gate。

仅 Volc `agent-plan` / `agent-plan-team` 与正式支持的 Claude Code、Codex、OpenCode、OpenClaw 可配置：

```bash
arkcli helper configure <claude-code|codex|opencode|openclaw> \
  --profile <agent-plan-profile> \
  --with-routing
```

`--with-routing` 隐含业务 MCP；Supabase 必须显式加 `--with-supabase`，且只有真实配置成功后才进入 Provider 集合。Hermes/Trae 缺少可验证的 hook + MCP 闭环时必须失败且保持 Harness 配置不变；Platform、Coding Plan 与非 Volc 产品不支持。

上下文出现 `ARKCLI ROUTING ENFORCEMENT (routing.v1)` 时，每轮必须：

1. 原样读取本轮 `turn_id` 与 `Available providers`，不得伪造或复用上一轮 ID。
2. 把混合问题拆成可验收 task，并按固定映射选择唯一 Provider：`latest_public_web -> doubao_search`、`structured_dataset -> datapro`、`persistent_knowledge -> openviking`、`application_data -> supabase`、本地计算或需澄清 -> `none`。
3. 在任何受管业务 Provider 前恰好调用一次 `arkcli-agentplan-search-router.submit_route_plan`，提交完整 `routing.v1` 计划；只调用原生 Web 工具时不提交计划。
4. 计划接受后只调用已授权且真实可用的 Provider；正确 Provider 不可用时说明缺失并等待用户重新配置，不得 fallback。

总领 Skill / 当前 Agent 负责理解用户意图，`arkcli-agentplan-search-router` 只接收计划并由 gate 执行授权，不是搜索服务，也不自行读取 Prompt。选择豆包搜索、DataPro、OpenViking 或 Supabase 时才进入 `routing.v1`；用户明确要求原生 WebSearch/WebFetch/WebExtract 时可直接调用，它们不消耗也不继承受管 Provider 授权。

豆包搜索的同 Provider 入口顺序固定：已加载时优先调用 `byted-web-search` Skill；Skill 未安装、用户明确要求 MCP，或 Skill 在**业务请求发出前**因本地命令/配置错误失败时，保留原始错误并只改用一次 `mcp-server-askecho-search-infinity`。业务请求已经发出后出现鉴权、超时或服务错误必须停止，不得换入口重试。不得自行提取、打印或向工具手工传递 API Key。

受管 Provider 之间禁止互相替代、先调用后补计划，也不得把路由 MCP 当作搜索 Provider。原生 WebSearch/WebFetch/WebExtract 位于该约束之外；但用户明确指定某个受管 Provider 时，不得在其失败后静默换成原生工具。Route Plan schema、混合请求与授权生命周期见 [`references/arkcli-helper.md`](references/arkcli-helper.md#routingv1-强制检索路由协议)。

`helper mcp` 是**不改模型的非交互 Harness 工具入口**。不传 `--capability` 时保持存量整组 MCP 行为：个人版 `agent-plan` 注入四台，团队版 `agent-plan-team` 只注入豆包搜索 + DataPro。传 `--capability` 时只配所选单项，不带入其他已开启工具。

## Platform Endpoint 配置

Platform profile 只负责配置 Agent 的 model/provider：`model` 写为用户选择的 Endpoint ID，base URL 使用 Platform 的 `/api/v3`，协议继续由各 harness 保持现有行为（Chat 或 Responses）。

- 只展示**当前 SSO 子用户创建**、`Running`、模型被明确验证为**文本输出**的 Endpoint；VLM（图文输入、文本输出）可用。
- 生图、生视频、生 3D、音频、Embedding、内容生成或未知模型一律不展示，也不能通过 `--model` 绕过。
- Agent 配置中的 `model` 仍写 Endpoint ID；context window、max completion tokens、输入/输出模态按 Endpoint 绑定的基础模型名，复用 Agent Plan / Coding Plan 现有的 ArkModels 元数据富化规则（ArkModels 权威数据优先；查不到或字段为零时用内置 Plan 模型能力表兜底，仍未知才省略扩展字段并在 helper configure 输出中提示）。元数据查询失败同样 best-effort，不阻断已通过资格校验的 Endpoint。
- Hermes Agent 支持把 Platform Endpoint 写成 `volcengine-platform` model/provider；仍不支持 MCP 注入。
- 该接入只为 Platform 增加元数据调用方，不修改 Plan 模型清单、默认模型、元数据查询、MCP 或 Supabase 行为。
- 没有自己创建的 Endpoint 时，向导打开 `https://ark.volcengine.com/region:cn-beijing/endpoint/create?agentMode=close`；创建完成后选择“刷新列表”。已有但未运行的 Endpoint 需先启动再刷新。
- Platform **不支持** MCP、OpenViking、Supabase 或强制路由；`--with-mcp`、`--with-supabase`、`--with-routing` 会报错。

非交互调用：

```bash
arkcli helper configure codex \
  --profile <platform-profile> \
  --endpoint <ep-id>
```

配置成功后，Endpoint 可按 OpenAI 兼容入口调用：`/responses` 或 `/chat/completions`，请求的 `model` 均使用该 `<ep-id>`。

## 注入的是哪几台 MCP(写死,勿幻觉)

| server | 传输 | key 来源 |
|--------|------|---------|
| `mcp-server-askecho-search-infinity`（豆包搜索） | stdio(`uvx`) | **Agent Plan 的 API Key**(env `ASK_ECHO_SEARCH_INFINITY_API_KEY`,与 dataPro / 控制面**同一把** plan key);取不到则写占位符 |
| `dataPro-search` | http(streamable) | **Agent Plan 的 API Key**(header `X-Agent-Plan-Key`,裸 key) |
| `openviking-dataplane`（**仅个人版 agent-plan**） | http(streamable) | **OpenViking 库的访问 key**（数据面；header `Authorization: Bearer <key>`）;经 vikingdb 两步取:列库 → 按库取 key。账号多库时要选库(见下)。账号下 0 个库时跳过 |
| `openviking-controlplane`（**仅个人版 agent-plan**） | stdio(`uvx`) | **Agent Plan 的 API Key**（控制面；env `AGENTPLAN_API_KEY`）;不依赖 OV 库列表,个人版有 Agent Plan 即注入 |

> **个人版 vs 团队版:** 上表 OpenViking 两台是**个人版 `agent-plan`** 专属。**团队版 `agent-plan-team` 与 OpenViking 无关** —— 只注入豆包搜索 + dataPro 两台,`openviking-dataplane` / `openviking-controlplane` 都不配,也不会去查 vikingdb。

## 豆包搜索成功后的原生 WebSearch 选择

仅当豆包搜索 MCP 被选择、`byted-web-search` Skill 安装成功，且所选 Agent Plan 的同一把 API Key 已写入 Skill 根目录 `.env` 的 `WEB_SEARCH_API_KEY` 后，`helper` 才处理目标客户端的原生 `WebSearch/web_search`。TTY 会让用户选择“关闭（推荐）/保留”；非交互命令默认关闭，传 `--keep-native-websearch` 时保留。普通 model/provider 配置、Platform、Coding Plan、显式跳过豆包搜索、MCP 注入失败、Skill 安装失败或 Skill 凭证写入失败都不触发。Helper 不修改第三方 Skill 源码，也不得打印该 Key。

- 范围严格限定为 `WebSearch/web_search`；**不会**关闭 `webfetch`、`web_extract`、`x_search`、浏览器或其它工具。
- 项目级配置从当前目录向上寻找 `.git`（worktree 的 `.git` 文件也支持）；找不到时使用当前目录。
- 配置成功后命令立即输出：客户端、作用域、配置路径，以及 `arkcli helper reset <harness>`。该 reset 是**完整回滚**，会同时移除 ArkCLI 写入的 model/provider/MCP，不是“只恢复搜索”。
- ArkCLI 只恢复自己拥有且未被用户再次修改的字段。用户原本已经关闭搜索时不接管；配置后同一字段被外部修改时，reset 保留用户当前值并警告。
- TRAE 写入 `.trae/hooks.json` 后，还必须在 `TRAE Settings > Hooks` 中启用当前项目；命令会明确提示“待激活”，不得宣称仅写文件就已生效。

## host ≠ target(最关键的概念)

- **host** = 你(这个 AI Agent)此刻跑在哪 —— 命令读环境变量自动检测,无需你判断。
- **target** = 要把 MCP 写进谁的配置 —— 可以是 host 自己,也可以是另一个 agent。
- 二者解耦:人在 OpenCode 里,也能给 Claude Code 配 MCP。

→ 用户在 prompt 里**点名了某个 agent**(如"给 opencode / codex 配"):跑 `arkcli helper mcp opencode` / `arkcli helper mcp codex`
→ 用户说"**当前 / 这个 Agent**"或没点名:跑 `arkcli helper mcp`(自动检测当前 host)

## 子命令穷举

| 调用 | 说明 |
|------|------|
| `arkcli helper mcp [target] [--capability <datapro\|web-search\|agent-memory\|cua>] [--profile P] [--ov-resource <库名>] [--scope project] [--keep-native-websearch] [--codex-config-scope profile\|global] [--codex-profile <name>] [--dsh-config-scope home\|profile] [--dsh-profile <name>]` | 不传 `--capability` 保持整组 MCP 旧行为；显式选择时只配一项，**不改 model**。`web-search` 还会安装 Skill 并按原规则处理原生 WebSearch；`agent-memory` 仅个人版，多库用 `--ov-resource`；`cua` 仅个人版 Medium/Large/Max，只给 target 安装 `byted-util-ark-cua` Skill，不写任何 MCP；Codex 的 MCP 默认写 profile `~/.codex/arkcli.config.toml`；DSH 默认写 home 层，profile 层需同时指定两个 DSH flag |
| `arkcli helper configure <harness> [--profile P] [--model M\|--endpoint ep-id] [--with-mcp] [--keep-native-websearch] [--with-supabase] [--with-routing] [--codex-config-scope profile\|global] [--codex-profile <name>]` | Plan 用 model/provider；Platform 必须用 `--endpoint` 选择文本 Endpoint。仅 Plan 可加 MCP/Supabase；强制路由仅 Agent Plan 可用。`--keep-native-websearch` 只影响豆包搜索配置，不改变路由 gate。 |
| `arkcli helper list` | 查支持的 agent + 安装/配置状态(只读) |
| `arkcli helper supabase [--profile P]` | **非 MCP**:装 byted-supabase-cli + skill + 注入火山登录态(打通 byted-supabase 数据库能力);跟 harness 无关。仅 Agent Plan(个人版全档 + 团队版全档) |
| `arkcli helper reset <harness>` | 完整移除 arkcli 注入的 model/provider/MCP/托管路由 hook/tool/plugin，并恢复 ArkCLI 拥有且未被外部修改的原生搜索字段 |
| `arkcli helper` | TTY 交互向导(需终端;非交互场景改用上面的);进入向导前会检查登录态,未登录/SSO 过期时按当前登录上下文拉起 SSO(火山走 volc-sso;全新用户无明确上下文时走 auth login 菜单),成功后继续向导；Agent Plan/Team 按 profile→model→Agent→能力多选→逐项配置推进，随后只基于已验真的真实 Provider 动态确认强制路由 |

> ⚠️ 想"只加 MCP" → `helper mcp`;想"把 agent 接到 plan、连模型一起" → `helper configure --with-mcp`。别用 `configure` 去只加 MCP(它会一并(重)写 model)。
>
> 🎯 用户说"**把(我 plan 的)全套 harness 工具都配上 / 都给我 set 好**"(MCP + Supabase 一次到位)→ `arkcli helper configure <harness> --with-mcp --with-supabase`。这是 agent 唯一能一条命令配齐 MCP 三件套 + Supabase 的路径(交互向导 `arkcli helper` 要 TTY、agent 跑不了;它的自动 SSO 只服务真人终端);`--with-mcp` 只配 MCP、不含 Supabase,想带 Supabase **必须显式加 `--with-supabase`**(资格不够 / 失败只 warn,不阻断 harness 配置)。
>
> 🧭 用户明确要“强制受管 Provider 路由 / routing.v1 / 防止豆包搜索、DataPro、OpenViking、Supabase 串路由”→ 使用本 Skill 的“Agent Plan 强制检索路由”协议，并执行 `helper configure <harness> --with-routing`。该 flag 隐含业务 MCP；Supabase 仍需显式加 `--with-supabase`，原生 Web 工具不受此 flag 管理。

## byted-supabase 数据库能力(`helper supabase`;**非 MCP**)

`arkcli helper supabase` 跟上面的 MCP 注入是**两类能力**:它**不写** agent 的 mcp.json,而是**装 byted-supabase-cli + skill** 并用当前火山 Agent Plan 登录态**注入登录态**(打通 byted-supabase / Volcengine Supabase 数据库平台)。**跟选哪个 harness 无关** —— 它配的是 byted-supabase-cli 这个独立工具本身。

- **门槛**:仅 Agent Plan —— 个人版 `agent-plan` **全档**支持(含 small/medium/large/max),团队版 `agent-plan-team` **全档**支持。不合格命令直接报错说明。
- **动作**:装 CLI(`npx -y @byted-supabase/cli@latest install`,连匹配的 byted-supabase agent skill 一起)→ 用所选 Agent Plan 身份的 STS + refresh_token 组装 Console Login 凭证 → `byted-supabase-cli login --credential-file`(个人版带 `--is-agent-plan`,团队版额外 `--agent-plan-seat-id <实时反查>`)注入到固定 profile `ark_login`。
- **触发**:用户说『配 byted-supabase / 打通数据库 / 装 supabase cli / 连接 supabase / 用我的 plan 连数据库』。
- **三条配置入口**(同一内核 `supabase.Configure`):① 专配 `arkcli helper supabase`;② 非交互/agent 顺带配 `arkcli helper configure <h> --with-supabase`(无确认框、失败只 warn,不阻断 harness 配置);③ 交互向导 `arkcli helper` 末尾可选步骤(仅合格 plan)。`helper mcp` / `configure --with-mcp` **不含** Supabase —— 想顺带配必须显式 `--with-supabase`。
- ⚠️ **v3 ve handoff 身份 (source=ve) 不能配 supabase**: `helper supabase` 内核 `gatherInputs` 强依赖 `LoadIdentityTokenFull(key)` 读 identity_store `token.json` 里的 refresh_token 组装 Console Login 凭证; 而 source=ve 身份**不落 token.json** (refresh_token 由 volcengine-go-sdk 内部管, arkcli 侧拿不到明文), 会报 `读 identity token: ...`。命中该报错时告诉用户: 想配 supabase 需要走 arkcli 原生 SSO 登录一次 (跳过 ve handoff), 具体做法是先 `ve logout` (或让 `ve` 处于未登录态) 再跑 `arkcli auth login` 拿到 `source="arkcli"` 的 identity。

## 范围边界(管好,别越界)

- **model/provider 可配置 target**:`claude-code` / `codex` / `deepseek-harness` / `pi` / `opencode` / `openclaw` / `hermes` / `zcode` / `workbuddy`。其中 Hermes 与 Pi 支持 Plan 和 Platform Endpoint，但都不支持 MCP;DSH 支持 Plan / Platform / Coding Plan 与 MCP 注入;ZCode 与 WorkBuddy 支持 model + MCP。
- **可注入 MCP 的 target 有 8 个**:`claude-code` / `codex` / `opencode` / `openclaw` / `trae` / `deepseek-harness` / `zcode` / `workbuddy`。CUA 是 Skill 而非 MCP，还可精确安装给 `hermes` / `pi`；命令只传单个 target，不扫描或改动其他 Agent。
- 本 skill 会被 `arkcli +connect` 装进很多 agent(cursor / gemini-cli / codex …40+),但 **MCP 注入只支持上一条列出的 8 个 target**。host 是其它 agent 时:要么用户点名其一作 target,要么命令会报"请显式指定" —— **绝不静默配错对象**。
- `codex` 支持 model/provider + MCP。默认 **profile 模式**写 `~/.codex/arkcli.config.toml`,需用 `codex --profile arkcli` 启动 terminal/TUI 才生效;传 `--codex-config-scope global` 才写 `~/.codex/config.toml`,该范围可能被 Codex CLI/TUI、Codex App、IDE extension 共享读取。
- `trae`(AI IDE)是 MCP-only:不配 model/provider;**无运行态宿主检测**(不会被自动当成 host),只能显式 `arkcli helper mcp trae`。默认写用户级 `~/.trae/mcp.json`,加 `--scope project` 写项目级 `./.trae/mcp.json`(项目级需在 Trae「设置 → MCP」开启「启用项目级 MCP」开关 + 重开项目)。豆包搜索成功后另写项目根 `.trae/hooks.json`，还需在 `TRAE Settings > Hooks` 启用当前项目。
- `hermes` 支持 Plan / Platform Endpoint 的 model/provider 配置，但暂不支持 MCP 注入 → MCP 请求命中就直说"暂不支持"。
- `deepseek-harness` (DSH):默认 home scope，可直接运行 `arkcli helper configure deepseek-harness --profile <plan>` 或 `arkcli helper mcp deepseek-harness`；若要写 profile scope，传 `--dsh-config-scope profile --dsh-profile <name>`，Helper 会确保 profile 目录存在。DSH provider/模型配置落到 `$DSH_HOME/settings.yaml` 的 `llm-pi-ai.providers.arkcli-<planType>:` 段（独立 provider,**不覆盖** `llm-deepseek:` 下的 DeepSeek 官方模型;用户可在 DSH 里切换官方 DeepSeek 与 arkcli 火山模型）；默认模型写 `agent-default-model:`;API key 写入 `$DSH_HOME/.credentials.yaml` version 1 文档的 `refs.ARKCLI_<PLAN>_API_KEY`，由 DSH credentials service 按 `apiKeyEnv` 解析（不覆盖用户的 `DEEPSEEK_API_KEY`）。这是 breaking change：检测到旧版 ArkCLI 的 `llm-deepseek:` 指纹、纯 flat credentials，或 version 1 文档顶层残留的 legacy credential key 时，`configure` 不自动迁移，而是拒绝写入并要求先运行 `arkcli helper reset deepseek-harness`；`reset` 兼容清理这些旧格式，但任意自定义 `apiKeyEnv` 仍视为用户配置。settings 与 credentials 作为一个可回滚事务更新；`reset` 只删除仍保持 ArkCLI provider fingerprint 的条目、对应凭据以及仍指向它的默认模型。MCP / 原生 WebSearch 关闭落在 `$DSH_HOME/profiles/<name>/cordis.patch.yml`(profile scope)或 `$DSH_HOME/cordis.patch.yml`(home scope)。DSH 不导出 env marker,`helper mcp` 不传 target 不会自动选 DSH。原生 WebSearch 用 patch 层对 `@deepseek-ai/dsh-tool-web` 加 `disabled: true`(trae 风格 hook 兜底,精确作用当前 profile,reset 时按 marker 还原)。
- `pi`(minimal terminal coding harness,`npm i -g --ignore-scripts @earendil-works/pi-coding-agent`)支持 Plan / Platform Endpoint 的 model/provider 配置,但**没有原生 MCP host 能力** → DataPro / 豆包搜索 / Agent 记忆请求命中就直说"暂不支持"。`helper mcp pi --capability cua` 是例外，因为 CUA 只安装 Skill、不注入 MCP；仍受个人版 Medium/Large/Max 资格闸约束。provider/模型清单写 `~/.pi/agent/models.json`(openai-completions provider,键=plan type),默认 provider/model 写同目录 `~/.pi/agent/settings.json` 的 `defaultProvider` / `defaultModel`(bare model id);目录可被 `$PI_CODING_AGENT_DIR` 覆盖。
- `zcode` 支持 Plan / Platform Endpoint 的 model/provider 配置与 MCP 注入。模型配置写 `$ZCODE_HOME/v2/config.json`（默认 `~/.zcode/v2/config.json`）；ZCode 不会热加载 provider registry，配置后必须完全退出 ZCode 主进程、重新打开并新建会话，不能只切模型或复用旧会话。遇到 Coding Plan `InvalidParameter` / token 上限 400 时不要手改某个模型的固定值，重新运行 `helper configure zcode`，由 Helper 写入权威 `limit.context` 并把 `limit.output` 转成方舟网关接受的十进制 token 上限；详见 reference。
- `workbuddy`(CodeBuddy IDE,VSCode 系):同时支持 model + MCP,**无运行态宿主检测**(不会被自动当成 host),只能显式 `arkcli helper configure workbuddy` / `arkcli helper mcp workbuddy`。model 与 MCP 同时写入**三个落点** `~/.workbuddy-ai/`(IDE 真正读取的运行态目录)、`~/.workbuddy/` 与 `~/.codebuddy/`:model 清单进各自 `models.json`(OpenAI 兼容,`url` 为含 `/chat/completions` 的完整路径,arkcli 写入项打 `arkcliManaged` marker 供精确 reset),MCP 进各自 `mcp.json`(`mcpServers.*`,schema 与 claude-code 同构)。**不支持原生 WebSearch 关闭**(无 hook seam),豆包搜索配置只注入 MCP、不接管原生搜索。
- 强制路由正式支持 `claude-code` / `codex` / `opencode` / `openclaw`；Hermes/TRAE 走能力门禁，缺 hook + MCP 闭环时零 Harness 配置写入并报错，不做提示词降级。
- Plan 的 model 可以配成路由模型 `ark-code-latest`(交互向导会先问「智能路由 (ark-code-latest) vs 指定具体模型」;非交互直接 `--model ark-code-latest`)。**切换 ark-code-latest 的路由目标(Auto 智能调度 / 锁定某个底层模型)不在本 skill** —— 走 `arkcli plans model-apply --plan <plan> --model <目标>`(见 [`../arkcli-plans/`](../arkcli-plans/SKILL.md)),写控制面、与控制台联动。

## 前提

- **必须有 Agent Plan 订阅**(豆包搜索 / dataPro 要 Agent Plan 的 key;OpenViking 两台是个人版专属)。命令自动定位账号下的 Agent Plan profile,**与当前 active profile 无关**;个人版 `agent-plan` 与团队版 `agent-plan-team` 都能注入,但**两者不等价**:个人版注入四台,**团队版 `agent-plan-team` 与 OpenViking 无关,只注入豆包搜索 + dataPro 两台**。没有就引导 `arkcli auth login` 开通;账号同时有多个 Agent Plan profile(如个人版 + 团队版)时，命令会列出本次检测到的 profile name；使用宿主提供的结构化选择能力让用户选一个，选定后带 `--profile <name>` 重跑，不得擅自偏向个人版、团队版或 active profile。
- 注入后 **agent 需重启**才会加载新 MCP。Codex profile 模式还需用 `codex --profile <name>` 启动;Trae 还需去「设置 → MCP」面板确认 MCP 已启用(项目级文件额外要开「启用项目级 MCP」开关)，并在 `Settings > Hooks` 启用当前项目后重开项目。

## OpenViking 库的选择(openviking-dataplane 专属;**仅个人版 agent-plan**)

> 本节只适用于**个人版 `agent-plan`**。**团队版 `agent-plan-team` 与 OpenViking 无关**,命令对团队版直接跳过下面整套列库/选库流程,只注入豆包搜索 + dataPro。

`openviking-dataplane` 的 key 绑定到某个 OpenViking 库(库名 ↔ ResourceID 1:1)。命令先列账号下的库,按数量分流:

- **0 个库** → 自动跳过 openviking-dataplane(仍注入另三台,包括 openviking-controlplane),并提示去 `https://console.volcengine.com/vikingdb/openviking/region:openviking+cn-beijing/create` 建库后重跑。可直接接受跳过。
- **1 个库** → 直接用,无需选择。
- **多个库** → 命令报错并列出所有库名(形如 `检测到多个 OpenViking 库,请用 --ov-resource <库名> 指定其一:[a, b, c]`)。把这次错误中返回的真实库名交给宿主提供的结构化选择能力；不要写死某个宿主工具名，也不要凭记忆补库。拿到选定库名后带 `--ov-resource <库名>` 重跑同一条命令；宿主没有结构化选择能力时退化为精简编号列表。
- 取 key 失败(非 0 库)→ openviking-dataplane 写占位符 `Bearer <OPENVIKING_KEY>`,提示用户手动替换。
- `openviking-controlplane` 不受上述影响:个人版只要有 Agent Plan 就始终注入(团队版不注入,见上)。

## 路由判断 / 反触发

- "给 Agent 配 MCP / 豆包搜索 / 联网搜索 / dataPro / web search" → `arkcli helper mcp`
- "只给 Codex 配某一项 Harness 能力" → `arkcli helper mcp codex --capability <datapro|web-search|agent-memory>`；只配云电脑 → `arkcli helper mcp codex --capability cua`
- "把 agent 指向我的 plan(连模型一起)" → `arkcli helper configure`
- "打开 helper 看 Harness 抵扣/超额开关和本地配置矩阵" → 真人 TTY 运行 `arkcli helper`；Agent 不自动驾驶该菜单
- "把全套 harness 工具都配上 / 一次配齐 MCP + Supabase" → `arkcli helper configure <h> --with-mcp --with-supabase`
- "启用强制受管 Provider 路由 / routing.v1 / 防止豆包搜索、DataPro、OpenViking、Supabase 串路由" → 使用本 Skill 的强制检索路由协议，执行 `arkcli helper configure <h> --with-routing`
- "配 byted-supabase / 打通数据库 / 装 supabase cli / 连接 supabase" → `arkcli helper supabase`(非 MCP,装 CLI+skill+注入登录态);或在配 harness 时顺带 `configure --with-supabase`
- 把 arkcli skills **安装**进 agent → 走 [arkcli-connect](../arkcli-connect/SKILL.md),与本 skill 无关
- 401 / 登录 / 鉴权失败 → 走 [arkcli-auth](../arkcli-auth/SKILL.md)
- 生图 / 生视频 → 走 [arkcli-gen](../arkcli-gen/SKILL.md)

详细 flag、输出样例、错误码、边界 case 见 [`references/arkcli-helper.md`](references/arkcli-helper.md)。

## References / 详细参考

- [`references/arkcli-helper.md`](references/arkcli-helper.md)：完整 flag、MCP/Skill/CLI 落点、Harness 矩阵状态机与错误处理。
- [`references/evals.md`](references/evals.md)：Trigger、反触发、写操作守卫和回归用例。
- [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)：所有 arkcli Skill 共用的认证、输出与安全约束。
