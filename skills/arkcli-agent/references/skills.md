# Skills

## Skill 选择

- 创建 Agent 的默认 Skill 选择顺序是 custom 优先、market 兜底。只要用户没有明确指定 market，先查本账号 custom skill；custom 没有合适候选后再搜索 market。用户明确指定 market 时可以跳过 custom。
- Market skill：custom 没有命中或用户明确要求 market 时使用 `agent skill search/list`，创建时优先直接使用返回的 `AgentSkills`；SkillHub V1 列表不返回版本信息，缺省传 `{type: skill_hub, skill_id: <id>}`，由服务端解析最新可用版本。
- Custom skill：本账号已有 custom skill 的默认选择流程是按需分页：先取第一页 100 条，由调用 arkcli 的 AI agent 按名称、描述和能力判断；没有合适候选时，再使用返回的 `NextPage` 继续取下一页，直到命中或没有下一页。不要一开始用用户描述做服务端 `Name` 过滤，也不要默认一次性拉完整 catalog。CLI 内部入口叫 `ListSkillsForTop`，实际 TOP Action 是 `ListSkills`。只使用该接口返回的 `skill-...` ID；同接口返回的 `s-...` 不作为 custom skill 使用。
- 创建 Agent 时可直接传裸 custom ID：`--skill skill-xxx`。CLI 会省略 `Version`，由服务端始终解析最新版本；不会把 `latest` 字面量发送给服务端。裸 `--skill s-xxx` 会被拒绝；`s-...` 只从 SkillHub 链路来。
- Custom skill：用户给本地 zip 时，用 `agent skill create --zip` 或 `agent agent create --skill-zip`，返回后作为 `{Type: custom, SkillId, Version}` 注入。
- Custom skill 上传走数据面 `POST /api/v3/skills`，不是 TOP `CreateSkill`；需要当前 profile 有可用 ARK API Key。
- `agent skill list/search` 支持 `market`、`custom`、`ark`、`all`：`market` 走公开 SkillHub；其他三种走 TOP `ListSkills`，其中 `all` 不发送 `Source`。`get` 和 `versions` 可读取 custom/ark；Ark Skill 是只读的，update/delete/download/set-protection 会在 CLI 侧拒绝。
- TOP Skill 摘要保留 `DisplayTitle`、`Source`、`UpdateTime`、`Name`、`ProtectionEnabled`。Agent 挂载 Skill 的响应保留 `UseLatest`、具体 `Version` 和 `DisplayName`，不要把实际版本和“是否跟随最新”混为一谈。
- `agent skill create --protection-enabled` 可在上传时设置保护开关；已有 custom Skill 用 `agent skill set-protection <id> --enabled=true|false`。该命令允许显式传 `false`，并调用 TOP `UpdateSkill`。
- `agent skill update <id> --zip <file>` 不是原地修改旧版本，而是调用 OpenTOP `CreateSkillVersion` 上传新版本；可用 `agent skill versions <id>` 查看版本，再用 `agent skill download <id> --version <version>` 下载指定版本。
- `agent skill download` 先调用 TOP `GetSkill`（未指定版本时取 `LatestVersion`），再调用 `GetSkillVersion` 获取 `PreSignedTOSURL`，最后用不携带 ARK 凭证的普通 GET 下载到本地；默认文件名为 `<skill-id>-v<version>.zip`。
- `agent skill download ... --dry-run` 会零网络预览版本解析、下载和本地保存步骤。未显式传 `--version` 或 `--output` 时，必须把 `unresolved` 中的版本/文件名当作占位符，不能声称已在线解析，也不会创建或覆盖文件。
- `agent skill delete <id>` 调用 TOP `DeleteSkill`；它不会级联删除 SkillVersion，只能在全部版本都已删除后执行。
- `agent skill delete <id> --version <version>` 调用 TOP `DeleteSkillVersion`，只删除指定版本。删除 latest 版本前，必须先删除全部非 latest 版本。
- `agent skill create/update/delete` 都支持命令级 `--dry-run`，用于零网络预览上传或删除请求。删除预览不要求 `--yes`，但不会在线校验版本依赖顺序；真实删除仍必须先读取完整版本列表，并在交互确认后执行或显式传 `--yes`。
- Market skill 的 `agent skill search/list --source market` 直接调用无需鉴权的 SkillHub V1 `GET /v1/skills`，并在 CLI 内映射为原有 `ListMarketSkills` 输出；不得向 `skills.volces.com` 发送 Ark STS、AK/SK 或 API Key。custom skill 的 `agent skill list --source custom` 仍调用 TOP `ListSkills`。
- 选择 market skill 时看 `Items[].Name`、`Description`、`Keywords`、`EvaluationScore`；创建入参直接取返回的 `AgentSkills`，或手动组 `{Type: skill_hub, SkillId}`。只有从其他可信接口明确拿到语义版本时才补 `Version`，不得从 `Metadata` 猜测版本。
- 选择 custom skill 时先读取第一页完整 `Items`，综合比较 `Id`、`Name`、`Description`、`LatestVersion`，不要只看第一条或只依赖服务端关键词匹配；如果没有足够相关的候选，再用 `NextPage` 继续读取下一页。`AgentSkills` 只会包含 `skill-...` custom skill。确认候选后可直接取 `AgentSkills`，或传裸 `--skill <Items[].Id>`。
- 多个候选接近时，使用宿主结构化选择能力展示本轮结果中的 `Id`、`Name`、`Description`、`LatestVersion`，宿主不支持时才退化为精简编号列表。用户明确要求自动完成时，只有相关度和版本证据能形成清晰第一名才自动选择并说明依据；否则仍需澄清，不能用“自动完成”授权任意挑选。通用数据分析优先考虑覆盖面广的分析 skill，再按需要追加 Excel/CSV/报表/可视化类 skill。
- 分页选择优先使用“按需读取”：第一页使用 `--limit 100`，后续把上一次响应的 `NextPage` 原样传给 `--page`。只有用户明确要求完整清单、需要离线分析全部 Skill，或多页都没有命中时，才使用 `--page-all`；它受全局 `--page-limit` 限制，不能把截断的 catalog 当成完整候选集。
- SkillHub V1 列表没有 `LatestVersionStatus`；创建 Agent 时 `SkillId` 使用 `Items[].Id`，默认省略 `Version` 以跟随服务端最新版本。

