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
   - 数据面 API Key 鉴权 / 权限错误（Runtime 调用） → 参考 [`../../arkcli-auth/references/auth-modes.md`](../../arkcli-auth/references/auth-modes.md) 的"API Key 模式的错误恢复"。**命令失败且症状=key 失效 / `InvalidApiKey` / 突然 401 且没主动换过 key**（疑似后端轮换）→ 先 `arkcli profile keys refresh` 同步后端 key 再重试一次（最轻的反应式自愈，agent 可自动跑）；**仅在失败时触发，不要每次命令前预防性 refresh**。refresh 救不了（权限不足 / env 覆盖 / SSO 过期）再 `arkcli auth apikey` 或重登
3. profile / Base URL / Region / API Key / Project 覆盖混乱
   - 普通诊断先用 `arkcli auth status --format json` 与 `arkcli auth whoami --format json` 查看当前身份/Profile 摘要；默认模型与路由转 `arkcli resources list --format json`。
   - 只有用户显式要求核对持久 Profile 时才执行 `arkcli profile show [name]`，并先说明它可能同步远端 Key、回写本地 Key 库存或默认 Key。脱敏输出不代表无配置副作用。
   - 根命令不再接受 `--region` / `--project-name`，`ARK_REGION` / `ARK_PROJECT_NAME` 也不覆盖运行时。
   - 若数据面使用 `--base-url`，必须同时显式提供 `--api-key`；无 Profile 的 stateless 模式必须将两者成对提供。
   - 两者仍对不上时，转 [`../../arkcli-config/SKILL.md`](../../arkcli-config/SKILL.md) 查看完整归因链路。
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
