# 输出视频内容安全拦截（output_video_safety）

> 路由进来：`arkcli doctor error "OutputVideoSensitiveContentDetected"`（**无后缀**）→ `reference=error-output-video-safety`。先读 [`../SKILL.md`](../SKILL.md)。

## 1. 解读

模型**输出**的视频触发了内容安全策略（无 `.PolicyViolation` 后缀 = 内容安全，不是版权）。和"输入内容安全"不同的是：这里输入可能看着没问题，是模型按 prompt / 素材**生成**出了敏感画面。

## 2. 诊断步骤

- 输出敏感几乎都源于**输入引导**：回看 prompt 和参考素材里有没有可能被模型放大成敏感画面的描述 / 元素。
- 若有参考视频 / 图，看它们的风格、动作、场景是否本身偏敏感。

## 3. 修复方案

调整输入 prompt / 素材中会导向敏感画面的部分后重试（用户动作）：
```bash
arkcli +gen --model <原模型> --input @<更安全的素材> "<去掉会被放大成敏感画面的描述后的 prompt>"
```
- 收敛 prompt 里可能引发敏感画面的动作 / 场景 / 着装等描述。
- 替换偏敏感的参考素材。

## 4. 闭环验证

调整后重试，输出不再被拦即通过。若同一 prompt 反复输出敏感，提示用户换主题或显著弱化相关描述。
