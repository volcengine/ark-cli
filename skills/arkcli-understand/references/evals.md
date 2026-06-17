# arkcli-understand 最小评估用例

目标：验证本 skill 在「该唤起 / 自动路由 / 反触发 / 写操作与隐私守卫」上行为稳定，并防止常见幻觉（不存在的 flag、命令拼写、错误的 JSON 路径）。

## 1) 该唤起（Trigger）— 明确产出形态的理解任务

输入：

- "帮我把这段录音 `@interview.mp3` 转成文字"
- "把这张图 `@scene.jpg` 里所有的人都框出来，给坐标"

期望行为：

- 路由 `arkcli-understand`
- 录音转写 → `arkcli +understand asr --input @interview.mp3`
- 框选定位 → `arkcli +understand image-grounding --input @scene.jpg "框出所有人"`，并说明输出是 bbox `(x1,y1,x2,y2)`+confidence、需 `jq -r .content | jq .` 二次解析
- **不要**用 `+chat` 去做这两个有固定产出形态的任务

## 2) 自动路由（Trigger）— 省略 sub-skill

输入：

- "理解一下 `@speech.mp3` 这段音频在讲什么"（未点名 sub-skill）

期望行为：

- `arkcli +understand "<prompt>" --input @speech.mp3`，说明会按音频模态自动路由到 `asr`
- 若用户真正想要的是「翻译 / 字幕 / 多说话人」，提示这些**不会**被自动路由命中，要显式写 `ast` / `asr-align` / `asr-speakers`

## 3) 反触发（Anti-trigger）— 开放式带图对话

输入：

- "看着这张图 `@cat.jpg`，陪我聊聊它，我再追问几轮"

期望行为：

- 路由 [`arkcli-chat`](../../arkcli-chat/SKILL.md)：开放对话 + 多轮接续（`--store` / `--previous-response-id`）是 `+chat` 的能力
- **不要**硬塞 `+understand`（它每次是单轮专项理解，无 `--previous-response-id`）

## 4) 反触发（Anti-trigger）— 生成而非理解

输入：

- "帮我生成一张赛博朋克风格的猫的图片"

期望行为：

- 路由 [`arkcli-gen`](../../arkcli-gen/SKILL.md)
- **不要**用 `+understand`（它是理解已有素材，不生成）

## 5) 守卫（Guard）— 大音频 / 隐私上传

输入：

- "转写这个 `@2hours-meeting.wav`（约 300MB）"

期望行为：

- 先提示 audio 走 base64 内联**上限 25MB**、超出不被拦截但很可能失败；建议切片 / 降码率 / 转码后再跑，**不要**假装大文件能直接跑通
- 提示 `--input` 会把文件完整上传 / 内联，敏感素材需用户确认

## 6) 守卫（Guard）— 数据面鉴权失败不重试

输入（命令返回）：

- `Error code: 403 - {"code":"AccessDenied",...}` 或 `ark runtime: API Key is required`

期望行为：

- **不要原地重试**；按 [`../../arkcli-auth/references/auth-modes.md`](../../arkcli-auth/references/auth-modes.md)「API Key 模式的错误恢复」引导用户跑 `arkcli auth apikey`，选好 key 后回到原任务

## 7) Agent 反幻觉清单（重点）

下列在评测里 Agent 给出都视为失分：

- `arkcli understand ...`（少了 `+`）
- `arkcli +understand ...` 不带 `--input`（除非纯演示报错）
- 给 `+understand` 用 `--instructions`（它没有这个 flag；应是 `--system-prompt-override` / `--system-prompt-append`）
- 给 `+understand` 用 `--tools` / `--tools-file` / `--text-format` / `--previous-response-id` / `--include-events`（这些是 `+chat` 的 flag，`+understand` 没有）
- 把内置默认模型 `doubao-seed-1-6` / `doubao-seed-2-0-lite` 当作「裸族名会 404」去强行补版本号（默认值已验证可用，**只有用户显式覆盖 `--model` 时**才需要完整版本化 ID）
- 用 `jq '.output[].content[].text'` 解析（输出已扁平化，应 `jq -r .content`；JSON 类产出再 `| jq .`）
- 编造不存在的 sub-skill 名（合法只有 12 个：`image-caption` / `image-grounding` / `image-gui` / `doc-extract` / `video-summary` / `video-qa` / `vau` / `asr` / `asr-align` / `asr-speakers` / `ast` / `meeting-minutes`）
- 把 `vau` 当成视频自动路由默认（视频默认是 `video-qa`，`vau` 必须显式指定）

## 8) Happy-path（端到端，需前置条件）

前置：已 `arkcli auth status` 登录且配好 ARK API Key。

```bash
# 图片描述（最易复现）
arkcli +understand image-caption --input @<本地图片> "用一句话描述" | jq -r .content
```

检查点：退出码 0；`.content` 非空且是对图片的描述。其余模态各跑 1 条对应 happy-path（音频 ≤25MB）。

## 9) 配套机器评测

机器评测资产应放在 `tests/skills/arkcli-understand/`（不要塞进 `skills/`）。复跑：

```bash
cd skill-creator
python3 -m scripts.run_arkcli_skill_benchmark \
  --skill-path ../skills/arkcli-understand \
  --workspace /tmp/arkcli-understand-bench \
  --iteration 1 \
  --runs-per-config 2 \
  --runtime claude
```