## 从 GitHub 导入

GitHub 导入固定使用公开仓库默认分支 HEAD，不支持私有仓库、branch 或 tag 参数：

```bash
arkcli agent skill github scan owner/repo --format json
arkcli agent skill github import https://github.com/owner/repo --path skills/foo --path skills/bar --format json
arkcli agent skill github import 'npx skills add owner/repo' --all --protection-enabled=true --format json
arkcli agent skill github import owner/repo --selected @./selected.json --format json
```

- Agent 构造命令时，优先使用服务端已知的输入形式：`owner/repo`、完整 GitHub HTTPS URL，或完整的 `npx skills add ...` 字符串。这些是调用指引，不是 Agent 或 CLI 的输入白名单。
- 用户给出仓库输入后，不要在 Agent 或 CLI 侧解析、改写或拒绝；应原样传给 `ScanGithubSkill.Input`，由服务端负责解释和校验。CLI 绝不启动 shell 或执行输入中的 npx。
- `--path` 可重复，或用 `--all`；两者互斥，并且只能选择 scan 返回的有效候选。
- `--selected` 文件是 `{Path, ProtectionEnabled}` 数组；它和全局 `--protection-enabled` 互斥。
- scan 是只读命令；import 是 `WorkflowClientPreview`，`--dry-run` 只展示 scan、选择和 import 请求，占位符不会发网络。
- import 会输出每个条目的 `ok`、`updated`、`skipped_duplicate` 或 `failed`；只要任一条失败，命令在保留完整结果后返回非 0。

## Custom Skill 删除顺序

Skill 删除有严格的依赖顺序。执行任何删除前，先用 `agent skill versions <id>` 获取完整版本列表并确认 latest 版本：

