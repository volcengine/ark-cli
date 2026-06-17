# 输入内容安全拦截（input_content_safety）

> 路由进来：错误码形如 `Input*SensitiveContentDetected`（`*` 取 Text / Image / Video / Audio，**无后缀**）→ `reference=error-input-content-safety`。先读 [`../SKILL.md`](../SKILL.md)。
>
> 可直接跑的真实例子（上面的 `*` 是占位、**别原样复制**）：
> ```bash
> arkcli doctor error "InputTextSensitiveContentDetected"
> arkcli doctor error "InputImageSensitiveContentDetected"
> ```

## 1. 解读

输入触发了内容安全策略（无 `.PolicyViolation` / `.PrivacyInformation` 后缀 = 既非版权也非隐私，而是敏感内容本身）。这类**没有自动修复**——只能调整输入。

## 2. 诊断步骤

- 按错误码前缀定位是哪个输入：`InputText…`=文案，`InputImage…`=参考图，`InputVideo…`=参考视频，`InputAudio…`=参考音频。
- 看该输入里有什么敏感内容（暴力 / 色情 / 政治敏感 / 违禁等）。你可以直接看用户的素材判断。

## 3. 修复方案

调整输入中的敏感内容后重试——这是**用户动作**，不是可自动跑的 `+gen` 命令：
- 文案：改写 prompt，去掉敏感表述。
- 图 / 视频 / 音频：换一个不含敏感内容的素材。

## 4. 闭环验证

调整输入后重试，不再返回该错误码即通过。若反复命中同类拦截，提示用户该主题的内容安全风险较高、建议换方向。
