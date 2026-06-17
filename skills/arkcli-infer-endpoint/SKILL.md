---
name: arkcli-infer-endpoint
version: 1.1.0
description: "arkcli 推理接入点管理能力：创建、获取、列表、启动、停止、更新推理接入点（Endpoint 全生命周期 CRUD + 启停）。优先使用产品命令 `arkcli infer endpoint ...`，而不是直接调用 Raw API。注意：如果用户只是想**一键把模型部署成新 Endpoint**，走 arkcli-deploy（`+deploy` 快捷方式）；本 skill 覆盖创建之外的获取/列表/启停/更新等管理操作。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli infer endpoint --help"
---

# arkcli infer endpoint

**CRITICAL — 开始前 MUST 先读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)。**

## 使用原则

- 推理接入点相关需求优先使用 `arkcli infer endpoint ...`
- 这些命令虽然是标准 CLI 类型，但实现入口仍然来自 `shortcuts/inferendpoint/`
- 只有产品命令无法覆盖时，才回退到 [`../arkcli-api-explorer/SKILL.md`](../arkcli-api-explorer/SKILL.md)
- `infer endpoint create` 成功后会返回 Endpoint `Id`
- 这个 `Id` 应作为后续 `get / start / stop` 的输入，也可以传给 [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md) 生成带真实 `endpoint-id` 的调用示例
- 如果已经通过 `infer endpoint create` 拿到 `Id`，不要再调用 `+deploy` 试图"二次部署"；`+deploy` 本身就是创建 Endpoint 的工作流
- `infer endpoint create --billing-method` 当前只支持 `token`；它是可选项，显式传 `token` 时会先校验模型是否支持 token 推理方式，创建请求本身保持默认行为

## 「我的接入点」语义

用户说**"我的推理接入点 / 我创建的 / 我有多少个 / 列出我的"** → 必须加 `--mine --page-all`：

```bash
arkcli infer endpoint list --mine --page-all --page-size 100 --format json
```

服务端按 `sys:ark:createdBy` tag 过滤，只返回当前 SSO sub-user 创建的 endpoint。  
需要 SSO 子账号登录；root 账号 / AK-SK 直接报错（引导重登）。  
详细行为见 [`references/arkcli-infer-endpoint-list.md`](references/arkcli-infer-endpoint-list.md)。

## 命令一览

| 命令 | 说明 |
|------|------|
| `arkcli infer endpoint create` | 创建推理接入点 |
| `arkcli infer endpoint get <endpoint-id>` | 获取推理接入点详情 |
| `arkcli infer endpoint list [--mine]` | 列出推理接入点；用户说**"我的 / 我自己的 / 我创建的 / 我有多少"**时必须加 `--mine`（SSO sub-user 过滤） |
| `arkcli infer endpoint start <endpoint-id>` | 启动推理接入点 |
| `arkcli infer endpoint stop <endpoint-id>` | 停止推理接入点 |
| `arkcli infer endpoint update <endpoint-id>` | 更新推理接入点（名称 / 描述 / 限流） |

## 参考

- [`references/arkcli-infer-endpoint-create.md`](references/arkcli-infer-endpoint-create.md)
- [`references/arkcli-infer-endpoint-get.md`](references/arkcli-infer-endpoint-get.md)
- [`references/arkcli-infer-endpoint-list.md`](references/arkcli-infer-endpoint-list.md)
- [`references/arkcli-infer-endpoint-start.md`](references/arkcli-infer-endpoint-start.md)
- [`references/arkcli-infer-endpoint-stop.md`](references/arkcli-infer-endpoint-stop.md)
- [`references/arkcli-infer-endpoint-update.md`](references/arkcli-infer-endpoint-update.md)
- [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md)
