# billing list

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **跟 `usage stats` 的关系：** `stats` 出"用了多少 token / 多少次请求"（推理量），`billing list` 出"花了多少钱"（结算金额）。两者数据源不同：stats 走 ARK BFF 聚合，billing 走火山**计费中心**拆账。计费中心数据按账期（账单月）出，**T+1 出账**（次日凌晨可见前一天明细），月初对账时建议拉前一个完整账期。
>
> 这也是为什么 billing 是 `arkcli billing` 顶层命令组而不是 `arkcli usage billing` 子命令: 数据源 / 时效 / 鉴权边界都跟 usage 不一样, 视为同一个命令组会掩盖差异。

查询指定账期（YYYY-MM）或月范围内的拆分账单明细。**结果始终是完整范围数据**，service 层强制全量分页拉完，不暴露 Limit / Offset，避免"翻一页就拿去对账"的脚枪。

> **重要 — `--start` 单独使用 = 单月查询**（不是"从 `--start` 到现在"的开放区间）。`billing list` 跟 `usage stats` 看着相像但语义不同：
> - **`usage stats`** 是日级时间范围（`--start --end` 两端,缺一个端点会用今天补,所以 `--start 2026-05-01` 单传 = 从 5/1 到今天）
> - **`billing list`** 是月单位,wire 上是单 `BillPeriod`字段,**单传 `--start 2026-05` 就是查 5 月这一个月**;要查多月范围必须显式加 `--end`(例:`--start 2026-03 --end 2026-05` 查 3-5 月,客户端 fan-out 模拟)。

## 命令

