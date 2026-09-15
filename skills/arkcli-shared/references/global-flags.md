# 常用全局 flags

> 各命令自身的 flag 以 `arkcli <domain> <verb> --help` 为准。`--region` 与
> `--project-name` 已从根命令删除；Region/Project 由所选 profile 决定。

| Flag | 作用 |
|------|------|
| `--profile` | 选择本次调用的持久上下文；控制面和数据面都可使用 |
| `--api-key` | 仅覆盖本次数据面调用的 ARK API Key |
| `--base-url` | 仅覆盖本次数据面调用的 API Base URL |
| `--env` | 选择控制面环境：`prod` 或 `stg` |
| `--format` | 输出格式 |
| `--transform` | 对输出做路径提取 |
| `--page-all` | 自动翻页 |
| `--page-limit` | 限制翻页次数 |
| `--page-delay` | 翻页间隔（ms） |
| `--debug` | 输出调试信息到 stderr |

`--dry-run` 不是全局 flag。它只由支持 Client Preview 的叶子命令注册；以该
叶子命令的 `--help` 为准。支持时只在客户端生成 `preview.v1`，不联网、不写文件、
不启动子进程；不支持时传入该 flag 会明确报错，禁止静默忽略。

## Volc 方舟体验入口

`arkcli exp` 是给人类终端使用的 Volc 方舟体验入口，不是全局 flag。已安装时直接
启动 `arkexp`；未安装时不再询问，直接从官方 CDN 安装，提示可用
`arkcli exp uninstall` 卸载后立即启动。交互终端显示安装动画，CI、管道和重定向
环境降级为静态进度提示。调用可能触发本地安装，AI Agent 只有在用户明确要求打开
方舟体验时才能执行；卸载必须有用户明确意图。其他产品构建不提供该命令。

## 解析与组合规则

- Profile：`--profile` > `ARK_PROFILE` > `default_profile` > 第一个 platform profile > `"default"`。
- API Key：`--api-key` > `ARK_API_KEY` > profile > 同 identity store > 产品 `.env` 兼容值。
- Base URL：`--base-url` > `ARK_BASE_URL` > profile 自定义值 > 按 profile 的 Region/Type 派生 > 产品默认值。
- `--profile` 选择持久上下文；它不覆盖显式的 `--api-key` / `--base-url`，也不会被临时调用写回。
- API Key 与 Base URL **不是无条件双向成对**：显式 Base URL 必须有显式 API Key；但显式 API Key 可以和 Endpoint 组成安全调用，CLI 会读取 Endpoint 的权威 region 后派生 platform Base URL。
- API Key + Base URL + Endpoint 构成 stateless 数据面模式，profile 对连接的影响为 `none`。
- 只有 API Key + 模型名时，CLI 仅在该 Key 唯一匹配本地 profile 时推导 plan lane；无法唯一匹配就要求补 Base URL/Endpoint 或显式选择 profile，绝不根据 `ark-*` 文本猜 Key 类型。
- Endpoint 只有在当前/兼容 profile 能提供 paygo Key 时才可直接调用；显式 `--profile` 不兼容时是硬错误，不静默借别的 profile。
- 控制面或本地命令不消费这两个数据面 flag：显式传入会 fail-fast；环境变量在这些命令中被忽略。
- `ARK_REGION` / `ARK_PROJECT_NAME` 不再是运行时覆盖入口。修改持久上下文应创建/切换 Profile；火山老状态中的产品 `.env` Project 只作旧 Profile 缺字段时的兼容兜底。

完整的五类 Profile × 三模态矩阵、组合决策和 Client Preview 边界见
[`execution-context.md`](execution-context.md)。
