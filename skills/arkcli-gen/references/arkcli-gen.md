# +gen

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **三方渠道订单首用签署：** 抖店/移动商城等三方渠道购买的个人版 plan，首次使用前（含本命令）会被数据协议签署闸门拦截——TTY 下走一屏式交互签署（数字键打开协议 · Space 勾选 · Enter 继续 · Esc 退出），非 TTY 硬拒 `plan_agreement_required`（`--yes` 不放行，逃生 env `ARKCLI_ALLOW_HEADLESS_PLAN_AGREEMENT=1`）。详见 [`../../arkcli-plans/references/arkcli-plans-agreement-signing.md`](../../arkcli-plans/references/arkcli-plans-agreement-signing.md)。

图片/视频生成的执行层文档（`+gen` 全参数）。**这是三步工作流的第 3 步**；完整工作流（① `resources list` 查 profile 资源或 `resources resolve` 解析显式 Endpoint → ② 模型名用 `models get` 查 supported_params → ③ `+gen` 生成）见 [`../SKILL.md`](../SKILL.md)。

> **⚠️ 视频任务默认异步**：提交即返回 `task_id`（`status: queued`），用 `arkcli gen get <task_id>` 轮询；要同步阻塞加 `--wait`。图片任务同步返回。

## 命令

`--model` 使用本次资源解析确认的调用 ID。套餐别名、版本化 ID 和 Endpoint 不是可任意互换的名称。
Agent Plan / Team 的默认视觉别名保持原样，例如 `doubao-seedream-5.0-lite`；
`models get` 返回的规范 Name/Version 是能力查询身份，不能自动替换套餐调用 ID。
Platform / Coding Plan / Coding Plan Team 生成使用已部署的 `ep-*` 和后付费凭证。
只有用户给的是尚未解析的模型族名时，才查询权威模型详情与当前资源候选，不用正则猜版本，
也不强制给所有模型名追加 `primary_version`。示例展示 flag 语法，不是可复制到任何 Profile/模型的固定组合；
执行时用本轮选定的 `$MODEL`，并按 [`intent-and-validation.md`](intent-and-validation.md) 检查能力与成品。

```bash
# 1) 文生图 (T2I) — seedream 模型
arkcli +gen --model doubao-seedream-5-0-260128 "简约商务笔记本电脑，深蓝色，白色背景"

# 2) 图生图 / image-edit (I2I) — 用 --input 艾特参考图
arkcli +gen --model doubao-seedream-5-0-260128 \
  --input @ref.jpg \
  "保留构图，把背景换成黄昏的海滩"

# 多张参考图（多张同时传 --input）
arkcli +gen --model doubao-seedream-5-0-260128 \
  --input @style1.jpg --input @style2.jpg \
  "融合两张图的风格，主体是一只柴犬"

# 一次出多张候选（image-count > 1 自动启用 sequential）
arkcli +gen --model doubao-seedream-5-0-260128 --image-count 4 "未来城市天际线 4 个风格变体"

# 3) 文生视频 (T2V)
arkcli +gen --model doubao-seedance-2-0-260128 "一只柴犬在樱花树下奔跑，慢镜头"

# 4) 图生视频 / 首帧到视频 (I2V) — 第 1 张图默认作为首帧
arkcli +gen --model doubao-seedance-2-0-260128 \
  --input @first.jpg \
  "镜头缓慢拉远，主体保持不动"

# 显式指定 first / last 帧
arkcli +gen --model doubao-seedance-2-0-260128 \
  --input first:@start.jpg --input last:@end.jpg \
  "从这两帧之间生成补间动画"

# 5) 参考视频 (R2V) — 把视频当作参考素材
arkcli +gen --model doubao-seedance-2-0-r2v-260128 \
  --input ref:@reference.mp4 \
  "保持参考视频的运动轨迹，替换主角为机器人"

# 5b) 视频续写 — reference_video 当前必须是服务端可访问 URL
# ratio 按精确模型/EP 能力与用户要求选择；不要把所有续写固定成 adaptive
arkcli +gen --model "$MODEL" --modality video \
  --input reference_video:https://example.com/selected.mp4 \
  --input reference_image:@character.jpg \
  --ratio 16:9 --format json --no-open \
  "从参考视频结尾后的下一瞬开始继续，不重演已有内容；保持人物身份、服装、场景、光线、镜头轴线和运动方向连续"

# 6) 参考音频 — 让视频节奏与音频同步
arkcli +gen --model doubao-seedance-2-0-260128 \
  --input ref:@beat.mp3 \
  "随节拍切换的城市夜景蒙太奇"

# 进阶 flag — 图片任务
arkcli +gen --model doubao-seedream-5-0-260128 \
  --guidance-scale 7.5 --optimize-prompt --output-format png \
  "产品摄影风格"

# 进阶 flag — 视频任务
# 仅在目录明确允许这些键和值时使用，不能照抄到不支持的模型
arkcli +gen --model "$MODEL" --modality video \
  --duration 5 --ratio 16:9 \
  "城市夜景，固定视角"

# 输出完整 JSON
arkcli +gen --model doubao-seedance-2-0-260128 --format json "产品广告视频"

# 自定义推理接入点：先解析权威模态；唯一 image/video 时可自动识别
arkcli resources resolve ep-20260416234150-zsd4v --format json
arkcli +gen --model ep-20260416234150-zsd4v "简约商务笔记本"

# 只在元数据缺失/冲突时显式指定模态
arkcli +gen --model ep-20260416234150-zsd4v --modality video "一只柴犬奔跑"

# Endpoint + 临时 API Key：未给 Base URL 时从 Endpoint 权威 region 派生
arkcli +gen --model ep-... --api-key '<temporary-key>' "一只柴犬奔跑"
```