```bash
# 查整账期月汇总(最常用) — 默认账号维度 ARK 推理全量(7 个 arkProducts),不含订阅
arkcli billing list --start 2026-05
#   ↳ stderr: no scope filter; querying account-wide ARK billing (auto-filtered to ARK product series; pass --product to override / include ark_subscription); use --endpoint/--apikey/--apikey-sid to drill down or --split-dim to slice by dimension

# 查多个账期(闭区间,--end 包含) — 常用于季度/年度对账
arkcli billing list --start 2026-03 --end 2026-05

# 按天拆分(必须传 --day 在同月内,--day 不兼容 --end 范围)
arkcli billing list --start 2026-05 --interval day --day 2026-05-15

# 看每条结算明细
arkcli billing list --start 2026-05 --interval detail

# ─── Scope 控制(单值三选一,详见下面 "Scope 解析" 表) ───
# 1. 单 EP 排查 (用户反馈"这个 EP 花了多少"时切到具体 ep_id)
arkcli billing list --start 2026-05 --endpoint ep-20260415xxx-xxxxx

# 2. 显式查具体 API Key (完整 ark- 字符串, 自动经 arkbff 反查 SID)
arkcli billing list --start 2026-05 --apikey ark-xxxxxxxxxxxx

# 3. 直接传 API Key SID (advanced; 跳过 arkbff 反查, 自己从 `apikey.list` 抽 SID 时用)
arkcli billing list --start 2026-05 --apikey-sid apikey-20260520143458-gv2nj

# ─── 账号维度切片(无 ItemID,按维度拆账) ───
# 4. 账号下每个 EP 各花了多少
arkcli billing list --start 2026-05 --split-dim endpoint

# 5. 账号下每个 API key 各花了多少
arkcli billing list --start 2026-05 --split-dim apikey

# 6. 只看我创建的 API key 的费用 (SSO 子用户视角; 不含订阅类账单)
arkcli billing list --start 2026-05 --mine

# ─── 过滤 ───
# 按产品编码过滤(ARK 大模型推理是 ark_bd)
arkcli billing list --start 2026-05 --product ark_bd

# 跳过折后价为 0 的行
arkcli billing list --start 2026-05 --ignore-zero

# 按账单类型 / 计费模式过滤
arkcli billing list --start 2026-05 --bill-category consume-use
arkcli billing list --start 2026-05 --billing-mode 2
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--start` | **是** | string | **单个**账期 `YYYY-MM`(距今 24 个月内)。单独使用 = 查这一个月;配 `--end` = 范围 |
| `--end` | 否 | string | 范围终点 `YYYY-MM`(闭区间)。**省略 = 单月查询(不是到今天)**。范围最多 24 个月 |
| `--day` | 否 | string | 账单日 `YYYY-MM-DD`，必须在 `--start` 同月内；**仅 `--interval=day\|detail` 生效**；**不兼容 `--end` 范围**（month-day 语义模糊） |
| `--interval` | 否 | string | 聚合粒度：`month`（默认，账期汇总）/ `day` / `detail` |
| `--endpoint` | 否 | string | 过滤到具体 EP ID（`ep-...`）。直接传给上游 SplitItemID。**跟 `--apikey` / `--apikey-sid` 互斥** |
| `--apikey` | 否 | string | 过滤到具体 API Key 值（`ark-...` 完整字符串）。**shortcut 自动经 arkbff `OpenGetApiKeySID` 反查成 SID** 后填给上游 SplitItemID（apikey 维度上游期待 SID 而非 secret 字符串）。跟 `--endpoint` / `--apikey-sid` 互斥 |
| `--apikey-sid` | 否 | string | 直接传 API Key SID（`apikey-{ts}-{xxx}`，从 `arkcli api apikey.list` 抽出来）。**跳过 arkbff 反查**，advanced 路径。跟 `--endpoint` / `--apikey` 互斥 |
| `--split-dim` | 否 | string | 账号维度切片:`apikey` / `endpoint` / 空(默认账号汇总)。**显式 scope 自带 dim,split-dim 必须一致或不传** |
| `--product` | 否 | string[] | 按火山产品编码过滤（可重复，ARK 大模型推理是 `ark_bd`） |
| `--project` | 否 | string[] | 按项目 ID 过滤（可重复，成本归集维度;一个 project 内可能有多个 product） |
| `--instance-no` | 否 | string | 过滤到具体计费实例 ID |
| `--bill-category` | 否 | string[] | 账单类型，可重复，可选值见下表 |
| `--billing-mode` | 否 | string[] | 计费模式，可重复：`1` 包年包月 / `2` 按量 / `3` 合同 / `4` 履约 |
| `--ignore-zero` | 否 | bool | 跳过折后价 = 0 的行（实测 2025-12 一个月能少 ~10k 行） |
| `--mine` | 否 | bool | 仅看当前 SSO 子用户名下资源产生的账单。维度由 `--mine-by` 决定(默认 `apikey`)。**跟 `--endpoint` / `--apikey` / `--apikey-sid` / `--split-dim` 互斥**。订阅类账单(Agent Plan / Coding Plan)不含,因为账号级归属没 IAM owner |
| `--mine-by` | 否 | string | `--mine` 的维度: `apikey`(默认,**我的消费** — 我创建的 API Key 产生的所有账单,跨全部 7 个 ARK 推理产品) / `endpoint`(**我的基础设施** — 我 `+deploy` 的 endpoint 上产生的账单,跨 LLM / 图片 / 视频等多种产品,详见下方"`--mine-by=endpoint` 跨产品覆盖"段) |

### Scope 解析(必读)

**默认 = 账号维度全量 ARK 推理账单**(不传任何 scope flag → service 自动锁 `Product` 到 7 个 ARK 推理产品 [`ark_bd`, `ark_open_source_llm`, `Doubao-image-generation`, `Doubao-Seedream`, `db_research`, `ark_ym_sanfang`, `ark_sm_sanfang`],防 EIP/TOS/VDB 等非 ARK 产品脏返回。**不含 `ark_subscription`(Agent Plan / Coding Plan)**,跟主流推理混在一起容易让对账失真;专门看订阅类要显式 `--product ark_subscription`。stderr 会出软提示)。要进一步缩范围,三选一传:

