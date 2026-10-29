# 原生 WebSocket ASR

火山 Agent Plan **个人版/企业版**支持 Seed ASR 2.0 原生转写和字幕，使用当前
`agent-plan` 专属 Key 或 `agent-plan-team` 席位 Key。先读 `../SKILL.md`
及 `../../arkcli-shared/SKILL.md`；调用归因 `ARKCLI_SKILL_NAME=arkcli-understand`。
用户已授权的素材和真实调用直接沿用；不创建 Endpoint，不换普通 Key。

## 执行

1. 只预览：`+understand asr --input @speech.wav --audio-protocol ws-bidirectional --dry-run`。
   不读音频、不加载 profile/认证、不请求网络、不写文件；真实音频分包和资格标为 unresolved。
2. 真实处理：核对 `auth status`、已有套餐 profile 和素材；个人版先执行使用协议检查，
   企业版走席位 Key；鉴权后再读取文件。保存失败不自动重复识别。
3. 使用明确的 `asr`（纯转写）或 `asr-align`（真实时间戳字幕），必须且只能一个 input。
   `asr-speakers`、翻译和会议纪要继续使用 Responses 配方。
4. 输出成功必须收到服务器最终包；中断和错误不写部分文件。`content` 为完整
   转写或 SRT/VTT；`utterances` 保留服务器的句子、词级时间戳及 additions。

```bash
arkcli +understand asr --input @speech.wav \
  --audio-protocol ws-bidirectional --stream

arkcli +understand asr-align --input @speech.mp3 \
  --audio-protocol ws-unidirectional --subtitle-format vtt --save-to ./subtitles

arkcli +understand asr --input @speech.pcm --model doubao-seed-asr-2.0 \
  --asr-options '{"enable_itn":true,"enable_punc":true}' --dry-run
```

| Flag | 契约 |
| --- | --- |
| `--audio-protocol` | `responses` 默认；原生选择 `ws-bidirectional` / `ws-unidirectional` |
| `--model` | 原生仅 doubao-seed-asr-2.0 或 volc.seedasr.sauc.duration；不指定协议时默认双向 WS |
| `--input` | 单个本地 @file、file:// 或 HTTP(S) URL；客户端最大 64 MiB，不自动转码 |
| `--audio-format` | wav、pcm、mp3、ogg（Opus）；省略按扩展名推断 |
| `--audio-channels` | PCM 声道 1（默认）或 2；WAV 以 header 为准 |
| `--asr-language` | 仅单向 WS 按官方支持的语言值传递 |
| `--asr-options` | request JSON；可配置 ITN、标点、顺滑、热词等官方支持项；model_name 固定 bigmodel，result_type 固定 full；字幕必须 show_utterances=true |
| `--subtitle-format` | 仅 asr-align；srt（默认）/vtt；句级时间戳从服务端毫秒转换 |
| `--save-to` | 可选目录；成功后创建唯一 asr-*.txt/srt/vtt，不覆盖已有文件 |
| `--stream` | NDJSON `asr.delta`（完整的当前文本，可能修订；不是追加文本）与 `asr.completed`（最终结果） |
| `--timeout` | 含读取、分包和识别的截止时间，默认 120s |

PCM/WAV 要求 PCM s16le、16000 Hz、1/2 声道；WAV 同时验证格式与数据 chunk。
MP3、Ogg Opus 由服务端解码。不支持 m4a。原生 ASR 不接受配方 prompt、
system-prompt、thinking、temperature、store 等 Responses 参数，也不接受
临时 `--api-key` / `--base-url`。纯转写可没有话语内容；字幕没有真实时间戳时
会报错，不伪造打轴。`audio_info` 和 `usage` 只保留当前服务器返回的字段，
不承诺历史语音用量查询。错误保留服务端 Code 与 LogID，不自动重试请求。

接口分别为 `/api/v3/plan/sauc/bigmodel_async` 与
`/api/v3/plan/sauc/bigmodel_nostream`，Resource ID `volc.seedasr.sauc.duration`。
[官方 Agent Plan 语音文档](https://ark.volcengine.com/region:cn-beijing/docs/agent-plan-personal-voice-model)。