## 视频续写：独立于普通参考视频生成

当用户说“续写 / 接着生成 / 从结尾继续 / 延长视频 / 第二段 / continue / extend”时，按续写处理；
“参考这个视频的风格、节奏或运镜重新生成”仍是普通 R2V，不能强行套续写规则。

续写提交必须同时满足：

1. 先确认源视频存在、可解码并读取时长/画幅；这只是结构检查，不等于 Agent 看过内容。
2. 当前 CLI 会把本地视频编码成 data URL，但数据面的 `reference_video` 要求 web URL。因此源视频应使用
   `--input reference_video:https://...`；不要提交 `reference_video:@<local.mp4>` 后用真实创建反复探错。
   先前任务返回的 `output_url` 仅在能证明它对应同一源视频且仍有效时复用；没有 URL/上传路径就如实停止。
3. 可选人物/服装/风格图使用 `--input reference_image:@<path>`；不要把它标成 `first_frame`，也不要因为示例
   包含人物图就从历史目录自动补图。只使用用户本轮给出或明确授权复用的素材。
4. `ratio` 不存在续写通用固定值。逐值读取精确模型/Endpoint 的 `supported_params`：用户要求 `16:9` 且
   该值支持就保留；用户要求继承源画幅且 `adaptive` 支持才使用 `adaptive`；未要求时优先省略。
5. `resolution`、`ratio`、`duration` 独立校验。先看是否有已授权资源原生支持目标总时长；只有单任务上限
   不足才提出分段，并在提交前说明额外任务数、计费和连续性风险。
6. prompt 从“结尾后的下一瞬”描述新动作，并明确不重演第一段；保持人物、服装、场景、光线、镜头轴线、
   景别和运动方向。`--return-last-frame` 只要求返回本次生成的最后一帧，不会把普通 R2V 变成续写。

### 多候选提交与对账

- 提交前先列出完整任务图。例如 6 个候选、每个 2 段意味着最多 12 次创建，且第二段依赖第一段 URL；
  不能声称可以一次拿到全部 12 个 ID。按 [`intent-and-validation.md`](intent-and-validation.md) 做套餐/
  免费额度预检；明确耗尽时不启动第一批。
- N 个候选就是 N 次独立创建。默认不加 `--wait`，每次用 `--format json --no-open`，收到一个结构化
  `task_id` 后立即持久记录，再提交下一个；不要并发重复编码/上传同一个大文件。
- `--name` 只可作本地可读标签，不能当服务端幂等键；当前 `gen list` 不回显足以按 name/prompt/content
  唯一认领任务的证据。命令超时或输出中断且没有 `task_id` 时，将该项记为 `UNKNOWN` 并停止，不能靠
  “相同 model 的最近任务”自动补记，也不能盲目重提。
- 所有候选提交完成后再统一轮询已记录的 ID。一个候选失败不自动复制提交；保留每个 ID、状态和文件路径。

