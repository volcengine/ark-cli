# 语音合成（+gen --modality audio）

`arkcli +gen --modality audio <text>` 支持 HTTP、WebSocket 单向和双向流式 TTS，
逐块接收音频并在成功结束后保存文件。默认 `--audio-protocol http`；双向 WS
还支持从 stdin 增量发送文本。支持火山 Agent Plan **个人版和企业版**；
个人版使用 `agent-plan` profile 的专属 API Key，企业版使用 `agent-plan-team`
profile 的席位 API Key。两者都由既有 profile 配置解析。Platform、Coding Plan、
临时 `--api-key` / `--base-url` 和 ASR 不属于该命令支持范围。

先读取 `../SKILL.md` 和 `../../arkcli-shared/SKILL.md`。命令均须使用共享协议
的调用归因前缀，`ARKCLI_SKILL_NAME=arkcli-gen`。

## 执行顺序

1. 用户仅要求预览时，直接运行 `+gen --modality audio ... --dry-run`；不先执行任何在线
   auth/resources/models 查询。Preview 是零网络、零文件写入的 `preview.v1`，
   不验证资格、音色权限、额度或音频效果。`unresolved` 说明真实执行需要的
   profile、访问协议和返回后才确定的文件名。
2. 真实合成前确认文本、音色和费用授权。用户已给出的内容与授权直接沿用。
   通过 `arkcli auth status` 核对当前 profile 为 `agent-plan` 或 `agent-plan-team`；
   需要选择已有套餐时用单次 `--profile`，不要切换默认配置、创建 Endpoint 或普通 Key。
3. 使用 `+gen --modality audio`，agent 默认传 `--no-open`。默认音色为
   `zh_female_vv_uranus_bigtts`，格式 mp3、采样率 24000 Hz、超时 2m。
4. 根据结果 `local_path` 交付已保存音频；`audio_bytes` 为实际音频字节数，
   `log_id` 来自服务端 `X-Tt-Logid`，`usage` 原样保留服务端返回项。
   成功必须有终止码 `20000000` 和非空音频。错误或中断不会保存部分音频。
   保存失败时不要直接重新合成，先排查输出目录。

个人版继续执行使用协议检查；企业版沿用团队席位路径，不产生个人版使用协议
intent。没有企业版席位 Key 时按 Profile/Plans 的既有席位流程处理，不回退到
个人版或普通 API Key。

## 模型与模态

使用 `+gen --modality audio` 指定语音生成。当前音频默认模型固定为
`doubao-seed-tts-2.0`，`--model` 可显式指定该名称或别名 `seed-tts-2.0`；
仅指定这两个模型名时也能本地推断 audio。其他音频模型和 Endpoint ID 会被拒绝。
音频模型不来自 Resources 默认值或模型目录接口；服务端路由使用语音资源
`seed-tts-2.0`，请求体不传 ArkRuntime 的 `model` 字段。

`--output-format` 与 `--timeout` 按模态应用默认值：音频为 mp3 / 2m，
图片和视频沿用各自默认值。音频支持 `--stream`，不接受图片/视频的 `--input`、
`--wait`、`--generate-audio` 等参数；`--generate-audio` 仍是视频配套音轨参数。
`--speaker`、`--sample-rate`、`--speech-rate`、`--loudness-rate`、`--emotion`、
`--additions`、`--audio-protocol`、`--text-stream`、`--bit-rate`、`--emotion-scale`、
`--enable-subtitle`、`--subtitle-format` 仅用于 audio，传给图片或视频会报错。

## 参数与示例

```bash
arkcli +gen --modality audio "你好，欢迎使用语音合成。" \
  --speaker zh_female_vv_uranus_bigtts \
  --save-to ./audio \
  --no-open

arkcli +gen --modality audio "欢迎收听今天的播报。" \
  --speech-rate=-10 \
  --additions '{"context_texts":["用平静温柔的语气朗读"]}' \
  --no-open \
  --dry-run
```

