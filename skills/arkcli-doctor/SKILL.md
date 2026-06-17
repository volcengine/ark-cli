---
name: arkcli-doctor
version: 0.1.0
description: "arkcli doctor 生视频诊断:把火山生视频生成被内容安全 / 版权策略拦截返回的错误码,翻译成结构化根因 + 可执行的 +gen 修复方案。当用户的视频生成报 InputImage/InputVideo/OutputVideo…SensitiveContentDetected 等拦截错误码、问『为什么被拦 / 怎么修』、或 +gen 生成失败需要诊断修复时使用。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli doctor error --help"
---

# arkcli Doctor（生视频诊断）

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)（认证闸门、命令选择顺序、输出与安全规则）。**

## 适用场景

- 用户的 seedance 视频生成被内容安全 / 版权策略拦截，拿到一个错误码，问"为什么被拦 / 怎么修"
- `arkcli +gen` 生成视频失败，错误码形如 `OutputVideoSensitiveContentDetected.PolicyViolation`
- 输入参考图 / 视频含真实人脸触发隐私拦截；输出视频因人物版权或自动配乐版权被拦

> 只覆盖**生视频拦截**类错误码。其它 doctor 体检（CLI / account / model / infer-endpoint）不在本 skill 范围。

## 核心范式（先理解再执行）

doctor 把"诊断"做成确定性的 Go 查表，把"解读 + 修复编排"留给本 skill。一次诊断是这样流动的：

```
 用户给错误码(+可选上下文: prompt / 有没有参考图 / 有没有参考音频)
        │
        ▼  ① 跑诊断(只读, 不消耗推理 token, 不需登录)
   arkcli doctor error <code>
        │  返回结构化 JSON: subtype / root_cause / hint / rules / needs_backend / skill / reference
        ▼  ② 按 reference 字段读对应细则
   references/<reference>.md      （如 error-output-video-copyright.md）
        │
        ▼  ③ 跟着 reference 给用户: 根因 + 一条可跑的 +gen 修复命令
   ★ 复合错误码(rules 有多个候选)→ 用 reference 里的判定步骤 + 用户的实际输入消歧, 择一
   ★ 修复命令会消耗推理 token → 让用户确认后再跑(本 skill 不替用户执行)
```

## 第一步：跑诊断

```bash
arkcli doctor error "<错误码原值>"
# 例:
arkcli doctor error "OutputVideoSensitiveContentDetected.PolicyViolation"
```

返回字段（结构化 JSON，**按需消费；未来可能新增字段，别按这张表做严格 schema 校验**）：

| 字段 | 含义 |
|------|------|
| `category` | 固定 `policy`（内容安全与版权同属该类；未来扩展/过滤维度用） |
| `subtype` | 归并后的 policy 子类型（5 种之一） |
| `code` | 命中的火山真实错误码（原值回显，用于确认输入原码） |
| `root_cause` | 根因，用户友好的一句话 |
| `hint` | 修复指引（可跑的 `+gen` 命令，或需用户做的动作） |
| `rules` | 触发信号候选；**多个 = 复合错误码，需消歧** |
| `needs_backend` | 依赖但尚未上线的能力（如 `deface`）——**不要假装能自动修** |
| `skill` | 固定 `arkcli-doctor`（就是本 skill） |
| `reference` | 该读哪个 references 文件 |

## 第二步：按 reference 读细则

`arkcli doctor error` 返回的 `reference` 字段直接告诉你读哪个文件。路由全景（5 个 subtype）：

| subtype | reference 文件 | 何时命中 |
|---------|---------------|----------|
| `input_real_face` | [`references/error-input-real-face.md`](references/error-input-real-face.md) | 输入参考图/视频含真实人脸（隐私，`.PrivacyInformation`） |
| `input_copyright` | [`references/error-input-copyright.md`](references/error-input-copyright.md) | 输入素材涉版权（`Input…PolicyViolation`） |
| `input_content_safety` | [`references/error-input-content-safety.md`](references/error-input-content-safety.md) | 输入触发内容安全（`Input…` 无后缀） |
| `output_video_copyright` | [`references/error-output-video-copyright.md`](references/error-output-video-copyright.md) | 输出视频版权（`Output…PolicyViolation`，**1:N 需消歧**） |
| `output_video_safety` | [`references/error-output-video-safety.md`](references/error-output-video-safety.md) | 输出视频内容安全（`Output…` 无后缀） |

## 第三步：给用户答案

- 摘 `root_cause` + reference 里的修复方案，给用户**一条可跑的 `+gen` 命令**，不要复述整段 JSON。
- 复合错误码：用 reference 的判定步骤 + 用户实际输入（有没有参考图 / prompt 是否含人物 / 有没有参考音频）消歧，**只给最可能的那一条**，并说明备选。
- `needs_backend` 非空（如 `deface`）：如实告诉用户该能力尚未上线，给出当前可行的替代（如换合规参考图），**不要假装能一键修**。

## 约束

- **只读**：`arkcli doctor error` 不重新生成、不改任何资源、不消耗推理 token。真正的修复（重新 `+gen`）由用户确认后执行。
- **不替用户烧 token**：修复命令一律先给用户看、等确认。
- 修复用的就是 `arkcli +gen` 的原生参数（`--generate-audio false` / `--input @<图>` 等），生成走 [`arkcli-gen`](../arkcli-gen/SKILL.md) 工作流；本 skill 只负责诊断 + 给命令。
- 新增错误码 = 加映射 + 加一份 reference，见 [`CONTRIBUTING.md`](CONTRIBUTING.md)。