“任务成功”只证明服务端完成。续写成品还要比较源视频最后约 1 秒与新视频最初约 1 秒，检查是否重演、
人物/服装/场景突变、镜头跳轴和运动方向反转；没有实际时序证据时只能报告“续写语义未验”，不能声称无缝。

## 临时执行上下文

用户给出 Profile / API Key / Base URL / Endpoint 的任意组合时，先读
[`../../arkcli-shared/references/execution-context.md`](../../arkcli-shared/references/execution-context.md)。

- 显式 Endpoint 优先于 active profile 的套餐模型/default，不要因当前是 Agent Plan
  或 Coding Plan 就改写用户给出的 `ep-...`。
- 显式 Base URL 必须同时显式提供 API Key。
- Endpoint + API Key 可以不传 Base URL：CLI 读取 Endpoint 权威 region 后派生
  platform Base URL。
- API Key + Base URL + Endpoint 是 stateless 单次调用，不切换/修改 profile。
- 不根据 API Key 的 `ark-*` 文本形态判断套餐或凭证类型。
- `+gen --dry-run` 只在客户端构造 `preview.v1`，不联网、不读取素材、不下载或
  打开产物。未知模型/Endpoint 必须显式补 `--modality image|video`；在线执行
  才能解析的 profile、路由和素材内容会列在 `unresolved` 中。

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<prompt>` | 是 | positional | 生成提示词（位置参数，放在命令最后） |
| `--model` | 否：可 fallback 到 active profile 的 `Resources.<modality>.Default` | string | 本次 Profile/凭证兼容的调用 ID：套餐合法别名、当前数据面允许的模型 ID 或 `ep-*`。能力查询规范名不能替换收费路径中的调用 ID。默认为空时先 `resources list`；一次生成不自动设置 default。 |
| `--modality` | 见说明 | string | 生成模态：`image` 或 `video`。真实调用按 `显式值 > ArkModels output_modalities > FoundationModel task types > unknown` 解析；Client Preview 不联网，按 `显式值 > 已知 seedream/seedance 名称 > unknown` 本地判断，未知模型或 Endpoint 必须显式指定 |
| `--input` | 否 | string（可重复） | 参考素材引用，按出现顺序进入 content[]。本地文件用 `@<path>`，远程使用 `https://...` / `tos://...`。可选 role 前缀 — 简写：`first:` `last:` `ref:` `none:`；SDK 显式：`first_frame:` `last_frame:` `reference_image:` `reference_video:` `reference_audio:`。**重要**：本地图片可被内联；本地视频虽会被编码为 data URL，但当前数据面的 `reference_video` 只接受 web URL，续写时必须提供服务端可访问 URL。**图片任务**折叠为 image union；**视频任务**第 1 张图默认首帧，其它图为参考图，视频→ref_video，音频→ref_audio |
| `--name` | 否 | string | 可读任务名覆盖；不能作为服务端幂等键，也不能依赖 `gen list` 按名称恢复 task ID |
| `--version` | 否 | string | 模型版本覆盖 |
| `--size` | 否 | string | 图片输出尺寸，如 `1920x1920`；像素数过小时会被后端拒绝 |
| `--image-count` | 否 | int | 图片任务输出张数；`>1` 时自动转为 `sequential_image_generation=auto + max_images=N` |
| `--n` | 否 | int | `--image-count` 的别名（图片任务）；两者同传时 `--n` 优先 |
| `--ratio` | 否 | string | 输出宽高比覆盖，如 `16:9` / `9:16` / `1:1` / `adaptive`（视频任务）。包括续写在内都按精确模型/EP 的值级能力与用户要求选择；`adaptive` 不是通用强制值 |
| `--resolution` | 否 | string | 输出分辨率，如 `480p`、`720p`、`1080p`（视频任务） |
| `--duration` | 否 | int | 视频时长（秒） |
| `--frames` | 否 | int | 视频帧数（在支持的模型上覆盖 duration） |
| `--seed` | 否 | int | 随机种子，相同 seed 可复现结果 |
| `--watermark` | 否 | bool | 是否添加水印 |
| `--generate-audio` | 否 | bool | 是否同步生成音频（视频任务） |
| `--guidance-scale` | 否 | float | 图片任务的 classifier-free guidance scale，如 `7.5` |
| `--optimize-prompt` | 否 | bool | 图片任务：启用服务端 prompt 优化 |
| `--output-format` | 否 | string | 图片输出格式：`jpeg` 或 `png` |
| `--response-format` | 否 | string | 图片响应格式：`url`（默认）或 `b64_json`（直接返回 base64 编码图片，不存 URL） |
| `--prompt-thinking` | 否 | string | 图片任务：prompt 优化思考模式 `auto` / `enabled` / `disabled`（仅在 `--optimize-prompt=true` 时生效） |
| `--prompt-mode` | 否 | string | 图片任务：prompt 优化执行模式 `standard` / `fast` |
| `--sequential` | 否 | string | 图片任务：序列图模式 `auto` / `disabled`。留空时延续 `--image-count > 1` → `auto` 的自动行为；显式传 `disabled` 强制单图，即使 `--image-count > 1` |
| `--stream` | 否 | bool | 图片任务：流式 NDJSON 输出，详见 [`image-stream.md`](image-stream.md) |
| `--camera-fixed` | 否 | bool | 视频任务：固定虚拟镜头 |
| `--return-last-frame` | 否 | bool | 视频任务：返回最后一帧 URL（用于续接生成） |
| `--draft` | 否 | bool | 视频任务草稿模式：更快 / 更便宜 / 质量更低 |
| `--priority` | 否 | int | 视频任务调度优先级 0-9，越高越优先。**受 supported_params 约束**——Step 2 先查模型是否支持及范围（实测 seedance-2.0 / 2.0-fast 支持 `[0,9]`）；模型不支持时传了会被校验拒 |
| `--service-tier` | 否 | string | 视频任务：服务等级 |
| `--safety-id` | 否 | string | 视频任务：调用方传入的安全标识 |
| `--execution-expires-after` | 否 | int | 视频任务服务端 TTL（秒） |
| `--callback-url` | 否 | string | 视频任务：服务端在生命周期事件（created/running/succeeded/failed）上 POST 通知到此 URL |
| `--wait` | 否 | bool | **视频任务**：阻塞到任务完成再返回。**默认 false**——提交即返回 `task_id` 异步轮询（2.0 起默认行为，旧版默认同步等待） |
| `--extra-body` | 否 | string（JSON 对象） | 视频任务 forward-compat 通道：传 JSON 对象字符串，里面的 key 会 merge 到 top-level 请求 body，让你不升级 arkcli 也能透传服务端新增字段。例：`--extra-body '{"new_field":"value"}'` |
| `--tools` | 否 | string（可重复） | 工具开关，目前支持 `web_search` |
| `--save-to` | 否 | string | 保存生成产物的本地目录，默认 `.`（当前目录）；传 `--save-to=""` 显式关闭自动下载。下载失败不阻塞主流程 |

