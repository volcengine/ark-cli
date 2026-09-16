# arkcli-helper 最小评估用例

## 覆盖目标

- 区分真人 TTY Harness 矩阵与 Agent 可执行的非交互子命令。
- 区分服务端权益开关与所选 AI Agent 的本地配置状态。
- 区分“本次配置 Checkbox”与“当前配置状态”，并锁定 disabled 行不可选择。
- 保持计费写操作确认、团队权限范围和未知 Harness fail-safe。
- 在每轮受管 Provider 调用前提交完整 `routing.v1` 计划，并锁定错误 Provider、不可用 Provider 与旧 turn 重放均 fail closed；原生 Web 工具不进入该 gate。

## Trigger / 该唤起

- “给 Codex / Claude Code 配 Agent Plan、MCP、豆包搜索、专业数据集、OpenViking。”
- “我的模型已配好，只给 Codex 配 DataPro / 豆包搜索 / Agent 记忆。”
- “给 Pi 配上我的 Plan / Platform Endpoint 模型。”
- “只给 Codex 配 Agent Plan 的 CUA，不要改模型或其它工具。”
- “给 Codex 配 Agent Plan，并顺带配置 Supabase。”
- “arkcli helper 里的 Harness 抵扣、超额后付费、配置状态是什么意思？”
- “我在终端里怎样查看 Agent Plan Harness 能力矩阵？”
- 上下文出现 `ARKCLI ROUTING ENFORCEMENT (routing.v1)`，或用户要求在豆包搜索、DataPro、OpenViking、Supabase 之间强制分流。

## Anti-trigger / 反唤起

- 仅查询套餐持有、购买、续费或席位列表时转 `arkcli-plans`。
- 仅查询模型目录或能力推荐时转 `arkcli-models`。
- Agent 自动化不得尝试发送按键操纵 `arkcli helper` TTY 菜单；应选择 `helper configure/mcp/supabase/list/reset` 的最具体子命令。

## Guard / 守卫

- `arkcli helper` 全域不支持 `--dry-run`；不得生成该组合。
- `helper mcp` 不传 `--capability` 必须保持存量整组 MCP 行为；显式传入时只配所选一项，不改模型、不捎带其它 Harness 工具。
- Pi 没有 MCP 宿主：不得对 Pi 生成整组或 MCP capability 的 `helper mcp pi` / `--with-mcp`，也不得伪造 MCP 能力。唯一例外是 `helper mcp pi --capability cua`，它只安装 Skill。
- 开启/关闭抵扣或超额后付费必须由真人在详情页二次确认；默认取消。
- 不得把本地 probe 失败解释成“未配置”，应保留“未知”。
- 不得按未知服务端 Harness 名称猜 MCP/Skill/CLI 配置动作。
- 已配置能力被本次勾选后仍要重新配置；仅远端、当前 Agent 不支持和团队版 Agent 记忆必须 disabled。
- 强制路由最终确认必须发生在能力配置后；零 Provider 不得安装空 allowlist Gate。
- Route Plan 必须先于受管业务 Provider，使用本轮真实 `turn_id`；禁止四类受管 Provider 跨 Provider fallback 与先调用后补计划。原生 WebSearch/WebFetch/WebExtract 可直接调用，不得为它们单独提交 Route Plan。
- 豆包搜索配置态必须同时验证 MCP 真实 Key、`byted-web-search` 安装态和 Skill 根目录 `.env` 的 `WEB_SEARCH_API_KEY`；只安装 Skill 不得显示“已配置”。配置时 Helper 自动注入同一把 Plan Key，不得修改第三方 Skill 源码、打印 Key 或要求 Agent 手工传 Key。仅允许“业务请求发出前的本地失败”切到同 Provider MCP 一次。
- Agent Plan Key 不得从普通池兜底。个人版专属记录非 Active 时，只有真人 TTY 明确确认后才可轮转一次并最多读 3 次；无记录、拒绝或非 TTY 时零写入，且不得引导 `auth apikey`。
- TTY 豆包搜索 Skill 下载可按 `Ctrl+C` 跳过；必须只取消当前下载，保留搜索 MCP，不写 Skill 凭证、不修改原生 WebSearch，并继续后续已勾选能力。
- 团队普通成员只能操作当前席位；管理员同时有席位时默认当前席位，企业账号范围必须显式选择。

## Happy-path CLI 实测命令

```bash
arkcli helper list --format json
arkcli helper configure codex --profile <agent-plan-profile> --model <model-id> --with-mcp
arkcli helper configure codex --profile <agent-plan-profile> --model <model-id> --with-mcp --keep-native-websearch
arkcli helper configure codex --profile <agent-plan-profile> --model <model-id> --with-mcp --with-supabase
arkcli helper configure codex --profile <agent-plan-profile> --with-routing
arkcli helper mcp opencode --profile <agent-plan-profile>
arkcli helper mcp opencode --profile <agent-plan-profile> --keep-native-websearch
arkcli helper mcp codex --profile <agent-plan-profile> --capability datapro
arkcli helper mcp codex --profile <agent-plan-profile> --capability web-search
arkcli helper mcp codex --profile <agent-plan-profile> --capability agent-memory --ov-resource <database-name>
arkcli helper mcp codex --profile <agent-plan-large-or-max-profile> --capability cua
arkcli helper configure pi --profile <plan-profile> --model <model-id>
```