| 入参形态 | 实际 scope | stderr hint |
|---|---|---|
| 裸跑(没传 scope flag) | 账号维度 ARK 推理全量(Product 自动锁 7 个 arkProducts;不含 `ark_subscription`) | `no scope filter; querying account-wide ARK billing (auto-filtered to ARK product series; pass --product to override / include ark_subscription); use --endpoint/--apikey/--apikey-sid to drill down or --split-dim to slice by dimension` |
| `--endpoint ep-X` | 单 EP 过滤(SplitItemID=ep-X), 排查具体 EP 用 | 无 |
| `--apikey ark-X` | 显式查具体 key(经 arkbff 反查 SID) | 无 |
| `--apikey-sid apikey-X` | 直接喂 SID 给 SplitItemID(跳过 arkbff 反查) | 无 |
| `--split-dim apikey/endpoint`(配合裸跑) | 账号维度按 apikey / endpoint 切片汇总(无 SplitItemID;apikey 走显式 SplitDimension={...:apikey_id},endpoint 是 ark_bd 隐式默认仅靠 Product=[ark_bd] 过滤) | 同裸跑 hint |

**互斥**: `--endpoint` / `--apikey` / `--apikey-sid` 三方互斥(SplitItemID 单值)。`--split-dim` 与显式 scope 必须一致或不传 — `--apikey ark-X --split-dim endpoint` 这种矛盾组合直接 ErrValidation。

**ark-\* → SID 反查只走 arkbff 一条路**(`OpenGetApiKeySID`)。早期版本曾有 `ListPersonalKeys + tail-suffix` fast path 试图省一次 BFF 调用,但 ListApiKeys 上游返回的 Key 是 masked 形式 (`abcd****wxyz`),全量 ark key 末 12 字符的 `HasSuffix` 在生产**永远命不中**(mask 中间的 `*` 直接破坏 suffix);加上 12 字符后缀本身也不是身份证明(撞尾概率虽低但会悄悄返错 SID)。所以彻底删掉,BFF 一条路 — 一次毫秒级 RPC 换确定性的归属校验。

**何时 ErrValidation 拒绝**:
- 三个 scope flag 同时传超过一个
- `--apikey` 字符串太短(< 16 字符)→ 早 fail 不浪费 BFF round-trip
- `--apikey` 经 arkbff 反查后归属校验失败(key 不属于本账号 / 不存在)→ 提示切 profile 或直传 `--apikey-sid`
- `--split-dim` 取值非 `apikey` / `endpoint` / 空
- `--split-dim` 与显式 scope 维度不一致(如 `--endpoint ep-X --split-dim apikey`)

**不支持的视图**(刻意收紧避免脚枪):
- 跨 key per-apikey + per-endpoint 联合聚合: 用 `arkcli api apikey.list` 列全后逐 key 跑 billing 客户端合并

### 视图能力边界(必读)

不同 scope 视图覆盖的账单子集不一样,**Agent Plan / Coding Plan 等订阅类账单只能在账号维度看,任何 IAM/EP/apikey 维度都看不到**(spec 限制,见 [`arkProducts` 注释](../../../internal/service/billing/service.go))。

| 视图 | 含 ARK 推理(`ark_bd` 等 7 个)? | 含订阅类(`ark_subscription`)? | 含非 ARK(EIP/TOS/VDB)? |
|---|---|---|---|
| 裸跑(默认账号维度) | ✅ | ❌(自动锁过滤) | ❌(自动锁过滤) |
| `--product ark_subscription`(显式) | ❌ | ✅ | ❌ |
| `--product ark_bd --product ark_subscription` | ✅ | ✅ | ❌ |
| 完全不锁(走 raw API,不传 Product) | ✅ | ✅ | ✅ |
| `--endpoint ep-X` | 该 EP 在任一 ARK 产品的行(LLM 走 ark_bd / 图片走 Doubao-image-generation / 视频走 ark_bd 等,见下方"跨产品覆盖") | ❌(订阅没 EP 概念) | ❌ |
| `--apikey ark-X` / `--apikey-sid` | 仅该 key 的行 | ❌(订阅没 apikey 概念) | ❌ |
| `--split-dim apikey` | 全 ARK 推理按 apikey 拆 | ❌(SplitDimension 拒) | ❌ |
| `--split-dim endpoint` | 全部 ARK 产品按 EP 拆(SplitItemID=ep-X 跨产品 unique) | ❌ | ❌ |
| `--mine --mine-by=apikey`(默认) | 我创建的 apikey 的行(跨全部 7 个 ARK 推理产品) | ❌ | ❌ |
| `--mine --mine-by=endpoint` | 我部署的 ep 上的行(跨 LLM / 图片 / 视频等多种产品) | ❌ | ❌ |

