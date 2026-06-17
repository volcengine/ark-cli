# 故障分流与能力边界

> 本文从 `arkcli-shared` 正文拆出。命令报错、不知道问题归类（认证 / 配置覆盖 / 参数 / 资源名 / 覆盖不足）时查阅。

## 故障分流

1. 先区分问题类型
   - 认证问题
   - 配置覆盖问题
   - 参数问题
   - 资源不存在或资源名不确定
   - 产品命令覆盖不足
2. 认证问题
   - 先看 `arkcli auth status`
   - 控制面鉴权错误（OpenTOP Action） → 转 `arkcli auth login`，并参考 [`../../arkcli-auth/references/auth-modes.md`](../../arkcli-auth/references/auth-modes.md) 的"控制面用 STS 还是长效 AK/SK"
   - 数据面 API Key 鉴权 / 权限错误（Runtime 调用） → 参考 [`../../arkcli-auth/references/auth-modes.md`](../../arkcli-auth/references/auth-modes.md) 的"API Key 模式的错误恢复"，先 `arkcli auth apikey`
3. profile / base-url / region / API Key / Project Name 覆盖混乱（两段式，别只看一处）
   - **看运行时生效值**：先 `arkcli auth status` —— 它给出**当前实际生效**的 `project_name` / `ark_api_key`（以及实名等）。flag / env / `.env` / identity store 这些**运行时覆盖**只会反映在这里。
   - **核对持久化字段**：再 `arkcli profile show [name]` —— 它 `SkipConfig`，**只读 `profile.yaml`** 的持久字段（type / region / project / resources / 派生 base_url），**不解析** `--project-name` / `ARK_PROJECT_NAME` / `.env` / identity store 覆盖。**别把它当 resolved 视图**，否则会顺着错线索查。
   - 两者对不上 = 有运行时覆盖在生效；进一步归因转 [`../../arkcli-config/SKILL.md`](../../arkcli-config/SKILL.md) 的配置归因链路。
   - 需要切 profile 时再用 `arkcli profile use <name>`
4. 模型、接入点、任务 ID 等资源名不确定
   - 优先走对应查询命令，例如 `arkcli models search`
5. 产品命令确实不覆盖
   - 先 `arkcli <domain> --help`
   - 再看对应 skill
   - 最后才转 [`../../arkcli-api-explorer/SKILL.md`](../../arkcli-api-explorer/SKILL.md)

## 当前能力边界

- 当前仓库已经覆盖：`models`、`+chat`、`+gen`、`+deploy`、`infer endpoint`、`usage stats`、`+connect`
- Endpoint 管理型命令已有独立 skill [`../../arkcli-infer-endpoint/SKILL.md`](../../arkcli-infer-endpoint/SKILL.md)，支持 `create/list/get/start/stop/update`
- 当用户说"先看看有没有现成 Endpoint"时，用 `arkcli infer endpoint list` 查询，不要假设必须先 `+deploy`
- 如果用户目标是正式接入且列表为空，再转 `+deploy` 创建
