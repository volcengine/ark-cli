# 给 arkcli-doctor 加错误码 / scope check / fixer

这份文档是 doctor 的**贡献入口**，划清三个最常见改动点的内聚边界——加错误码、加 scope check、加 fixer。

## 模块边界

```
internal/service/doctor/
  errorcode/diagnose.go        ← 错误码总册：Ark + 公共码查表（code → subtype → meta），委托 video 包
  video/diagnose.go            ← 生视频 5 个 wire-stable subtype（category=policy）
  scope/<name>/checks.go       ← 各 scope 的 check 实现（account / infer-endpoint / model / ...）
  fixer/<name>.go              ← Fixer 实现（A 型路由 / B 型代码）
  vmp/                         ← VMP query + preflight（共享）
  ark/                         ← Ark OpenAPI client（共享）

skills/arkcli-doctor/          ← 一个 skill 收口三个 domain
  SKILL.md                     ← 总入口（路径决策、scope 路由表、subtype 路由表）
  CONTRIBUTING.md              ← 本文件
  references/
    error-codes.md             ← 错误码总册（一个文件，2 大业务领域 + 生视频 5 子节）
    scope-infer-endpoint.md    ← 单接入点诊断细则（doctor infer-endpoint 的 reference）
    scope-model.md             ← 单模型诊断细则（doctor model 的 reference）
```

doctor 命令壳（`cmd/doctor/`）、`go:embed` skill 打包**不用动**——加错误码 / check / fixer 都不改它们。

> **历史形态**（已废弃）：错误码曾按 subtype 拆成 `references/error-input-real-face.md` 等独立文件。新形态统一收到 `references/error-codes.md`，`reference` 字段统一为 `error-codes`，agent 用 `subtype` 在文件内定位小节。

---

## 加一个错误码

`arkcli doctor error <code>` 是只读查表（不需登录、不消耗推理 token）。

错误码按业务领域分 2 个 H2 段：**方舟错误码 / 公共错误码**。每段顶部有索引表覆盖该领域所有官方码（含 HTTP / 释义 / 处理方式），高频或复杂的码再展开为完整小节——生视频 5 个 subtype 同根聚在一起作为 §1.1（下挂 5 个 H4）；`ModelAccessDenied` / `RateLimitExceeded` 作为 §1.2 / §1.3 H3。精调相关码归入 §1 索引表（与其它方舟码同表）。

> **注意：前置依赖闸门（VMP / TLS / TOS 跨服务授权）不是错误码**——它们由 doctor 自身在跑数据面诊断前主动探活，不在 `error-codes.md` 里覆盖。这类内容由 `arkcli doctor account` 输出 + 共享 skill 联合处理。

按落点深浅有 3 种贡献场景。

### 场景 A：已有 subtype 下，多一个真实错误码

新错误码的根因 / 修法和某个现有 subtype 一致——**只加一行 Go 映射 + 索引表加一行**。

按 subtype 落点选包：

- 生视频 5 个 wire-stable subtype（`input_real_face` / `input_copyright` / `input_content_safety` / `output_video_copyright` / `output_video_safety`）→ 编辑 `internal/service/doctor/video/diagnose.go` 的 `codeToSubtype`（per-subtype meta 模型）：

  ```go
  "SomeNewSensitiveContentDetected.PolicyViolation": SubtypeInputCopyright,
  ```

- 其余所有 subtype（Ark + 公共码）→ 编辑 `internal/service/doctor/errorcode/diagnose.go` 的 `codeMeta`（per-code meta 模型，每码自带 `root_cause`/`hint`，可直接抄索引表两列）：

  ```go
  "SomeNewAuthCode": {
      category:  CategoryArk,
      subtype:   SubtypeAuthError,
      rootCause: "...",
      hint:      "...",
  },
  ```

并在 `error-codes.md` 对应业务领域的索引表里加一行（"处理方式"列指向已存在的 §1.1.x / §1.2 / §1.3）。无需新小节。

### 场景 B：现有业务领域下新 subtype（完整小节）

例如：在 §1 方舟错误码下把 `OperationDenied.ServiceOverdue` 从默认建议升级为完整小节；在 §1.1 生视频家族下加新模态的 subtype。

**包选择同场景 A**：生视频 subtype 在 `video` 包；Ark / 公共其它 subtype 在 `errorcode` 包。

四步（以 `errorcode` 包为例，video 包形状对称）：

1. **加 subtype 常量**（`errorcode/diagnose.go`）：
   ```go
   SubtypeServiceOverdue = "service_overdue"
   ```

2. **加 codeMeta 条目**（每码一份 meta，category 字段必填）：
   ```go
   "OperationDenied.ServiceOverdue": {
       category:  CategoryArk,
       subtype:   SubtypeServiceOverdue,
       rootCause: "账单逾期，所有数据面调用被风控",
       hint:      "open https://console.volcengine.com/finance/recharge",
   },
   ```

3. **同 subtype 下加更多原生码** —— 直接复制条目，subtype 字段写同一个常量即可。

