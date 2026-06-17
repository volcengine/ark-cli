# 输入真人脸 / 隐私拦截（input_real_face）

> 路由进来：`arkcli doctor error "InputImageSensitiveContentDetected.PrivacyInformation"`（或 `InputVideo…PrivacyInformation`）→ `reference=error-input-real-face`。先读 [`../SKILL.md`](../SKILL.md)。

## 1. 解读

你作为参考图 / 参考视频喂进去的素材里，检测到了**真实人脸**，触发隐私保护拦截（后缀 `.PrivacyInformation` 就是隐私信号，区别于版权 `.PolicyViolation`）。

## 2. 诊断步骤

- 直接看用户提供的参考图 / 视频里有没有清晰的真实人脸（你是多模态的，能直接判断）。
- 确认这张人脸是不是用户**有意要保留**的人物身份——这决定能不能直接换图。

## 3. 修复方案

**换用合规 / 虚拟人像作参考图后重试**：
```bash
arkcli +gen --model <原模型> --input @<合规图> "<原 prompt>"
```
- 若用户只是要"某种长相"而非特定真人：用 seedream 生成一张虚拟人像替换（走 [`arkcli-gen`](../../arkcli-gen/SKILL.md)）。
- 若用户坚持要还原这张真人脸：见第 4 节。

## 4. 能力缺口（deface 待上线）

对真人脸做**去身份化（deface）**后再喂进去，可以在"保留相似长相"的同时通过隐私拦截。该能力**尚未上线**（`needs_backend: ["deface"]`）。当前只能换合规/虚拟人像，**不要**承诺一键 deface。

## 5. 闭环验证

重试后不再返回 `.PrivacyInformation` 即为通过。