### `--mine-by=endpoint` 跨产品覆盖(实测确认 2026-06-11)

**关键事实**: SplitItemID `ep-{ts}-{xxx}` 是 ARK endpoint 的**跨产品 unique 主键**——`+deploy` 出 LLM、图片、视频等模型用同一格式的 ep_id, 但账单**归属哪个 product 跟模型 family 没有简单对应**:

| 部署的模型 | 实际账单 Product 字段 |
|---|---|
| `doubao-seed-1-6` / `doubao-seed-2-0-pro` (LLM) | `ark_bd` |
| `doubao-seedream-4-5` / `doubao-seedream-5-0` (图片) | `Doubao-image-generation`(**不是** `Doubao-Seedream`!) |
| `doubao-seedance-2-0-fast` (视频) | `ark_bd`(**不是** `Doubao-Seedance`!) |
| `glm-5-1` / `deepseek-v4-pro` 等开源 LLM | `ark_bd` 或 `ark_open_source_llm` |

**所以**: 直接用 SplitItemID 跨产品匹配比按 Product 过滤更准确。service 层从前在 endpoint 维度强锁 `Product=[ark_bd]` 是 silent-data-loss bug——把 seedream 等图片 ep 的账单丢掉了; 现已改为不锁 Product, 让 wire 按 SplitItemID 自然跨产品命中。

**例外**:
- `arkProducts` 列表里的 `Doubao-Seedream` 跟 `+deploy` seedream 模型出来的 endpoint **不是同一个产品**: `Doubao-Seedream` product 是平台直发模型 API 的计费类目(SplitItemID 形态另算), `+deploy` 出来的 seedream endpoint 走 `Doubao-image-generation`。
- `ark_open_source_llm` / `Doubao-image-generation` 还看到过 `ep-m-{ts}-{xxx}` 形态的 SplitItemID(model-id 而非 endpoint-id)和 `-` (不归属), 这部分不在 endpoint 视图里, 但归属本来就不在 endpoint 维度。
- `ark_subscription` (Agent Plan / Coding Plan) 没 endpoint 也没 apikey 维度, 全账号级归属。

**用户视角实务**:
- 想看"我所有 ARK 推理消费"(LLM + 图片 + 视频 + 开源 LLM 等) → `--mine --mine-by=apikey`(以 apikey 归属为主, 跨全部产品)
- 想看"我部署的 endpoint 上的总消费"(含他人调我的 ep) → `--mine --mine-by=endpoint`(以 ep_id 归属为主, **跨多产品**)
- 两个视图同一笔消费各出现一次但归属维度不同, **不可加和**

### `--mine` vs `PayerID` / `OwnerID`(易混点)

火山计费中心 wire 有 `PayerID` / `OwnerID` 入参字段,**它们跟 `--mine` 完全不是一个概念**,用错会得到误导性结果:

| 维度 | `--mine`(arkcli) | `PayerID` / `OwnerID`(wire) |
|---|---|---|
| 过滤对象 | **IAM 子用户**(SSO 登录的 sub-user,JWT 里的 UserID,如 `77325505`) | **账号 owner**(火山 AccountID,10 位整数,如 `2100566469`) |
| 触发场景 | 多个 IAM 子用户共享同一个账号,只看"我的费用" | **财务托管**(主账号代付多个 owner 子账号,要拆到单个 owner) |
| 实现路径 | 客户端先列我的资源 ID(apikey-SID 或 ep_id),service 层 per-resource fan-out: 每个 ID 一次 wire 调用,服务端按 SplitItemID 原生过滤(命中索引快),结果合并 | wire 原生 `PayerID` / `OwnerID` integer[] 字段, 服务端过滤 |
| 单账户无托管时 | 起过滤作用(只看本人的资源) | **是 no-op**(本来 PayerID 就一个,filter = 全账户) |
| 实测 2026-06 行数 | 0(`menglinhao` 在 6 月没消费,正确) | `PayerID=[2100566469]` Total = 24,302 = 全账号 baseline |