`--profile` / `--api-key` / `--base-url` 见共享
[`global-flags.md`](../../arkcli-shared/references/global-flags.md)。

## 返回值

**默认输出**（精简）：

```json
{
  "kind": "image",
  "model": "doubao-seedream-5-0-260128",
  "status": "succeeded",
  "output_url": "https://...",
  "output_urls": ["https://..."],
  "local_path": "/work/ark-gen.jpeg",
  "local_paths": ["/work/ark-gen.jpeg"]
}
```

- `output_url` / `output_urls`：TOS 预签名 URL，**24 小时后失效**，不要长期引用
- `local_path` / `local_paths`：由 `--save-to` 触发自动下载后落盘的本地绝对路径；是**持久产物**，优先用它而不是 URL。关闭自动下载（`--save-to=""`）时这两个字段不存在
- `task_id`：视频任务才有。**视频默认异步**——提交即返回此 id + `status: queued`；用 `arkcli gen get <task_id>`（或 `arkcli api arkruntime.get_content_generation_task --params '{"id":"<task_id>"}'`）轮询到 `succeeded` 再取 `output_url`。要同步阻塞用 `+gen --wait`

**`--format json`**：输出结构化结果；下载成功时包含 `local_path` / `local_paths`。
下载关闭或失败时字段可能缺失，不能仅凭生成状态成功声称文件已保存。

视频任务的完整对象除了 `id / status / output_url / ratio / resolution / duration / frames / generate_audio` 等常规字段外，还会回显服务端 echo 出的额外字段（按需出现，由模型与服务端决定）：

