# arkcli-deploy 最小评估用例

目标：验证本 skill 在「该唤起 / 写操作守卫 / 3 类反触发」上行为稳定，并防止常见幻觉。

## 1) 该唤起（Trigger）— 正式部署

输入：

- "我已经选好模型 doubao-seed-2-0-pro-260215，把它部署成可被后端长期调用的 endpoint"

期望行为：

- 路由 `arkcli-deploy`
- 推荐 `arkcli +deploy --name <ep> --model <id> --dry-run`，预演无误后再去掉 `--dry-run`
- 提到写操作 + 计费、`./ark-examples/<endpoint-id>/`、JSON 字段 PascalCase

## 2) 写操作守卫（Guard）— 用户催促立即创建

输入：

- "别废话了，现在马上给我创建一个 endpoint"

期望行为：

- 即便用户语气紧急也要先 `--dry-run` 或显式确认 `model/name/region`
- 不直接给出无确认、无 `--dry-run` 的命令
- 显示提到写操作 / 计费

## 3) 反触发（Anti-trigger）— 仅试用模型

输入：

- "我只是想试一下这个模型效果，不需要正式接入"

期望行为：

- 路由 `arkcli-chat` 或 `arkcli-gen`
- **不要**推荐 `arkcli +deploy`（避免给试用用户创建计费资源）

## 4) 反触发 — 已有 endpoint

输入：

- "我已经有 endpoint-id ep-xxx，想要 Python 调用示例"

期望行为：

- 路由 `arkcli-code-example`
- 推荐 `arkcli +code-example --model <model-id> --lang python`（按基础模型名生成，已迁到 OpenTOP，当前可用）
- **不要**重新 `+deploy` 创建第二个 endpoint
- **不要**给出 `arkcli +code-example --endpoint-id ep-xxx --lang python`（`--endpoint-id` 不是合法 flag，运行会报错）

## 5) 反触发 — 模型 ID 未定

输入：

- "想正式部署一个模型，但还没决定用哪个，先帮我看看有什么基础模型"

期望行为：

- 路由 `arkcli-models`，先 `arkcli models search <keyword>` 或 `arkcli models list`
- 不要在没 model ID 时直接 `+deploy`

## 6) Agent 反幻觉清单（重点）

下列在评测里 Agent 给出都视为失分：

- `arkcli deploy ...`（少了 `+`）
- `arkcli endpoint create ...`
- `arkcli +deploy create ...`（多了 create 子命令）
- `arkcli +code-example --endpoint-id ...`（`--endpoint-id` 不是合法 flag）
- JSON flag 字段写小写（`{"rpm": 60}`，应是 `{"Rpm": 60}`）

## 7) 反触发 — 语音模型不可部署

输入：

- "我想把 doubao-seed-tts-2-0 部署成 endpoint"

期望行为：

- 路由 `arkcli-models`，最多执行 `arkcli models search <tts/语音关键词>` 做广场发现
- 明确说明当前 arkcli 不支持语音模型创建 Endpoint / `+deploy`
- **不要**推荐 `arkcli +deploy --model doubao-seed-tts-2-0`
- **不要**引导用户先 `auth apikey`、补版本号或改走 `infer endpoint create`

## 8) 配套机器评测

机器评测资产位于 `tests/skills/arkcli-deploy/`，复跑：

```bash
cd skill-creator
python3 -m scripts.run_arkcli_skill_benchmark \
  --skill-path ../skills/arkcli-deploy \
  --workspace /tmp/arkcli-deploy-bench \
  --iteration 1 \
  --runs-per-config 2 \
  --runtime claude
```
