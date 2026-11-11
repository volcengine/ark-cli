# @volcengine/ark-plan-api

A [DeepSeek Harness (DSH)](https://github.com/deepseek-ai/deepseek-harness) cordis plugin that registers Volcengine Ark model provider routes on the native `llm-pi-ai` adapter. With the default DSH profile configuration, the routes appear in the model picker after install; see [DSH profile compatibility](#dsh-profile-compatibility) if your profile already configures `llm-pi-ai`.

火山方舟模型路由插件。在 DSH 默认 profile 配置下，安装后即可在原生模型选择器中看到方舟 Agent Plan、Coding Plan 与后付费模型路由；如果 profile 已经配置 `llm-pi-ai`，请先阅读 [DSH 配置兼容性](#dsh-profile-compatibility)。

## Features / 功能

- **Agent Plan** — `doubao-seed-evolving`, `doubao-seed-2.1-pro`, `deepseek-v4-pro` and more via the Agent Plan endpoint.
- **Coding Plan** — Coding-optimized models for both cn-beijing and BytePlus regions.
- **Postpaid (后付费)** — Direct pay-as-you-go access to Ark foundation models.
- **Native picker integration** — With the default profile configuration, providers show up alongside built-in ones in **Settings → Models**, no model IDs to memorize.
- **Anthropic Messages wire protocol** — Uses the same Messages API shape DSH uses for Anthropic, so tool use, vision and streaming work out of the box.

## Install / 安装

### From npm (recommended / 推荐)

```sh
npx -y @deepseek-ai/dsh plugin --profile web add @volcengine/ark-plan-api
```

### From GitHub Release

```sh
npx -y @deepseek-ai/dsh plugin --profile web add https://github.com/volcengine/ark-cli/releases/download/dsh-plugins%2Fv0.1.0/ark-plan-api.tgz
```

### Via awesome-dsh-plugin registry

```sh
npx -y @deepseek-ai/dsh plugin --profile web add volcengine/ark-cli#ark-plan-api
```

### From a local clone (development / 开发用)

```sh
npx -y @deepseek-ai/dsh plugin --profile web add ./dsh-plugins/ark-plan-api
```

Restart DSH after install, then open **Settings → Models**, pick an Ark provider and paste your API key. Each route reads its key from a separate env var (see table below); the Models UI writes the key to the correct name automatically.

安装后重启 DSH，打开 **Settings → Models**，选择 Ark provider 并粘贴 API Key。每条路由从独立的环境变量读取密钥（见下表），Models 设置页会自动写入对应变量。

## Providers / 提供的模型路由

| Route ID | Purpose / 用途 | Base URL | API Key Env |
|---|---|---|---|
| `ark-agent-plan-cn` | Agent Plan (cn-beijing) | `https://ark.cn-beijing.volces.com/api/plan` | `ARK_AGENT_PLAN_CN_API_KEY` |
| `ark-coding-plan-cn` | Coding Plan (cn-beijing) | `https://ark.cn-beijing.volces.com/api/coding` | `ARK_CODING_PLAN_CN_API_KEY` |
| `ark-coding-plan-byteplus` | Coding Plan (BytePlus/海外) | `https://ark.ap-southeast.bytepluses.com/api/coding` | `ARK_CODING_PLAN_BYTEPLUS_API_KEY` |
| `ark-cn` | Postpaid API 后付费 (cn-beijing) | `https://ark.cn-beijing.volces.com/api/v3/compatible` | `ARK_CN_API_KEY` |
| `ark-byteplus` | Postpaid API 后付费 (BytePlus/海外) | `https://ark.ap-southeast.bytepluses.com/api/v3/compatible` | `ARK_BYTEPLUS_API_KEY` |

Default wire protocol is the Anthropic Messages API. Base URLs intentionally omit the `/v1` suffix — pi-ai hands `baseURL` straight to the Anthropic SDK, which appends `/v1/messages`.

默认走 Anthropic Messages 协议。`baseURL` 已省略 `/v1` 后缀，pi-ai 会将其直接传给 Anthropic SDK，由 SDK 拼接 `/v1/messages`。

## DSH profile compatibility

DSH applies bundle patches before a profile's `cordis.patch.yml`. A row patch replaces the target row's entire `config`; it does not deep-merge it. If your profile already has an `llm-pi-ai` row with custom providers, that later row replaces this bundle's `providers` map and the Ark routes will not appear. Keep your existing providers and merge the Ark provider entries from this bundle's [`cordis.patch.yml`](./cordis.patch.yml) into the same `config.providers` map.

DSH 会先应用 bundle patch，再应用 profile 的 `cordis.patch.yml`。行补丁会整体替换目标行的 `config`，不会深度合并。如果 profile 已经有带自定义 provider 的 `llm-pi-ai` 行，后应用的配置会替换本 bundle 的 `providers`，Ark 路由因此不会显示。请保留现有 provider，并将本 bundle [`cordis.patch.yml`](./cordis.patch.yml) 中的 Ark provider 条目合并到同一个 `config.providers` 映射中。

## Model discovery / 模型列表

The plugin ships static model entries in `cordis.patch.yml`; use those entries in **Settings → Models**. DSH's “Fetch available models” flow is not supported across all Ark routes: the reported Agent Plan endpoint returns 404 for `/models`, and Coding Plan or postpaid discovery can return 401/404 even when inference works. Do not treat a model-discovery failure alone as evidence that an API key is invalid; select a bundled model entry instead.

插件在 `cordis.patch.yml` 中预置了模型条目，请直接在 **Settings → Models** 中选择。DSH「获取可用模型」操作并不适用于所有 Ark 路由：反馈显示 Agent Plan endpoint 的 `/models` 返回 404，Coding Plan 或后付费路由的模型发现也可能在推理正常时返回 401/404。不要仅凭模型发现失败判断 API Key 无效；请改用预置的模型条目。

## Requirements / 前置条件

- DSH ≥ 0.1.0
- Node.js ≥ 22.19.0 (ships with DSH)
- A Volcengine Ark API key with access to the corresponding endpoint

## License

Apache-2.0 © 2025-2026 Beijing Volcano Engine Technology Co., Ltd.