| 字段 | 说明 |
|---|---|
| `usage` | `{prompt_tokens, completion_tokens, total_tokens}` token 用量遥测 |
| `revised_prompt` | 服务端最终下发给模型的 prompt（经优化 / 安全过滤后） |
| `subdivisionlevel` | 任务细分等级（注意 JSON key 是单词式无下划线） |
| `fileformat` | 输出文件格式（同样单词式无下划线） |
| `safety_identifier` | 调用方传入的安全标识符回显 |
| `tools` | 任务实际用到的工具类型列表，例如 `["web_search"]` |

服务端不返回时这些字段不出现（omitempty），不影响 `output_url` / `local_path` 等核心字段的提取。

## 常见错误

| 错误 | 原因 | 处理方式 |
|------|------|---------|
| `model is required` | 未指定 `--model` | 必须指定模型名 |
| `Error code: 404 - InvalidEndpointOrModel.NotFound` | 资源不存在、当前身份不可见或调用 ID 不属于当前 lane；不能仅凭 404 认定缺版本 | 同一 Profile 下用 `resources list/resolve`、必要时 `models get` 核对；保留合法套餐别名，不自动改版本/计费路径 |
| `cannot determine generation modality` | Endpoint/模型元数据无法唯一识别 image/video，且未传 `--modality` | 先看 `resources resolve <ep>` 的 warnings；根据用户意图显式补 `--modality image|video` |
| 缺少 prompt | 未提供位置参数 | prompt 是必填的位置参数 |
| `image generation failed: InvalidParameter` | 图片尺寸像素数过小等参数错误；常见于 `1024x1024` 这类尺寸 | 改用 `1920x1920` 及以上尺寸，必要时加 `--debug` 看底层错误 |
| 视频迟迟没结果 | 视频默认**异步**，返回的是 `task_id` + `status: queued`（**非失败**） | 用 `arkcli gen get <task_id>` 轮询到 `succeeded`；**不要**重提 `+gen`（会建新任务）。要同步等可用 `+gen --wait` |
| `warn: auto-download failed: ...`（stderr） | 下载 `output_url` 失败（网络 / 403 等） | 不阻塞主流程；`output_url` 仍返回，用户可手动 `curl -o file.ext "<output_url>"` 抢救（注意 24h 时效） |

## 注意事项

- prompt 是位置参数，放在命令最后，建议用引号包裹
- 模型名、模型版本名、DisplayName、Endpoint 名称和 `ep-*` ID 是不同标识；路由能力只认结构化元数据，不认品牌前缀
- `--input` 是数据驱动的多模态入口：图扩展名走图通道，视频扩展名走视频通道，音频扩展名走音频通道；其他扩展名在视频任务里**会被静默丢弃**（content-generation 没有 input_file slot），在图片任务里也不会被识别为参考图
- 视频任务的多 `--input` 顺序是有意义的：第一张图默认作为 first frame；如果要表达"这张是参考素材，不是首帧"，用 `ref:` 前缀
- 图片生成建议从 `1920x1920` 起步，`1024x1024` 这类尺寸当前可能直接被后端拒绝
- 图片生成通常几秒完成，视频生成可能需要数十秒到数分钟
- **预签名 URL 24h 失效**：`output_url` 只能在短时间内 HTTP `GET` 下载（注意不是 `HEAD`，签名绑定 method）。需要持久保存就依赖 `local_path` 或及时下载
- 默认文件名：有 `task_id`（视频）时取 task_id，否则用 `ark-gen`；Content-Type 推扩展名（`.jpeg` / `.mp4` 等）；同名自动 `-1`/`-2` 后缀
- 如需底层排查或自己掌控提交 + 轮询节奏：直接走 raw API：`arkcli api arkruntime.create_content_generation_task --params '{"model":"<id>","content":[{"type":"text","text":"<prompt>"},{"type":"image_url","image_url":{"url":"https://..."}}]}'` 拿 `id`，然后用 `arkruntime.get_content_generation_task --params '{"id":"<id>"}'` 轮询

## 参考

- [arkcli-gen](../SKILL.md) -- gen skill 概览
- [arkcli-shared](../../arkcli-shared/SKILL.md) -- 认证和全局参数
