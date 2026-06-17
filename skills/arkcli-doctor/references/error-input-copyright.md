# 输入素材版权拦截（input_copyright）

> 路由进来：错误码形如 `Input*SensitiveContentDetected.PolicyViolation`（`*` 取 Text / Image / Video / Audio）→ `reference=error-input-copyright`。先读 [`../SKILL.md`](../SKILL.md)。
>
> 可直接跑的真实例子（上面的 `*` 是占位、**别原样复制**）：
> ```bash
> arkcli doctor error "InputImageSensitiveContentDetected.PolicyViolation"
> arkcli doctor error "InputVideoSensitiveContentDetected.PolicyViolation"
> ```

## 1. 解读

你喂进去的某个输入素材涉版权（后缀 `.PolicyViolation` = 版权信号）。具体是哪个素材，要看错误码前缀 + 用户实际传了什么。

## 2. 诊断步骤（按素材类型分流）

| 错误码前缀 | 涉版权的素材 | 处理方向 |
|-----------|------------|----------|
| `InputText…` | prompt 文案 | 改写文案，去掉具体作品/角色/品牌名 |
| `InputImage…` | 参考图（画面人物 / 受版权画作） | 换合规素材；画面人物→见缺口 |
| `InputAudio…` | 参考音频 | 换无版权音频，或去掉参考音频 |
| `InputVideo…` | 参考视频——**需再消歧** | 见下方 |

**参考视频 1:N 消歧**（`InputVideoSensitiveContentDetected.PolicyViolation`）：
- 视频**音轨**版权（背景音乐）→ 用剪辑工具**去掉音轨**后重新上传。
- 视频**画面人物**版权 → 换素材；或对人物 deface（见缺口）。
- 分不清 → 先去音轨重试；仍失败多半是画面。

## 3. 修复方案

```bash
# 去掉涉版权素材后重试(以参考音频为例:直接不传它)
arkcli +gen --model <原模型> --input @<合规素材> "<改写后去掉版权指向的 prompt>"
```
- 视频音轨版权：用户**自行去音轨**后重传（doctor 不替用户剪视频）。
- 文案版权：把 prompt 里的具体作品名 / 角色名 / 品牌名换成泛化描述。

## 4. 能力缺口（deface 待上线）

参考视频 / 图的**画面人物**版权，最稳是对人物 deface。该能力**尚未上线**（`needs_backend: ["deface"]`）。当前只能换素材，**不要**承诺一键处理画面人物。

## 5. 闭环验证

去掉/替换涉版权素材后重试，不再返回 `.PolicyViolation` 即通过。视频音轨类需用户确认已去轨。
