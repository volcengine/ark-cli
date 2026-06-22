# Ark 错误码诊断手册

> 这是 [`arkcli-doctor`](../SKILL.md) 的错误码总册。`arkcli doctor error <code>` 返回的 JSON 里 `reference` 字段指向本文件、`subtype` 字段指向具体小节。Scope 命令（[`doctor infer-endpoint`](scope-infer-endpoint.md) / [`doctor model`](scope-model.md)）输出的 `findings[].error_code` 也在这里查。
>
> 错误码分两大类：[**方舟错误码**](#1-方舟错误码) 和 [**公共错误码**](#2-公共错误码)。

## 目录

| 跳到哪 | 看什么 |
|--------|--------|
| [怎么用](#怎么用) | 三条路由规则，先读 |
| [§1 方舟错误码](#1-方舟错误码) | Ark 数据面调用时返回的错误（[官方文档](https://www.volcengine.com/docs/82379/1299023?lang=zh)） |
| [§1.1 生视频特有](#11-生视频特有seedance) | seedance 5 个完整 subtype：[real_face](#111-input_real_face) / [copyright](#112-input_copyright) / [content_safety](#113-input_content_safety) / [video_copyright](#114-output_video_copyright) / [video_safety](#115-output_video_safety) |
| [§1.2 ModelAccessDenied](#12-modelaccessdenied) | 模型未开通或权限不足（聚合 subtype） |
| [§1.3 RateLimitExceeded](#13-ratelimitexceeded) | 限流（聚合 subtype，覆盖多个 429 码） |
| [§2 公共错误码](#2-公共错误码) | 跨 VolcEngine 服务的公共错误码（[官方文档](https://www.volcengine.com/docs/6369/68677?lang=zh)） |

> doctor 的**前置依赖闸门**（VMP / TLS / TOS 跨服务授权）不是 API 错误码，由 `arkcli doctor account` 输出 + [`arkcli-shared`](../../arkcli-shared/SKILL.md) / [`arkcli-auth`](../../arkcli-auth/SKILL.md) 联合处理。本文件只覆盖 API 抛出的错误码。

## 怎么用

**三条路由规则**：

1. **来自 `arkcli doctor error <code>`** —— JSON 里 `subtype` 直接告诉你跳哪节，按完整 5 段（解读 / 诊断 / 修复 / 缺口 / 闭环）回答用户。
2. **来自 scope 命令的 `findings[].error_code`** —— 在所属 section 的索引表里搜这个 code，按"处理方式"列指向跳；表里没指向具体小节的，照"处理方式"直接答用户。
3. **复合错误码（`rules` 字段多于一个候选）** —— 用完整小节里的诊断决策树消歧，**只给最可能那一条**修复，备选作为"如果 A 不行试 B"附后。

**两个不要**：

- 不要复述完整 JSON。摘 `root_cause` + 一条修复给用户。
- 不要假装能修没上线的能力。`needs_backend` 非空（如 `deface`）时如实告诉用户该能力尚未上线，给当前可行替代。

**新错误码贡献**：见 [`../CONTRIBUTING.md`](../CONTRIBUTING.md)。

---

## 1. 方舟错误码

Ark 数据面调用时返回的错误码——内容安全 / 版权 / 隐私拦截、风险识别、参数 / 格式校验、鉴权 / 模型权限、限流、账号 / 计费状态、资源状态、精调操作约束。

| HTTP | Type            | 错误码                                                          | 释义                                                       | 处理方式                                                                                       |
|------|-----------------|----------------------------------------------------------------|------------------------------------------------------------|------------------------------------------------------------------------------------------------|
| 400  | BadRequest      | `MissingParameter`                                             | 缺必参                                                      | 按 message 里的字段名 + API 文档补                                                            |
| 400  | BadRequest      | `MissingParameter.{{Parameter}}`                               | 缺指定的请求参数                                            | 按 message 提示的字段名补                                                                     |
| 400  | BadRequest      | `InvalidParameter`                                             | 参数非法（笼统）                                            | 看 message 详情，逐字段对照 API 文档                                                          |
| 400  | BadRequest      | `InvalidParameter.{{Parameter}}`                               | 指定参数值不合法                                            | 按 message 提示的字段名校正                                                                   |
| 400  | BadRequest      | `InvalidParameter.UnsupportedParameter`                        | 该参数在此推理接入点不可用                                  | 换 endpoint，或去掉该参数                                                                      |
| 400  | BadRequest      | `InvalidParameter.TosURLInvalid`                               | TOS URI 非法                                                | 检查 URL 格式 / 桶权限                                                                         |
| 400  | BadRequest      | `InvalidArgumentError`                                         | messages 里某条缺 role 字段                                 | 补 role: user / assistant / system / tool                                                     |
| 400  | BadRequest      | `InvalidArgumentError.UnknownRole`                             | role 值不被支持，或 inference_role 未在配置中定义            | 用合法 role 值；检查模型配置                                                                  |
| 400  | BadRequest      | `InvalidArgumentError.InvalidImageDetail`                      | `image_url` 中 `detail` 参数无效                            | 用 `auto` / `high` / `low`                                                                    |
| 400  | BadRequest      | `InvalidArgumentError.InvalidPixelLimit`                       | `min_pixels` / `max_pixels` 无效                            | 校验 min ≤ max；不超过服务配置范围                                                            |
| 400  | BadRequest      | `InvalidImageURL.EmptyURL`                                     | 图片 URL 为空                                               | 客户端检查 base64 / URL 拼接                                                                  |
| 400  | BadRequest      | `InvalidImageURL.InvalidFormat`                                | 图片 base64 / 格式不对                                      | 校验 base64 padding；换 PNG/JPEG                                                              |
| 400  | BadRequest      | `OutofContextError`                                            | 文本 + 图编码后总 token 超 context                          | 减少图 / 截断文本 / 换长 context 模型                                                          |
| 400  | BadRequest      | `InvalidEndpoint.ClosedEndpoint`                               | 推理接入点已关闭或暂时不可用                                | 联系管理员或重启：`arkcli infer endpoint start`                                                |
| 400  | BadRequest      | `Duplicate.Tags.Key`                                           | tag 列表中存在重复 key                                      | 去重                                                                                           |
| 400  | BadRequest      | `InputTextSensitiveContentDetected`                            | 输入文本敏感                                                | 改写 prompt（生视频场景见 [§1.1.3](#113-input_content_safety)）                              |
| 400  | BadRequest      | `InputImageSensitiveContentDetected`                           | 输入图敏感                                                  | 换素材（生视频场景见 [§1.1.3](#113-input_content_safety)）                                   |
| 400  | BadRequest      | `InputVideoSensitiveContentDetected`                           | 输入视频敏感                                                | 见 [§1.1.3](#113-input_content_safety)                                                        |
| 400  | BadRequest      | `OutputTextSensitiveContentDetected`                           | 输出文本敏感（LLM）                                          | 改写 prompt / 收敛引导                                                                         |
| 400  | BadRequest      | `OutputImageSensitiveContentDetected`                          | 输出图敏感                                                   | 收敛 prompt / 换风格                                                                           |
| 400  | BadRequest      | `OutputVideoSensitiveContentDetected`                          | 输出视频敏感                                                 | 见 [§1.1.5](#115-output_video_safety)                                                          |
| 400  | BadRequest      | `OutputVideoSensitiveContentDetected.PolicyViolation`          | 输出视频版权（需消歧）                                       | 见 [§1.1.4](#114-output_video_copyright)                                                       |
| 400  | BadRequest      | `Input{Image,Video}SensitiveContentDetected.PrivacyInformation` | 输入素材含真实人脸                                           | 见 [§1.1.1](#111-input_real_face)                                                              |
| 400  | BadRequest      | `Input{Text,Image,Video,Audio}SensitiveContentDetected.PolicyViolation` | 输入素材涉版权                                       | 见 [§1.1.2](#112-input_copyright)                                                              |
| 400  | BadRequest      | `InputTextRiskDetection`                                       | 风险识别产品 - 输入文本拦截                                  | 改写 prompt                                                                                    |
| 400  | BadRequest      | `InputImageRiskDetection`                                      | 风险识别产品 - 输入图拦截                                    | 换素材                                                                                          |
| 400  | BadRequest      | `OutputTextRiskDetection`                                      | 风险识别产品 - 输出文本拦截                                  | 收敛 prompt                                                                                    |
| 400  | BadRequest      | `OutputImageRiskDetection`                                     | 风险识别产品 - 输出图拦截                                    | 收敛 prompt / 换风格                                                                           |
| 400  | BadRequest      | `ContentSecurityDetectionError`                                | 风险识别产品**本身**调用失败                                | 重试；持续失败提工单（带 CSDRequestId）                                                       |
| 400  | BadRequest      | `SensitiveContentDetected`                                     | 输入文本敏感（旧版兼容）                                    | 改写 prompt                                                                                    |
| 400  | BadRequest      | `SensitiveContentDetected.SevereViolation`                     | 输入文本严重违规                                            | 改写 prompt                                                                                    |
| 400  | BadRequest      | `SensitiveContentDetected.Violence`                            | 输入文本含激进行为                                          | 改写 prompt                                                                                    |
| 400  | Forbidden       | `InvalidSubscription`                                          | Coding Plan 套餐未订阅或已过期                              | 续订（走 `arkcli-plans`，本仓未独立成 skill 时回 [`arkcli-shared`](../../arkcli-shared/SKILL.md)） |
| 401  | Unauthorized    | `AuthenticationError`                                          | API Key / AK-SK 校验未通过                                  | 重新登录：`arkcli auth login`（走 [`arkcli-auth`](../../arkcli-auth/SKILL.md)）              |
| 401  | Forbidden       | `InvalidAccountStatus`                                         | 当前账号状态异常                                            | 联系平台管理员；可能涉及合规 / 封禁                                                            |
| 403  | Forbidden       | `AccessDenied`                                                 | 没有访问该资源的权限                                         | 见 [§1.2](#12-modelaccessdenied) 修复 B                                                       |
| 403  | Forbidden       | `OperationDenied.PermissionDenied`                             | 无权访问基础模型的配置                                       | 见 [§1.2](#12-modelaccessdenied) 修复 B                                                       |
| 403  | Forbidden       | `OperationDenied.ServiceNotOpen`                               | 模型服务未激活                                              | 控制台激活，等同 [§1.2](#12-modelaccessdenied) 修复 A                                         |
| 403  | Forbidden       | `OperationDenied.ServiceOverdue`                               | 账单已逾期，操作被拒                                        | 充值：`https://console.volcengine.com/finance/recharge`                                       |
| 403  | Forbidden       | `AccountOverdueError`                                          | 当前账号欠费（余额 < 0）                                    | 同上：火山费用中心充值                                                                        |
| 403  | Forbidden       | `OperationDenied.FileQuotaExceeded`                            | 文件存储额度耗尽                                            | 删历史文件，或控制台升级套餐                                                                  |
| 403  | Forbidden       | `OperationDenied.InvalidState`                                 | 关联资源处于非空闲态（Context InProgress / 缓存更新中 / File 非可用） | 等待任务完成 / 取消 / 换 ID                                                                   |
| 403  | Forbidden       | `OperationDenied.UnsupportedPhase`                             | 资源处于特殊状态（锁定 / 处理中等）                         | 检查资源状态；不可解则提工单                                                                  |
| 403  | Forbidden       | `OperationDenied.ConflictedValidationSet`                      | 同时上传验证集 + 设置训练集采样为验证集百分比               | 二选一（精调）                                                                                 |
| 403  | Forbidden       | `OperationDenied.UnsupportedCustomizationType`                 | 模型不支持该精调方法                                         | 换方法或换模型                                                                                 |
| 403  | Forbidden       | `OperationDenied.CustomizationNotSupported`                    | 基础模型版本不支持该精调方法                                 | 换基础模型版本                                                                                 |
| 404  | NotFound        | `InvalidEndpointOrModel.NotFound`                              | 模型 / endpoint 不存在或无权访问                             | 检查 ID + profile region；可能是 [§1.2](#12-modelaccessdenied)                                |
| 404  | NotFound        | `InvalidEndpointOrModel.ModelIDAccessDisabled`                 | 不允许直接用 model ID 调用，需用 endpoint ID                 | 用 endpoint ID 调（不要直接 model ID）                                                         |
| 404  | NotFound        | `ModelNotOpen`                                                 | 当前账号未开通该模型                                         | 见 [§1.2](#12-modelaccessdenied) 修复 A                                                       |
| 404  | NotFound        | `NotFound.{{Parameter}}`                                       | 指定资源找不到                                              | 按 message 提示的参数名核对                                                                   |
| 404  | NotFound        | `UnsupportedModel`                                             | 模型不支持 Coding Plan                                      | 换支持的模型                                                                                    |
| 429  | TooManyRequests | `RateLimitExceeded.EndpointRPMExceeded`                        | endpoint RPM 撞顶                                            | 见 [§1.3](#13-ratelimitexceeded)（接入点级瓶颈）                                              |
| 429  | TooManyRequests | `RateLimitExceeded.EndpointTPMExceeded`                        | endpoint TPM 撞顶                                            | 见 [§1.3](#13-ratelimitexceeded)（接入点级瓶颈）                                              |
| 429  | TooManyRequests | `ModelAccountRpmRateLimitExceeded`                             | 账号下该模型 RPM 撞顶                                        | 见 [§1.3](#13-ratelimitexceeded)（模型级瓶颈，修复 A 提配额）                                |
| 429  | TooManyRequests | `ModelAccountTpmRateLimitExceeded`                             | 账号下该模型 TPM 撞顶                                        | 见 [§1.3](#13-ratelimitexceeded)（模型级瓶颈）                                                |
| 429  | TooManyRequests | `ModelAccountIpmRateLimitExceeded`                             | 账号下该模型 IPM 撞顶（生图）                                | 见 [§1.3](#13-ratelimitexceeded)（模型级瓶颈）                                                |
| 429  | TooManyRequests | `APIAccountRpmRateLimitExceeded`                               | 账号下某 API 接口 RPM 撞顶                                   | 见 [§1.3](#13-ratelimitexceeded) 修复 D + A                                                   |
| 429  | TooManyRequests | `AccountRateLimitExceeded`                                     | 账号 RPM / TPM 撞顶（笼统）                                  | 见 [§1.3](#13-ratelimitexceeded)                                                              |
| 429  | TooManyRequests | `QuotaExceeded`                                                | 免费额度用尽 / 排队任务数超限 / 5h-周-月限额到顶            | 看 message 区分场景：免费额度 → 控制台开通付费；排队 → 等待 / 提配额                          |
| 429  | TooManyRequests | `SetLimitExceeded`                                             | 用户自己设的推理限额到顶                                    | 控制台改限额或关闭安心体验模式                                                                |
| 429  | TooManyRequests | `InflightBatchsizeExceeded`                                    | 充值金额对应的最大并发数撞顶                                | 充值升档或降并发                                                                              |
| 429  | TooManyRequests | `ServerOverloaded`                                             | 服务资源紧张（流量突增 / 长时间未用的 endpoint 冷启动）     | 退避重试；持续触发提工单                                                                      |
| 429  | TooManyRequests | `RequestBurstTooFast`                                          | Doubao-Seed-2.0+ 请求量激增触发系统保护                     | 放缓提速速度，逐步增加请求量                                                                  |
| 500  | InternalServerError | `InternalServiceError`                                     | Ark 内部异常                                                | 加退避重试（同 [§1.3](#13-ratelimitexceeded) 修复 D）；持续问题提工单带 RequestID            |

### 1.1 生视频特有（seedance）

seedance 视频生成场景下的内容安全 / 版权 / 隐私拦截。**错误码前缀虽然有 Input/Output，但这一段不分输入输出**——根因和修复都聚焦在生视频任务（参考图 / 参考视频 / prompt → 生视频），与其它模态不共享。

> 同名前缀错误码从非视频调用来（如 LLM 调用时拿到 `InputTextSensitiveContentDetected`）→ 看上方主索引表的"处理方式"列直接处理，不要进本节。

#### 1.1.1 `input_real_face`

> 输入参考图 / 视频含真实人脸（隐私拦截）。

**路由触发**：`InputImageSensitiveContentDetected.PrivacyInformation` / `InputVideo…PrivacyInformation`。

##### 解读

参考图 / 参考视频里检测到了真实人脸，触发隐私保护拦截。后缀 `.PrivacyInformation` = 隐私（区别于版权 `.PolicyViolation`）。

##### 诊断步骤

- 直接看用户提供的素材里有没有清晰的真实人脸（多模态可直接判断）
- 确认这张人脸是不是用户**有意保留**的人物身份——决定能不能直接换图

##### 修复方案

```bash
arkcli +gen --model <原模型> --input @<合规图> "<原 prompt>"
```

- 用户只是要"某种长相"而非特定真人 → 用 seedream 生成虚拟人像替换（走 [`arkcli-gen`](../../arkcli-gen/SKILL.md)）
- 用户坚持要还原这张真人脸 → 见能力缺口

##### 能力缺口（deface 待上线）

对真人脸做去身份化（deface）后再喂进去，可以在保留相似长相的同时通过隐私拦截。该能力**尚未上线**（`needs_backend: ["deface"]`）。当前只能换合规 / 虚拟人像，**不要**承诺一键 deface。

##### 闭环验证

重试不再返回 `.PrivacyInformation` 即通过。

---

#### 1.1.2 `input_copyright`

> 输入素材涉版权。

**路由触发**：`Input*SensitiveContentDetected.PolicyViolation`（生视频场景）。

##### 解读

输入素材涉版权（后缀 `.PolicyViolation` = 版权）。具体哪个素材，看错误码前缀 + 用户实际传了什么。

##### 诊断步骤（按素材类型分流）

| 错误码前缀 | 涉版权的素材 | 处理方向 |
|------------|--------------|----------|
| `InputText…` | prompt 文案 | 改写文案，去掉具体作品 / 角色 / 品牌名 |
| `InputImage…` | 参考图（画面人物 / 受版权画作） | 换合规素材；画面人物 → 见能力缺口 |
| `InputAudio…` | 参考音频 | 换无版权音频，或去掉参考音频 |
| `InputVideo…` | 参考视频——**需再消歧** | 见下方 |

**参考视频 1:N 消歧**（`InputVideoSensitiveContentDetected.PolicyViolation`）：

- 视频音轨版权（背景音乐）→ 用剪辑工具去掉音轨后重传
- 视频画面人物版权 → 换素材；或对人物 deface（见缺口）
- 分不清 → 先去音轨重试；仍失败多半是画面

##### 修复方案

```bash
# 去掉涉版权素材后重试（以参考音频为例：直接不传它）
arkcli +gen --model <原模型> --input @<合规素材> "<改写后去掉版权指向的 prompt>"
```

- 视频音轨版权：用户**自行去音轨**后重传（doctor 不替用户剪视频）
- 文案版权：把 prompt 里的具体作品名 / 角色名 / 品牌名换成泛化描述

##### 能力缺口（deface 待上线）

参考视频 / 图的画面人物版权，最稳是对人物 deface。该能力**尚未上线**（`needs_backend: ["deface"]`）。当前只能换素材，**不要**承诺一键处理画面人物。

##### 闭环验证

去掉 / 替换涉版权素材后重试，不再返回 `.PolicyViolation` 即通过。视频音轨类需用户确认已去轨。

---

#### 1.1.3 `input_content_safety`

> 输入触发内容安全（非版权、非隐私）。

**路由触发**：`Input*SensitiveContentDetected`（无后缀，生视频场景）。

##### 解读

输入触发了内容安全策略（无后缀 = 既非版权也非隐私，而是敏感内容本身）。**没有自动修复**——只能调整输入。

##### 诊断步骤

- 按错误码前缀定位输入：`InputText…`=文案，`InputImage…`=参考图，`InputVideo…`=参考视频，`InputAudio…`=参考音频
- 看该输入里有什么敏感内容（暴力 / 色情 / 政治敏感 / 违禁等），多模态可直接判断

##### 修复方案

调整输入中的敏感内容后重试——这是用户动作，不是可自动跑的命令：

- 文案：改写 prompt，去掉敏感表述
- 图 / 视频 / 音频：换一个不含敏感内容的素材

##### 闭环验证

调整后重试，不再返回该错误码即通过。若反复命中同类拦截，提示用户该主题的内容安全风险较高、建议换方向。

---

#### 1.1.4 `output_video_copyright`

> 输出视频版权（需消歧）。

**路由触发**：`OutputVideoSensitiveContentDetected.PolicyViolation`。

> 这是最常见、也最容易误判的一类。错误码本身**不告诉你**到底是人物还是配乐导致——`rules` 给 4 个候选信号（`prompt_named_ip` / `auto_music_copyright` / `model_generated_face` / `face_id_unstable`），结合用户实际生成输入消歧、择一修复。4 个信号都不匹配（参考图人脸正常但仍撞）→ 走 `deface`（兜底，见缺口段）。

##### 解读

输出视频触发版权策略。在 seedance 上常见以下几种，**最常见的是第一种**：

- **prompt 点名了版权 IP / 名人**：写了真人明星、版权角色或品牌 IP（如「钢铁侠」「漫威」「某明星」），输出自然撞版权。最常见、也最好修。
- **自动配乐版权**：你没要音频，但模型自动配的背景音乐撞了版权库。最容易忽略。
- **模型自主人脸版权**：没给参考图、prompt 里又有人物描述，模型自己"长"出来的人脸撞了某个真人 / 明星。
- **参考人物 ID 不稳**：给了参考图，但人脸占比太小 / 多张脸 / 拼接混用，输出人物漂移成了某个版权人物。

##### 诊断步骤（消歧决策树）

按用户实际的 `+gen` 输入从上往下走，命中即停：

```
prompt 里点名了真人明星 / 版权角色 / 品牌 IP 吗（如「钢铁侠」「迪士尼」「某明星」）?
 ├─ 是 ──► 【prompt 含版权 IP】（最常见）→ 修复 D：改写 prompt 去掉专有名词
 └─ 否 ──► 继续看素材信号:
      有参考图吗?
       ├─ 没有 ──► prompt 里有明确人物描述吗（人物 / 女生 / 男人 / 明星 / 演员 / 角色 / 跳舞…）?
       │           ├─ 有 ──► 【模型自主人脸版权】（可能叠加配乐）→ 修复 A
       │           └─ 没有 ─► 用户显式要过音频 / 参考音频吗?
       │                       ├─ 没有 ─► 【自动配乐版权】（最可能）→ 修复 B
       │                       └─ 有 ──► 信息不足，让用户补充或同时试 修复 B
       └─ 有 ────► 参考图人脸：占比过小（<画面 15%） / 多张脸 / 拼接混用?
                    ├─ 是 ──► 【参考人物 ID 不稳】→ 修复 C
                    └─ 否（单张清晰正脸）─► 【参考人物本身涉版权】→ 见能力缺口段
```

> 判定"参考图人脸占比 / 多脸 / 混用"时，可直接看用户素材，不需额外调接口。

##### 修复方案（先给用户看、等确认再执行）

**修复 D —— 改写 prompt 去掉版权 IP / 名人**（prompt 点名时最常见、最该先试）：

把专有名词换成泛化的外观 / 场景描述。例：「钢铁侠在城市上空飞行战斗」→「一位身穿银色未来感机甲的探险者在城市上空飞行战斗」。去掉品牌 / 作品名（漫威 / 迪士尼 / 某游戏名等）。可叠加 `--generate-audio false` 防自动配乐二次触发。

```bash
arkcli +gen --model <原视频模型> --generate-audio false "<去掉 IP/名人、换成泛化外观描述的新 prompt>"
```

**修复 B —— 关闭自动配乐**（最轻、命中率高，无参考图无音频时首选）：

```bash
arkcli +gen --model <原模型> --generate-audio false "<原 prompt>"
```

**修复 A —— 先生成合规人像当参考图，再生视频**（无参考图 + 人物 prompt）：

```bash
# 1) 先用 seedream 生成一张合规虚拟人像
arkcli +gen --modality image --model <seedream 模型> "<把人物外貌描述抽出来，强调虚构 / 非特定真人>"
# 2) 用这张人像当参考图重生成视频，并顺手关掉自动配乐（双保险）
arkcli +gen --model <原视频模型> --input @<上一步的人像> --generate-audio false "<原 prompt>"
```

**修复 C —— 稳定参考人物身份**（有参考图但人脸不稳）：

```bash
# 1) 把参考图里的人脸裁剪出来、放大占比（去掉多余背景 / 其它脸）
# 2) 用裁剪后的图当参考图，并在 prompt 里强化"保持该人物一致 / 单一主体"
arkcli +gen --model <原视频模型> --input @<裁剪后的人脸图> "<原 prompt>，保持人物外貌一致，画面只有一个主体人物"
```

> 三个修复都用 `arkcli +gen` 原生参数，具体生成走 [`arkcli-gen`](../../arkcli-gen/SKILL.md) 三步工作流（先 `resources list` 选模型、`models get` 查可用参数，再 `+gen`）。

##### 能力缺口（deface 待上线）

"参考图人脸清晰正常、但输出人物仍撞版权"（决策树最后一支），最稳是对人物做去身份化（deface）。该能力**尚未上线**（`needs_backend: ["deface"]`）。当前替代：用**修复 A** 生成全新合规虚拟人像替换原参考图。**不要**向用户承诺"一键去身份化"。

##### 闭环验证

修复命令跑完后：

- 视频任务 `succeeded` → 修复成功
- 仍报同一错误码 → 换决策树下一支（如 B 没成就试 A），或提示用户该 prompt / 素材的版权风险较高、建议更换主体描述

---

#### 1.1.5 `output_video_safety`

> 输出视频内容安全（非版权）。

**路由触发**：`OutputVideoSensitiveContentDetected`（无后缀）。

##### 解读

模型输出视频触发内容安全策略（无后缀 = 内容安全，不是版权）。和"输入内容安全"不同的是：这里输入可能看着没问题，是模型按 prompt / 素材**生成**出了敏感画面。

##### 诊断步骤

- 输出敏感几乎都源于输入引导：回看 prompt 和参考素材里有没有可能被模型放大成敏感画面的描述 / 元素
- 若有参考视频 / 图，看它们的风格、动作、场景是否本身偏敏感

##### 修复方案

调整输入 prompt / 素材中会导向敏感画面的部分后重试（用户动作）：

```bash
arkcli +gen --model <原模型> --input @<更安全的素材> "<去掉会被放大成敏感画面的描述后的 prompt>"
```

- 收敛 prompt 里可能引发敏感画面的动作 / 场景 / 着装等描述
- 替换偏敏感的参考素材

##### 闭环验证

调整后重试，输出不再被拦即通过。若同一 prompt 反复输出敏感，提示用户换主题或显著弱化相关描述。

---

### 1.2 `ModelAccessDenied`

> 模型未开通或权限不足（doctor 聚合 subtype，覆盖 `ModelNotOpen` / `AccessDenied` / `OperationDenied.PermissionDenied`）。

**路由触发**：用户拿到 `ModelAccessDenied`，或 `arkcli doctor infer-endpoint <ep-id>` 主导错误码是上述任一。

#### 解读

调用方拿不到这个模型——HTTP 403 / 404，常见三种：

- **模型未在你账号下开通**（`ModelNotOpen`）：付费模型（如生视频 seedance 系列）需要在控制台显式开通；首次使用必踩一次。最常见。
- **子账号缺 ARKFullAccess**（`AccessDenied` / `OperationDenied.PermissionDenied`）：账号已开通该模型，但当前 AK/SK 对应的子账号没挂 Ark 系统策略。
- **模型版本与 endpoint 配置不一致**：endpoint 绑了一个旧版本，新流量调用时新版本没续开通；或 endpoint 的模型字段被改过没保存对。

> 先用 `arkcli doctor infer-endpoint <ep-id>` 输出里的 `endpoint.model_info` 和 `endpoint.errors.overall.by_error_code` 一起看，多数能直接定位。

#### 诊断步骤

```
当前调用方是子账号还是主账号?（看 doctor account 的 identity 段）
 ├─ 子账号
 │   └─ 主账号下 doctor account 看 ARKFullAccess 是否挂 → 没挂就先挂
 │      挂了仍报 → 看下一支
 └─ 主账号 / 子账号已挂 ARKFullAccess
      └─ 控制台 → 模型广场 → 找对应模型 → 是否显示"已开通"?
           ├─ 未开通 ──► 修复 A
           └─ 已开通 ──► 检查 endpoint 绑定的模型 + 版本是否与开通的一致 → 修复 C
```

#### 修复方案

**修复 A —— 控制台开通模型**（账号未开通时最常见，对应 `ModelNotOpen`）：

1. 打开 `https://console.volcengine.com/ark/region:ark+cn-beijing/model`
2. 找到对应模型 → 点"开通服务" / "申请开通"
3. 等待开通生效（通常即时；个别模型需要审核）
4. 验证：`arkcli doctor infer-endpoint <ep-id>` 错误率应下降

> doctor 不替你点开通——这涉及计费授权与合同，必须用户 / 管理员显式确认。

**修复 B —— 子账号挂 ARKFullAccess**（`AccessDenied` / `OperationDenied.PermissionDenied`）：

让主账号或有 IAM 写权限的人：

1. 打开 `https://console.volcengine.com/iam/policymanage?PolicyType=System&Service=ark`
2. 找到 `ARKFullAccess`（写）或 `ARKReadOnlyAccess`（只读，仅查询场景）
3. 挂载到目标子账号
4. 验证：`arkcli doctor account` 的 `account.iam_system_policies` 应 ✓

**修复 C —— endpoint 模型版本对齐**（已开通但 endpoint 调旧 / 错版本时）：

1. 看 endpoint 当前模型版本：`arkcli infer endpoint get <ep-id>`
2. 与控制台已开通的版本对照
3. 如不一致：用 `arkcli infer endpoint update` 修正，或重新 `+deploy` 创建新 endpoint
4. 验证：重新调用，错误率应归零

#### 闭环验证

```bash
arkcli doctor infer-endpoint <ep-id>
# 看 endpoint.errors.overall.by_error_code 里 ModelAccessDenied / ModelNotOpen / AccessDenied 占比是否归零
```

如果还在，按决策树走下一支（A → B → C 顺序排查）。

---

### 1.3 `RateLimitExceeded`

> 限流（doctor 聚合 subtype，覆盖多个 429 码）。

**路由触发**：上方索引表中所有 429 码（`RateLimitExceeded.*` / `ModelAccount{Rpm,Tpm,Ipm}RateLimitExceeded` / `APIAccountRpmRateLimitExceeded` / `AccountRateLimitExceeded`）。

> `ServerOverloaded` / `RequestBurstTooFast` 是**服务侧资源紧张**（不是用户配额），单独看：先客户端退避；持续触发再提工单。`QuotaExceeded` / `SetLimitExceeded` / `InflightBatchsizeExceeded` 是**额度类**：看 message 走控制台开通 / 升档。

#### 解读

请求被限流——HTTP 429。Ark 上限流分两层（外加 API 接口层）：

- **接入点级（endpoint）**：每个 endpoint 自己的 RPM/TPM 限额。错误码 `RateLimitExceeded.Endpoint{RPM,TPM}Exceeded`。
- **模型级（account）**：账号下该模型的总 RPM/TPM/IPM。所有该模型的 endpoint 共享。错误码 `ModelAccount{Rpm,Tpm,Ipm}RateLimitExceeded`。
- **API 接口级**：账号下某接口的 RPM。错误码 `APIAccountRpmRateLimitExceeded`。

doctor 输出里 `endpoint.quota_pressure` 会告诉你两层各自压到哪个百分位。**真正决定限流由哪层触发的是较高的那一层**。

#### 诊断步骤

看 `arkcli doctor infer-endpoint <ep-id>` 的配额压力段：

```
接入点级 RPM/TPM 占比高（≥80%）?
 ├─ 是
 │   └─ 模型级也高（≥80%）?
 │        ├─ 是 ──► 模型级是真瓶颈 → 修复 A：提模型配额（控制台工单）
 │        └─ 否 ──► 接入点级单独瓶颈 → 修复 B：调流量分布 / 加副本，不必提模型配额
 └─ 否
     └─ 模型级高?
          ├─ 是 ──► 其它 endpoint 把模型配额吃了 → 修复 C：跑 doctor model <name> 找元凶
          └─ 否 ──► 限流不是真的——可能是瞬时突发或客户端重试风暴 → 修复 D：客户端退避
```

#### 修复方案

**修复 A —— 提模型级配额**（账号下该模型 RPM/TPM/IPM 撞顶）：

1. 控制台 → 模型广场 → 该模型详情 → "申请提升配额"（路径：`https://console.volcengine.com/ark/.../model`）
2. 工单写清：当前 RPM/TPM 实际峰值（doctor 已给）+ 期望值 + 业务理由
3. 等待审批通过
4. 验证：`arkcli doctor model <name>` 看上限是否抬升

**修复 B —— 调流量分布 / 加副本**（接入点级单独瓶颈）：

- 多副本：用 `arkcli infer endpoint update` 提该 endpoint 的并发上限（如果配额允许），或新建多个 endpoint 做客户端轮询
- 流量分布：客户端做 sharding，不要把全部流量打到单个 endpoint

> doctor 不直接动 endpoint 配置——`update` 走 [`arkcli-infer-endpoint`](../../arkcli-infer-endpoint/SKILL.md)。

**修复 C —— 找模型配额吃料元凶**：

```bash
arkcli doctor model <model-name> --window 1h
# 看 model.endpoint_total + endpoint top usage 找占用最高的 endpoint
```

定位后，要么提配额（修复 A），要么对吃料 endpoint 做 QoS 调整。

**修复 D —— 客户端退避**（瞬时突发 / 重试风暴 / `APIAccountRpmRateLimitExceeded` / `ServerOverloaded` / `RequestBurstTooFast`）：

- 加指数退避（exponential backoff）+ jitter
- 别在收到 429 后立即重试；建议初始等 1-2s，每次翻倍，上限 30s
- Doubao-Seed-2.0+ 模型遇到 `RequestBurstTooFast`：放缓提速速度，逐步增加请求量
- 长时间持续触发才考虑提配额

#### 闭环验证

```bash
arkcli doctor infer-endpoint <ep-id>
# 看 endpoint.errors.overall.by_error_code 中限流类码占比 + endpoint.quota_pressure
```

修复 A 起效需等审批；修复 B/D 立即可见。

---

## 2. 公共错误码

跨 VolcEngine 服务的公共错误码，来自 [VolcEngine OpenAPI 公共错误码](https://www.volcengine.com/docs/6369/68677?lang=zh)。这一类错误大多发生在请求**到达 Ark 之前**——SDK / arkcli 配置、签名、网络层故障；说明请求都还没进 Ark 业务逻辑。

> 与 §1 码名重叠的（如 `MissingParameter` / `AccessDenied`）：在 Ark 调用上下文中以 §1 表里的释义为准；本表是其它 VolcEngine 服务（IAM / VMP 等）调用时也可能遇到的通用版本。

| 错误码                       | CodeN  | 释义                                              | 处理方式                                                            |
|------------------------------|--------|---------------------------------------------------|---------------------------------------------------------------------|
| `MissingParameter`           | 100002 | 关键参数缺失（如 `Action` / `Version`）           | 检查 SDK 调用拼接 / 升级 SDK；Ark 上下文见 §1 表                     |
| `MissingRequestInfo`         | 100004 | 缺必要请求信息（如 `X-Date` 头）                  | 检查 SDK 完整性 / 升级版本                                          |
| `InvalidTimestamp`           | 100006 | 请求时间过期或签名时间异常                        | 校准本机时钟（NTP 同步）；用 UTC ISO 8601                           |
| `ServiceNotFound`            | 100007 | 请求的服务不存在                                  | 核对 Service 名称；检查 SDK / arkcli 版本                           |
| `InvalidActionOrVersion`     | 100008 | 接口或版本不存在                                  | 核对 Action / Version；查 API 文档                                  |
| `InvalidAccessKey`           | 100009 | AK 不合法                                         | 检查 AK 是否含空格 / 是否过期；重新登录：`arkcli auth login`        |
| `SignatureDoesNotMatch`      | 100010 | 签名计算结果不匹配                                | 检查 SK 是否正确；如手写签名换用官方 SDK                            |
| `AccessDenied`               | 100013 | 子用户权限不足                                    | Ark 上下文走 [§1.2](#12-modelaccessdenied) 修复 B；其它服务加 IAM 策略 |
| `InternalError`              | 100014 | 系统内部错误                                      | 退避重试；持续问题提工单带 RequestID                                |
| `InternalServiceTimeout`     | 100016 | 服务执行超时                                      | 退避重试；幂等请求可加大超时                                        |
| `FlowLimitExceeded`          | 100018 | 通用流控（非 Ark 自身的 RateLimitExceeded.\*）    | 客户端退避；持续触发提工单                                          |
| `ServiceUnavailableTemp`     | 100019 | 服务处于熔断状态                                  | 退避重试；持续触发提工单                                            |
| `InternalServiceError`       | 100023 | 服务故障（公共层；与 Ark 自身 500 同名码区分）    | 退避重试；持续问题提工单                                            |
| `InvalidAuthorization`       | 100024 | `Authorization` 头格式错（如缺 Region）           | 升级 SDK；如手写签名换用官方 SDK                                    |
| `InvalidCredential`          | 100025 | `Authorization` 头中 Credential 字段格式错        | 同上                                                                 |
| `InvalidSecretToken`         | 100026 | STS 临时凭证错误（过期或签名错）                  | 重新获取 STS token                                                  |