所以"我想看我自己的费用"用 `--mine`,**不要**手撸 `--params '{...,"PayerID":[<我账号 ID>]}'`(会等价于全账号查询)。`PayerID` 仅在主账号代付多个财务托管子账号场景下才有意义。

### `--endpoint` / `--apikey` / `--apikey-sid` 关系

- 三方**互斥**(上游 SplitItemID 是单值字符串,只能选一个维度)
- `--endpoint ep-...` 直接当 SplitItemID 喂 wire,默认 per-endpoint 维度
- `--apikey ark-...` shortcut 层先经 arkbff `OpenGetApiKeySID` 反查成 `apikey-{ts}-{xxx}` SID 再喂 wire,wire 维度切到 `{ark_bd:apikey_id}`(实测 secret `ark-*` 字符串永远返空 List)
- `--apikey-sid apikey-...` 跳过 arkbff,直接当 SID 喂 wire — 用于已经从 `apikey.list` 里抽出 SID 想省一次 RPC 的场景

### `--product` 编码

火山计费中心维护权威清单。**ARK 大模型推理(豆包大模型)的产品编码是 `ark_bd`**(实测确认,`ProductZh` = "字节跳动大模型服务（豆包大模型）")。其他产品编码可以从首条不带 `--product` 的响应里读 `Product` 字段:

```bash
# 第一次先不传 --product, 看本账号实际有哪些产品编码
arkcli billing list --start 2026-05 --page-limit 1 2>&1 | jq '.error // empty'  # 通常会因数据量大失败,换 raw API:
arkcli api billing.list_split_bill_detail --params '{"BillPeriod":"2026-05","Limit":300}' --transform 'Result.List.0.Product'
```

### `--bill-category` 取值

`consume-use` / `consume-new` / `consume-renew` / `consume-formalize` / `consume-modify` / `consume-trial` / `refund-terminate` / `refund-modify` / `transfer-manual` / `transfer-system`

## 返回值

```json
{
  "items": [
    {
      "BillPeriod":         "2026-05",
      "BillingDate":        "2026-05-15",
      "Product":            "ark_bd",
      "ProductZh":          "字节跳动大模型服务（豆包大模型）",
      "SplitItemID":        "ep-20260415xxx-xxxxx",
      "SplitItemName":      "doubao-pro-4k-prod",
      "Currency":           "CNY",
      "OriginalBillAmount": "123.4500",
      "PreferentialAmount": "10.0000",
      "PayableAmount":      "113.4500",
      "Count":              "1234567",
      "Unit":               "千tokens"
    }
  ]
}
```

字段含义:

| 字段 | 含义 |
|------|------|
| `BillPeriod` | 账期 `YYYY-MM` |
| `BillingDate` | 账单日 `YYYY-MM-DD`(`interval=month` 时可能为空) |
| `Product` / `ProductZh` | 产品编码 / 中文名 |
| `SplitItemID` / `SplitItemName` | 拆分项 ID / 名字。行是 endpoint(EP ID + DisplayName)或 apikey(api_key 字符串 + 命名),取决于 SplitItemID 入参类型 |
| `Currency` | 币种,通常 `CNY` |
| `OriginalBillAmount` | **优惠前**金额(原价) |
| `PreferentialAmount` | 优惠 / 资源包抵扣金额 |
| `PayableAmount` | **应付金额**(= Original - Preferential,对账主字段) |
| `Count` / `Unit` | 计量值 / 单位(字符串透传,JSON 数字精度有损) |

