# arkcli-api-explorer 最小评估用例

目标：验证本 skill 在“该唤起/不该唤起/排错”三个维度上行为稳定。

## 1) 该唤起（Trigger）

输入（用户说法）：

- “产品命令没有覆盖，我要验证某个 Action 的入参/出参契约”
- “我怀疑 registry 里没有这个 operation，帮我确认一下”
- “`arkcli api ...` 报 `unknown action`，怎么定位？”

期望行为：

- 先引导 `arkcli api --list`（或 `arkcli api`）枚举
- 再引导到 `internal/apis/<domain>/` 作为契约事实源
- 最后才执行 `arkcli api <action> --params '{...}'`

## 2) 不该唤起（Anti-trigger）

输入（用户说法）：

- “帮我找一个模型/列出模型/看模型详情”
- "我要试用对话/生成图片/生成视频"
- “我想部署一个 endpoint”
- “我想看用量统计”

期望行为：

- 明确拒绝把 `api` 当默认入口
- 将用户路由到对应产品命令与 skill（`models/+chat/+gen/+deploy/usage`）

## 3) 认证/配置分流（Guard）

输入（用户说法）：

- “我调用报错了/401/鉴权失败/没有权限”
- “profile/base-url/region 看起来不对”

期望行为：

- 先 `arkcli auth status`，必要时转 `arkcli-auth`
- 配置覆盖问题转 `arkcli-config`
- 不要直接用 `api` 反复重试

## 4) CLI 实测命令（可重复）

```bash
# list 模式可用
arkcli api --list --transform '0.name'

# invoke 模式可用（需要已登录且联网）
arkcli api model.list_foundation_models --params '{"PageSize":1,"PageNumber":1}' --transform 'Result.Items.#.Name'
```