4. **在 `error-codes.md` 加完整小节** + **更新该领域的索引表**：
   - 方舟错误码、生视频家族（§1.1）的新 subtype → H4 挂在 §1.1 下，继续 5 段模板
   - 方舟错误码、其它新 subtype → H3 紧随 §1.1/1.2/1.3 之后排号
   - 公共错误码领域 → 一般留在索引表里，如确需展开则加 §2.x H3
   - 同时把表里该码的"处理方式"列从默认建议改成 `见 §X.Y`

   完整小节套用[模板](#完整小节模板5-段)。**TOC 也补一行**。

### 场景 C：全新业务领域（顶层 H2）

整个新领域**与现有领域不共享根因模型与命名空间**——例如未来的 Agent / Batch / Eval 业务上线。

五步：

1. 在 `errorcode/diagnose.go` 加 `Category` 常量（如 `CategoryAgent Category = "agent"`）。
2. 在领域内加首批 subtype（同场景 B 的 1-3 步）。
3. **在 `error-codes.md` 加 H2 段落**（如 `## 4. Agent 错误码`），写引言 + 索引表 + 按需挂完整小节。**TOC 加一行**。
4. **在 [`SKILL.md`](SKILL.md) 的 subtype 路由表加行**。
5. 检查是否需要新的 scope reference（如 `references/scope-agent.md`）配套——错误码常常和 scope 命令成对出现。

---

## 完整小节模板（5 段）

每个完整 subtype 在 `error-codes.md` 是一个 H3 小节（生视频家族例外，是 §1.1 下的 H4）。结构固定：

```markdown
### <编号> `<subtype>`

> <这类错误的一句话名字>

**路由触发**：<什么样的错误码 / 命令输出会路由到这里>。

#### 解读
<用户友好地解释这个码意味着什么>

#### 诊断步骤
<怎么定位；复合码给消歧决策树，可让 agent 直接看用户素材>

#### 修复方案
<给可跑的 +gen / arkcli 命令；用户动作类说清要用户做什么>

#### 能力缺口（可省）
<如依赖 deface 等未上线能力，如实标注，不假装能修>

#### 闭环验证
<修完跑什么命令确认>
```

> 之所以把所有错误码放一个 markdown 文件而非每码一文件：错误码长尾稀疏，每码一文件会让 `references/` 指数膨胀；单文件 + 业务领域索引表 + 完整小节锚点跳转既保留了"按需读一段"的能力，也让"加错误码"退化成"加一行表 + 选择性加一节"。

---

## 加一个 scope check

scope check 是某个 doctor 命令（如 `arkcli doctor model`）输出里的某一项。

1. 在 `internal/service/doctor/scope/<name>/checks.go` 实现 `Check` 接口。命名约定：`<scope>.<group>.<item>`（如 `account.ecosystem.vmp.cross_service_auth`、`model.account_quota_pressure`）。
2. 在该 scope 的 `Runner.Diagnose` 里把 check 注册到执行序列。
3. 在 `references/scope-<name>.md` 的 check 清单里加一行（用户可见的 ID + 看什么 + 失败含义）。
4. 跑 scope 集成测试。

> **不要在 check 里直接 mutate**——check 是只读体检。需要修复行为时实现 Fixer。

---

## 加一个 fixer

Fixer 分两型（优先 A 型）：

- **A 型 — Skill-driven**（推荐）：本质是把"按步骤做"沉淀到 reference / scope reference，Fixer 负责路由 + 渲染 dry-run，doctor 进程**不 mutate**。比如『跳控制台开通模型』『一键授权 VMP』。
- **B 型 — Code-driven**：必须本地代码执行（凭证注入 / 二次签名）。`Apply()` 真正调 API。`Reversibility` 通常 `Mutates` 或 `Destructive`。

实现：

1. 在 `internal/service/doctor/fixer/<name>.go` 实现 `Fixer` 接口。`ID()` 推荐与对应错误码 subtype 同名（如 `fix-model-access-grant`）。
2. 在所属 scope 的 `Runner.Fixers()` 里注册。
3. **A 型**：在 `error-codes.md` 对应小节的「修复方案」段（或 `scope-*.md` 的对应 finding 段）写清步骤；不需要新文件。
4. **B 型**：实现完整 `DryRun()` + `Apply()`，`Reversibility` 标得准确（默认拒绝 `Destructive`）。
5. 加单元测试覆盖 dry-run / apply / reversibility 三档。

> **硬约束：用户二次确认**。无论 A/B 型，`--fix` 默认走 dry-run 给用户看，必须显式 `--apply`（或交互确认）才真正改。

---

## 提交前必跑（缺一不可）

```bash
# 1) lockdown 测试：抓"加了码漏配 meta" / "subtype/section 锚点对不上" / "category 漏注册"
go test ./internal/service/doctor/...

# 2) 编译 + embed 打包（reference 文件被 go:embed 收进二进制）
go build ./...

# 3) skill 结构校验
cd skill-creator && python3 -m scripts.quick_validate ../skills/arkcli-doctor

# 4) 真机手验
arkcli doctor error "<你的新错误码>"
arkcli doctor <scope> [<id>]            # 加了 check / fixer 时
```

---

## 硬约束（别破）

- **doctor 只读、不 mutate**（除显式 `--fix --apply`）：诊断本身不重新生成、不改资源、不烧推理 token。
- **缺口要诚实**：依赖未上线能力（如 `deface` / `request_id 反查`）一律用 `needsBackend` / 文档明示，Hint 里不许假装能一键修。
- **subtype / category 是 wire 稳定标识**：改名 / 删除属 breaking change；新增不算。
- **修复用原生命令**：doctor 不自带 +gen / +chat 引擎，修复命令一律是 `arkcli <existing-command> …`，跨界细节归对应 owning skill（[`arkcli-gen`](../arkcli-gen/SKILL.md) / [`arkcli-deploy`](../arkcli-deploy/SKILL.md) / [`arkcli-infer-endpoint`](../arkcli-infer-endpoint/SKILL.md) 等）。
- **reference 字段统一是 `error-codes`**：每码独立 `.md` 的旧形态已废弃。Agent 用 `subtype` 在 `error-codes.md` 里定位小节。
- **越权边界**：所有外部 API 用户 AK/SK 签名（Ark / VMP / IAM）；不读平台内部数据（k8s / 调度器等）；不跨账号横比。