> 金额字段都是**字符串**,不是数字 — JSON number 精度不够保留小数位,客户端按字符串透传。前端展示 / 对账加和需要先转 decimal,直接 `parseFloat` 会丢精度。

## 常见错误

| 错误 | 原因 | 处理方式 |
|------|------|---------|
| `--start is required (YYYY-MM)` | 没传 `--start` | 加 `--start 2026-05` |
| `--start ... must be YYYY-MM` | 格式错(如 `2026-5` 单数字月,或带日 `2026-05-01`) | 改成 `2026-05` |
| `--end ... must be >= --start` | 范围终点早于起点 | 调换或修正 |
| `range ... exceeds 24 months` | `--end - --start` 超过 24 个月 | 拆成多次查询 |
| `--day is single-day, incompatible with --start/--end range` | 同时传 `--day` 跟 `--end` | 去掉 `--end`(单月+具体日) 或去掉 `--day`(范围月份) |
| `--day requires --interval=day or --interval=detail` | `--interval=month` 时传 `--day` 无意义 | 加 `--interval=day` 或去掉 `--day` |
| `--day must be in the same month as --start` | `--day` 超出 `--start` 账期范围 | 改 `--day` 到同月内 |
| `--endpoint and --apikey are mutually exclusive` / `--endpoint, --apikey, --apikey-sid are mutually exclusive` | 同时传多个 scope flag | 三选一(SplitItemID 单值) |
| `--split-dim ... incompatible with --endpoint/--apikey` | `--split-dim` 取值跟显式 scope 隐含的维度不一致 | 让 split-dim 跟 scope 一致或不传(如 `--endpoint` 配 `--split-dim endpoint`,或干脆只传 `--endpoint`) |
| `--split-dim ... invalid` | `--split-dim` 取值非 `apikey` / `endpoint` | 改成允许值 |
| `--apikey value too short` | `--apikey` 字符串短到不像合法 secret(< 16 字符) | 改成完整 ark-* 字符串,或如果只有 SID 改用 `--apikey-sid` |
| `arkbff couldn't resolve API key (...) for your account` | arkbff 反查 + 主账号归属校验失败(key 不存在 / 跨账号) | 切到正确 profile 后重试,或直接传 `--apikey-sid` |
| `result exceeds N pages (... records)` | 数据量大 > 默认 cap (30000 条) | 加全局 `--page-limit=N` 显式放宽,或缩小过滤范围(如配 `--endpoint` / `--apikey`) |
| `not logged in` / SSO STS 失效 | 未认证 | `arkcli auth login volc-sso` |

## 注意事项

- **单月查询是 wire 约束, range 客户端模拟**: 计费中心 `ListSplitBillDetail` 上游只接受单 `BillPeriod`, `--start/--end` 由 service 层展开月份列表 + 并行 N 次调用 + 客户端合并 — 12 个月范围约 ~400ms (并行) 而不是 ~5s (串行)
- **数据量可能很大**: 实测单账期 ARK 账单经常 30000+ 行(per-product × per-resource 行级拆分);默认 `--page-limit` 兜底 30000 条触发即报错。建议:加 `--endpoint` / `--apikey` 过滤,或显式放宽 `--page-limit`
- **不暴露 Limit / Offset**: service 层强制全量分页, output 永远是完整范围。给账单求和的常见场景设计 — 翻页对账容易少算
- **数据有出账延迟**: 计费中心 T+1 出账,当月当日数据可能还没汇总完整。月度对账建议在月初拉**前一个完整账期**
- **金额是 CNY 字符串**: 加和 / 比较请用 decimal 库,`parseFloat` 会丢小数

## 参考

- [arkcli-billing](../SKILL.md) -- billing skill 概览
- [usage stats](../../arkcli-usage/references/arkcli-usage-stats.md) -- 推理用量(Token / 请求数)查询,跟 billing 配套使用
- [arkcli-shared](../../arkcli-shared/SKILL.md) -- 认证和全局参数