| Flag | 说明 |
| --- | --- |
| `--modality audio` | 指定音频生成 |
| `--model` | 可省略；支持 doubao-seed-tts-2.0 或 seed-tts-2.0 |
| `--audio-protocol` | `http`（默认）、`ws-unidirectional`（完整文本输入）、`ws-bidirectional`（支持增量文本） |
| `--text-stream` | 仅双向 WS；位置参数必须是 `-`，鉴权后读取 stdin；dry-run 不读 stdin |
| `--stream` | 输出 NDJSON `audio.delta`（base64）、`audio.subtitle`（时间戳）、`audio.completed`（最终结果） |
| `--bit-rate` | MP3 比特率，单位 bps；低于 64000 自动启用 additions.disable_default_bit_rate，冲突的 false 被拒绝 |
| `--emotion-scale` | 1..5，可用小数；必须同时指定音色支持的 `--emotion` |
| `--enable-subtitle` | 请求服务端词级时间戳；结果保留 subtitles；指定后没有字幕会报错 |
| `--subtitle-format` | srt（默认）、vtt、json；必须启用字幕；随音频保存同名 sidecar |
| `--speaker` | 服务端音色 ID，支持范围以官方音色目录与权限为准 |
| `--output-format` | `mp3`、`pcm`、`ogg_opus`；ogg_opus 保存为 `.ogg` |
| `--sample-rate` | 8000、16000、22050、24000、32000、44100、48000 Hz |
| `--speech-rate` / `--loudness-rate` | -50 到 100，显式 0 保留在请求中 |
| `--emotion` | 仅对支持该情感的音色使用 |
| `--additions` | JSON 对象，编码为 `req_params.additions` 的 JSON 字符串；不校验动态音色能力 |
| `--save-to` | 保存目录，默认 `.`；使用唯一 `voice-*` 文件名，保留已有文件 |
| `--save-to=` | 不写文件，结果通过 `b64_audio` 返回音频 |
| `--timeout` | 合成截止时间，如 `2m` 或 `5m`；正数 |
| `--open` / `--no-open` | 默认仅交互终端自动打开；agent 使用 `--no-open` |
| `--dry-run` | 叶子局部的零网络 Preview |

该命令不提供 wav：HTTP 分块中可能重复返回 WAV header，不能直接拼接为完整文件。
CLI 当前限制解码后音频为 64 MiB、HTTP 响应为 128 MiB；这是客户端保护边界。

## 计费与错误

个人版语音调用消耗套餐 AFP。未开启超额后付费时不消耗其他资源包或账户余额；
开启后，周期额度用尽时服务端会自动进入超额后付费。该个人版计费说明不用于
推断企业版的具体费用或额度；企业版按所选套餐和席位的计费规则执行。不能因
使用套餐 Key 就称作免费，也不根据 `usage.text_words` 自行估算金额。该响应中的用量不是
`usage stats` 已支持查询语音历史明细的证据。

HTTP、WebSocket 和 TTS 错误保留服务端状态及 LogID。鉴权错误按 Auth/Profile 技能处理；
不要重新创建 Key。音色权限或参数错误先核对原请求。流中失败或 EOF 不能
包装成成功音频，也不自动重试整段合成。

协议与参数参考：[Agent Plan 接入语音模型](https://ark.volcengine.com/region:cn-beijing/docs/agent-plan-personal-voice-model)、
[HTTP Chunked TTS 参数文档](https://docs.volcengine.com/docs/6561/1598757?lang=zh)。

## WS、情绪强度与字幕示例

```bash
arkcli +gen --modality audio "欢迎收听。今天阳光很好。" \
  --audio-protocol ws-unidirectional --stream \
  --bit-rate 16000 --enable-subtitle --subtitle-format srt \
  --save-to ./audio --no-open

arkcli +gen --modality audio "今天真开心！" \
  --audio-protocol ws-bidirectional --emotion happy --emotion-scale 3 \
  --enable-subtitle --subtitle-format vtt --save-to ./audio --no-open

printf '早上好。今天我们出发去公园。' | arkcli +gen - --modality audio \
  --audio-protocol ws-bidirectional --text-stream --stream \
  --save-to ./audio --no-open
```

字幕直接使用服务端词级 `startTime/endTime`（秒），输出 SRT/VTT 时按句子组合，
不会推测或补造时间戳。字幕 sidecar 保留已有文件。`--save-to=` 时音频在最终结果
的 `b64_audio`，字幕在 `subtitles`，不创建 sidecar。`--stream` 的 delta
仅表示中间片段，成功以 `audio.completed` 为准；CLI 仍聚合音频用于最终输出。
ASR 使用 [`arkcli-understand`](../../arkcli-understand/SKILL.md) 的原生 WS 入口。