1. 删除整个 Skill：先逐个删除全部非 latest SkillVersion，再删除 latest SkillVersion；确认版本列表为空后，最后删除 Skill。
2. 删除 latest SkillVersion：先逐个删除全部非 latest SkillVersion；确认只剩 latest 后，才能删除 latest。
3. 删除单个非 latest SkillVersion：可以直接删除指定版本，但删除后仍要重新查询版本列表确认结果。

每次删除后重新执行 `agent skill versions <id>` 回查。不要假设 `DeleteSkill` 会自动级联删除版本，也不要在仍存在 SkillVersion 时调用它。每个破坏性命令单独执行，并遵守交互确认或 `--yes` 规则。

## 常用命令

```bash
arkcli agent skill search "Excel 数据分析" --limit 10 --format json
arkcli agent skill list --query "Excel 数据分析" --limit 20 --format json
arkcli agent skill list --source custom --name "我的数据分析" --limit 20 --format json
arkcli agent skill list --source custom --limit 100 --format json
# 若上一页返回 NextPage 且没有合适候选，再继续：
arkcli agent skill list --source custom --limit 100 --page '<NextPage>' --format json
arkcli --page-all --page-limit 20 agent skill list --source market --format json
arkcli agent skill create --zip ./my-skill.zip --display-title "My Skill" --format json
arkcli agent skill create --zip ./my-skill.zip --display-title "My Skill" --protection-enabled=true --format json
arkcli agent skill set-protection skill-xxx --enabled=false --format json
arkcli agent skill list --source ark --limit 100 --format json
arkcli agent skill list --source all --limit 100 --format json
arkcli agent skill github scan owner/repo --format json
arkcli agent skill github import owner/repo --all --format json
arkcli agent agent create --name arkcli-local-skill-agent --model <items[].model from agent model list> --skill-zip ./my-skill.zip --format json
arkcli agent agent create --name arkcli-existing-custom-skill-agent --model <items[].model from agent model list> --skill skill-xxx --format json
arkcli agent skill update skill-xxx --zip ./my-skill-v2.zip --format json
arkcli --page-all agent skill versions skill-xxx --format json
# 将占位符替换为上一条 versions 响应中选定条目的 Version 原值；不要自行拼版本号：
arkcli agent skill download skill-xxx --version '<Version-from-versions>' --output ./skill-selected-version.zip --format json
# 用户只需最新版时省略 --version，由 CLI 解析 LatestVersion：
arkcli agent skill download skill-xxx --output ./skill-latest.zip --format json
# 删除整个 Skill 时，先逐个删除全部非 latest 版本：
arkcli agent skill delete skill-xxx --version <non-latest-version> --yes --format json
# 回查并确认只剩 latest 后，再删除 latest：
arkcli --page-all agent skill versions skill-xxx --format json
arkcli agent skill delete skill-xxx --version <latest-version> --yes --format json
# 回查并确认版本列表为空后，最后删除 Skill：
arkcli --page-all agent skill versions skill-xxx --format json
arkcli agent skill delete skill-xxx --yes --format json
```

`agent skill list --source custom` 默认先输出一页 custom catalog 和 `AgentSkills`。调用 arkcli 的 AI agent 应先判断这一页；没有合适候选时沿 `NextPage` 继续请求，不要臆造 `SkillId`，也不要在当前页没有相关性判断时直接选第一条。需要完整 catalog 时再加 `--page-all`。

`agent skill search --source custom "<name>"` 仍可用于用户明确指定 skill 名称、目录很大需要缩小候选，或对全量 catalog 做二次检索；它不是自然语言创建 Agent 时的默认入口。

`agent skill list --page-all` 会自动使用每页 100 条：custom skill 沿 TOP `NextPage -> Page` 拉取；market skill递增 `PageNumber`，并根据 `TotalCount` 停止。全局 `--page-limit` 仍是最多请求页数。按需分页时，custom skill 的 `--page` 对应 TOP 的 `Page` token。
