# 给 arkcli-doctor 加一个生视频错误码

这份文档是 doctor 生视频诊断的**贡献入口**。它划清了一个内聚的模块边界：你只需要动下面两处，就能让 `arkcli doctor error <新码>` 把一个新的拦截错误码翻译成结构化诊断 + 修复方案。

## 模块边界（你的领地）

```
 internal/service/doctor/video/diagnose.go   ← 诊断逻辑(查表): 错误码→subtype→根因/Hint/缺口
 skills/arkcli-doctor/references/error-*.md   ← 知识层: 给 agent 的解读 + 修复编排 + 验证
```

doctor 命令壳（`cmd/doctor/`）、skill 入口（`SKILL.md`）、`go:embed` 打包都**不用动**——加错误码不改它们。

## 场景 A：已有 subtype 下，多一个火山错误码

如果新错误码的根因/修法和某个现有 subtype 一致（5 个 subtype 见 `diagnose.go` 顶部），只加一行映射：

```go
// internal/service/doctor/video/diagnose.go 的 codeToSubtype
"SomeNewSensitiveContentDetected.PolicyViolation": SubtypeInputCopyright,
```

完事。无需新 reference（复用现有的）。

## 场景 B：全新一类拦截（新 subtype）

四步，全部在上面两个文件里：

1. **加 subtype 常量**（`diagnose.go`）：
   ```go
   SubtypeOutputAudioCopyright = "output_audio_copyright"
   ```
2. **加 meta 条目**（`subtypeMeta`）——根因 / Hint（给可跑的 `+gen` 命令）/ rules / 缺口 / reference：
   ```go
   SubtypeOutputAudioCopyright: {
       rootCause:    "……",
       hint:         "arkcli +gen … --generate-audio false",
       rules:        []string{"……"},
       needsBackend: nil, // 没缺口就留空;有未上线能力(如 deface)才填
       reference:    "error-output-audio-copyright",
   },
   ```
3. **加错误码映射**（`codeToSubtype`），注意后缀语义（`.PrivacyInformation`=隐私 / `.PolicyViolation`=版权 / 无后缀=内容安全）：
   ```go
   "OutputAudioSensitiveContentDetected.PolicyViolation": SubtypeOutputAudioCopyright,
   ```
4. **加 reference 文件** `references/error-output-audio-copyright.md`，套用下面的模板。

> 文件名 = meta 里的 `reference` 值 + `.md`，**必须逐字一致**（有双向校验测试盯着）。

## reference 模板（5 段）

```markdown
# <这类拦截的名字>（<subtype>）

> 路由进来:`arkcli doctor error "<错误码>"` → reference=<reference>。先读 ../SKILL.md。

## 1. 解读        <用户友好地解释这个码意味着什么>
## 2. 诊断步骤    <怎么定位;复合码给消歧决策树,可让 agent 直接看用户素材>
## 3. 修复方案    <给可跑的 +gen 命令;用户动作类说清要用户做什么>
## 4. 能力缺口    <如依赖 deface 等未上线能力,如实标注,不假装能修>(无缺口可省)
## 5. 闭环验证    <修完跑什么命令确认>
```

## 提交前必跑（缺一不可）

```bash
# 1) lockdown 测试:会抓"加了码漏配 meta"/"后缀消歧错乱"/"reference 文件名对不上"
go test ./internal/service/doctor/...

# 2) 编译 + embed 打包(reference 文件会被 go:embed glob 自动收进二进制)
go build ./...

# 3) skill 结构校验
cd skill-creator && python3 -m scripts.quick_validate ../skills/arkcli-doctor

# 4) 真机手验
arkcli doctor error "<你的新错误码>"
```

## 硬约束（别破）

- **doctor 只读、不 mutate**：诊断不重新生成、不改资源、不烧推理 token。修复命令交用户确认后执行。
- **缺口要诚实**：依赖未上线能力（如 `deface`）一律用 `needsBackend` 标注，Hint 里不许假装能一键修。
- **subtype 是 wire 稳定标识**：改名 / 删除属 breaking change；新增不算。
- **修复用 `+gen` 原生参数**：doctor 不自带生成引擎，修复命令一律是 `arkcli +gen …`，生成细节归 [`arkcli-gen`](../arkcli-gen/SKILL.md)。