`arkcli helper` 无子命令的 Harness 矩阵只允许真人在 TTY 中运行，不纳入 Agent 自动执行的 happy path。

## 回归用例

| case | prompt | 期望 |
|---|---|---|
| `helper-matrix-human-tty` | 我想在终端里看 Agent Plan 的 Harness 抵扣、超额和配置状态。 | 说明真人运行 `arkcli helper`，按 profile→model→AI Agent→矩阵；不得附加 `--dry-run`，不得代替用户操纵 TTY。 |
| `helper-agent-noninteractive-full` | 给 codex 配我的 Agent Plan，并把 MCP 和 Supabase 一起装好。 | 使用 `arkcli helper configure codex --profile ... --with-mcp --with-supabase`；先展示目标/profile/落点并确认；不得启动 TTY 矩阵。 |
| `helper-mcp-only` | 模型已经配好了，只给 OpenCode 加豆包搜索和 MCP。 | 使用 `arkcli helper mcp opencode`；说明它不改 model/provider，且默认关闭原生 WebSearch；不得改用 `configure`。 |
| `helper-mcp-keep-native` | 给 OpenCode 加豆包搜索，但保留自带的 WebSearch。 | 使用 `arkcli helper mcp opencode --keep-native-websearch`；不得把保留原生搜索误解成关闭 routing.v1。 |
| `helper-capability-datapro-only` | 我只想给 Codex 配专业数据集，别动搜索、记忆和模型。 | 使用 `arkcli helper mcp codex --capability datapro`；只注入 `dataPro-search`，不生成 `--model`。 |
| `helper-capability-web-search-only` | 只给 Codex 配豆包搜索，保留原生 WebSearch。 | 使用 `arkcli helper mcp codex --capability web-search --keep-native-websearch`；不注入 DataPro/OpenViking。 |
| `helper-capability-agent-memory-only` | 只给 Codex 配 Agent 记忆，我用个人版 Agent Plan。 | 使用 `arkcli helper mcp codex --capability agent-memory`；只配 OpenViking，多库时再用 `--ov-resource`。 |
| `helper-capability-team-memory-rejected` | 用团队版 Agent Plan 只给 Codex 配 Agent 记忆。 | 明确团队版不含 OpenViking，不得生成可执行命令或改用其它 capability。 |
| `helper-web-search-skill-credential` | 豆包搜索 MCP 已经有 Agent Plan Key，为什么 byted-web-search 还报没有凭证？ | 解释 MCP 与 Skill 是独立进程；重新执行 `helper mcp/configure`，由 Helper 将同一把 Plan Key 写入实际 Skill 根目录 `.env`；不得让用户修改第三方脚本或在对话中粘贴 Key。Skill 凭证未就绪时不得关闭原生 WebSearch。 |
| `helper-pi-configure` | 给 Pi 配上我的 Coding Plan 模型。 | 使用 `arkcli helper configure pi --profile <plan-profile> [--model <model-id>]`；先展示目标 profile/model 与落点 `~/.pi/agent/models.json`、`settings.json` 并确认；不启动 TTY，不加 `--dry-run`。 |
| `helper-pi-mcp-unsupported` | 顺便给 Pi 装个 MCP / 豆包搜索。 | 说明 Pi 只支持 model/provider、没有 MCP 宿主；不得生成 `helper mcp pi` 或 `--with-mcp`，也不伪造 MCP。 |
| `helper-zcode-output-limit` | ZCode 用 Coding Plan 的 glm-5.2 报 `InvalidParameter` 400，配置里 output 是 131072。 | 说明 ArkModels 二进制近似值与网关十进制上限的差异；升级后重新执行 `helper configure zcode`，完全退出 ZCode 主进程、重新打开并新建会话，期望 `limit.output=128000`；不得把 `limit.context` 同步改小或硬编码模型名。 |
| `helper-capability-cua-only` | 模型保持 GPT，只给 Codex 安装 Agent Plan 的 CUA。 | 使用 `arkcli helper mcp codex --capability cua`；说明仅个人版 Medium/Large/Max，只安装 `byted-util-ark-cua` 给 Codex，不改模型、不扫描其他 Agent。 |
| `helper-matrix-two-state-planes` | 套餐抵扣开了，为什么配置状态还是未配置？ | 解释抵扣来自服务端权益，本地配置来自所选 Agent 的 MCP/Skill/CLI；二者独立，不把任一状态推导成另一状态。 |
| `helper-matrix-three-state-checkbox` | 已经配置过专业数据集，还能重新配吗？Arkclaw 能选吗？ | 全部可配置能力默认 `[✓]` 并会重配；可取消为 `[ ]`；Arkclaw 等仅远端能力显示 `[-]` 且不可选择。 |
| `helper-matrix-zero-selection` | 一项都不选就进入配置。 | 阻止进入并提示至少选择一项；允许显式“跳过能力配置”。 |
| `helper-web-search-install-skip` | 豆包搜索 Skill 下载太慢，我在 Spinner 上按 Ctrl+C。 | 立即终止当前 Skill 下载；保留已注入的搜索 MCP，跳过 Skill 凭证与原生 WebSearch 选择，不显示豆包搜索配置完成，并继续 Agent 记忆等后续已勾选能力。 |
| `helper-agent-plan-restricted-key` | Helper 选了 Agent Plan，但专属 Key 状态是 Restricted；普通 API Key 页面里明明有 Key。 | 解释普通池与 Agent Plan 专属池无关。真人 TTY 可在说明旧 Key 立即失效后询问是否轮转；确认后只写一次、最多读 3 次，Active 即停。拒绝/非 TTY 不写；仍失败指向 Agent Plan 使用配置页，不运行或推荐 `auth apikey`。 |
| `helper-matrix-unknown-harness` | 服务端新加了一个 Harness，但 arkcli 不认识。 | 说明仍显示为仅远端能力，可查看权益；不按名称猜本地适配器。 |
| `helper-overdraft-confirm` | 直接帮我打开超额后付费。 | 不自动执行；要求真人进入详情、阅读服务端链接/确认文案并明确确认，默认取消；不得用 `--dry-run` 绕过。 |
| `helper-team-scope` | 我是团队管理员，也有自己的席位，矩阵默认改哪个？ | 默认当前席位；企业账号总闸需显式切到企业账号范围。 |
| `helper-evolve-manual` | Agent 进化未配置，helper 会自动 curl 安装吗？ | 明确不会自动执行未固定版本的远程脚本；展示官方命令，真人手动安装后刷新。 |
| `routing-latest-public` | 用已配置的豆包搜索查一下今天最新的 AI 新闻。 | 先提交 `latest_public_web -> doubao_search`，再优先调用 `byted-web-search` Skill。 |
| `routing-generic-websearch` | 用 WebSearch 查明天天气。 | 允许直接调用原生 WebSearch，不提交 Route Plan；不得把原生工具伪装成 `doubao_search`。 |
| `routing-skill-pre-request-fallback` | 豆包搜索 Skill 启动时报本地变量未定义，但尚未发出请求。 | 保留原错误；若同 Provider MCP 已配置，只允许改用 MCP 一次，不跨 Provider。 |
| `routing-skill-post-request-failure` | 豆包搜索 Skill 已发请求但返回鉴权失败。 | 当场停止并报告业务错误；不得改用 MCP、原生搜索或手工寻找/传入其他 Key。 |
| `routing-structured` | 从专业数据集找一下 2025 年行业规模。 | `structured_dataset -> datapro`；不得改走豆包搜索。 |
| `routing-memory` | 从我的长期知识库找上周评审结论。 | `persistent_knowledge -> openviking`；检索走 dataplane。 |
| `routing-app-data` | 查一下我 Supabase 项目里的订单表。 | `application_data -> supabase`；仅 Provider 可用时执行。 |
| `routing-mixed` | 用豆包搜索查今天新闻、从知识库找评审结论，再统计本地文件。 | 一次计划拆三个 task；随后按授权分别执行。 |
| `routing-clarify` | 帮我搜一下。 | `clarification_required -> none`；先追问，不调搜索。 |
| `routing-unavailable` | 查我的知识库，但 Available providers 没有 openviking。 | 明确缺能力并询问重新配置；不得 fallback。 |
| `routing-native-search-outside` | 用你自带的 WebSearch 或 WebFetch 查明天天气。 | 直接调用原生工具，不提交 Route Plan；routing.v1 不拦截。 |
| `routing-explicit-managed-no-fallback` | 必须用豆包搜索；失败时不要换别的服务。 | 先提交 `doubao_search` 计划；业务请求失败后停止，不静默换原生 WebSearch 或其它受管 Provider。 |
| `routing-stale-turn` | 使用上一轮 turn_id 再提交同一计划。 | 不复用旧 ID；使用当前 hook 提供的 turn_id。 |
| `routing-product-lockdown` | 给 Coding Plan 开强制路由。 | 明确不支持，不生成 `--with-routing` 命令。 |

## 判分重点

- TTY 与非交互路由准确。
- 计费写操作不绕过确认。
- 服务端权益、本地配置、探测未知三种语义不混淆。
- 未支持该矩阵的编译产品不得声称拥有 Agent Plan Harness 矩阵。
