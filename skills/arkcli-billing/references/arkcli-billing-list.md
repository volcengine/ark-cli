# billing list

> **前置**: 先读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) (认证 / 全局参数)。

> **`billing list` vs `usage stats`**: stats 出"用了多少 token"(近实时,ARK BFF 聚合);billing 出"花了多少钱"(T+1 出账,火山计费中心)。

查询指定账期(YYYY-MM)或月范围内的拆分账单明细。**默认全量分页**,撞 service 内部 cap (~3000 行 / 1 MB JSON) **软截断**(不 hard fail): 已抓数据正常返 + `is_truncated=true` + `partial_failures` 标注哪些 fan-out 截断 + stderr 强警告。`--limit/--offset` 切显式分页模式 (跨 fan-out 时 per-fan-out 各 N 行)。

**核心纪律**: agent 必须先看 `is_truncated` 和 `partial_failures` 头判断截断状况, **`is_truncated=true` 时不要 sum items 当总额** — 看 `partial_failures[*].total` 知真实规模, 然后决定下一步: 想全量明细 → `--output FILE` 落盘 / 只要总金额 → `--split-dim apikey|endpoint` 聚合 / sample 够就用现状 / 缩到单资源 → `--endpoint`/`--apikey`。

> **`--start` 单独用 = 单月查询**(不是开放区间到今天)。`billing list` 是月单位,wire 单 `BillPeriod` 字段 — 要查多月范围必须 `--end`(例: `--start 2026-03 --end 2026-05`,客户端 fan-out 模拟)。跟 `usage stats` 的日级 `--start --end` 不同。

## 命令

```bash
# 单月汇总(最常用) — 默认: ARK 推理 7 个产品(arkProducts,不含订阅) + profile.project
# 自动注入(若 profile.project 是具体 id, 如 "auto-test");profile.project="default" /
# 账号全部资源哨兵 / 空 时跳过注入 = 真账号全量。stderr 出软提示告知。
arkcli billing list --start 2026-05

# 多月范围(闭区间, 最多 24 个月)
arkcli billing list --start 2026-03 --end 2026-05

# 按天 / 按每条结算明细
arkcli billing list --start 2026-05 --interval day --day 2026-05-15
arkcli billing list --start 2026-05 --interval detail

# 单值 scope (三选一互斥)
arkcli billing list --start 2026-05 --endpoint ep-...
arkcli billing list --start 2026-05 --apikey ark-...        # 自动经服务端反查 SID
arkcli billing list --start 2026-05 --apikey-sid apikey-... # advanced, 跳过反查

# 账号维度切片
arkcli billing list --start 2026-05 --split-dim endpoint    # 每个 EP 各花了多少
arkcli billing list --start 2026-05 --split-dim apikey      # 每把 key 各花了多少

# 我自己的费用 (SSO 子用户视角)
arkcli billing list --start 2026-05 --mine                  # 默认 --mine-by=endpoint (对齐 usage stats)

# 其他过滤
arkcli billing list --start 2026-05 --product ark_subscription  # 看订阅类账单
arkcli billing list --start 2026-05 --bill-category consume-use --billing-mode 2
arkcli billing list --start 2026-05 --ignore-zero
```

## 参数

| 参数 | 必填 | 说明 |
|------|---|------|
| `--start` | **是** | 单个账期 `YYYY-MM` (距今 24 个月内) |
| `--end` | 否 | 范围终点 `YYYY-MM` (闭区间,省略 = 单月,最多 24 个月) |
| `--day` | 否 | 账单日 `YYYY-MM-DD`,**仅 `--interval=day\|detail` 生效**,不兼容 `--end` |
| `--interval` | 否 | `month` (默认) / `day` / `detail` |
| `--endpoint` / `--apikey` / `--apikey-sid` | 否 | 单值 scope,三方互斥 (SplitItemID 单值)。`--apikey ark-*` 自动经服务端反查成 SID;`--apikey-sid` 跳过反查 |
| `--split-dim` | 否 | 账号维度切片: `apikey` / `endpoint`。**显式 scope 已隐含维度,split-dim 必须一致或不传** |
| `--product` | 否 | 火山产品编码 (可重复)。**ARK 大模型推理 = `ark_bd`**;订阅类 = `ark_subscription` |
| `--project` | 否 | 项目 ID (可重复)。**不传时自动用 profile 的 ProjectName 当默认值**;`--project=` (空)= 账号全量 |
| `--instance-no` | 否 | 计费实例 ID |
| `--bill-category` | 否 | `consume-use` / `consume-new` / `consume-renew` / `consume-formalize` / `consume-modify` / `consume-trial` / `refund-terminate` / `refund-modify` / `transfer-manual` / `transfer-system` |
| `--billing-mode` | 否 | `1` 包年包月 / `2` 按量 / `3` 合同 / `4` 履约 |
| `--ignore-zero` | 否 | 跳过折后价 = 0 的行 |
| `--mine` | 否 | 仅看当前 SSO 子用户名下资源的账单。维度由 `--mine-by` 决定。**跟 `--endpoint` / `--apikey` / `--apikey-sid` / `--split-dim` 互斥**。订阅类不含 (没 IAM 归属) |
| `--mine-by` | 否 | `endpoint` (默认,**infra ownership** — 我部署的 EP 上的账单,含他人调;跟 usage stats 默认对齐) / `apikey` (**cost causation** — 我创建的 key 产生的账单)。两视图不可加和,前者按 infra 归集后者按 payer 归集 |
| `--limit` / `--offset` | 否 | 分页模式 (Limit 1-300,wire 硬上限),单页直返。在 fan-out 时是 **per-fan-out 各 N 行** (跨月 `--end` / `--mine` 多资源),`partial_failures[]` 标 `windowed sample` |
| `--output` | 否 | 写完整 JSON (items + summary + partial_failures) 到 FILE,stdout 只剩 metadata 不爆 context。自动放宽内部 cap 到 300k 行 |

## Scope 视图能力

不同 scope 覆盖的账单子集不一样。**`ark_subscription` (Agent Plan / Coding Plan) 只能在账号维度看,任何 IAM/EP/apikey 维度都看不到**(账号级归属,没 IAM owner)。

| 视图 | ARK 推理 (7 product) | 订阅类 (`ark_subscription`) |
|---|---|---|
| 裸跑 (默认) | ✅ | ❌ (默认 Product 锁过滤) |
| `--product ark_subscription` | ❌ | ✅ |
| `--product ark_bd --product ark_subscription` | ✅ | ✅ |
| `--endpoint` / `--apikey` / `--apikey-sid` / `--split-dim` / `--mine` | 各自 scope | ❌ |

**`--mine-by=endpoint` 跨产品覆盖**: SplitItemID `ep-...` 是跨产品 unique 主键,`+deploy` 出来的 endpoint 账单会落在不同 Product 字段下 (LLM → `ark_bd`,图片 → `Doubao-image-generation`,视频 → `ark_bd` 等);service 层不锁 Product 让 wire 按 SplitItemID 自然跨产品命中。所以 `--mine --mine-by=endpoint` 会把图片 / 视频 EP 的账单也包含。两个 mine 视图同一笔消费各出现一次,**不可加和**。

**`--mine` ≠ `PayerID`**:
- `--mine` 按 **IAM 子用户** UserID 过滤 (客户端先列我的资源 ID,service 层 per-resource fan-out)
- wire 的 `PayerID` / `OwnerID` 是 **账号 owner ID** (10 位整数),用于财务托管场景
- 单账户场景手撸 `--params '{"PayerID":[<我账号 ID>]}'` 是 no-op (= 全账号查询);要看自己用 `--mine`

## 返回值

```json
{
  "items": [
    {
      "BillPeriod":         "2026-05",
      "BillingDate":        "2026-05-15",
      "Product":            "ark_bd",
      "ProductZh":          "字节跳动大模型服务（豆包大模型）",
      "SplitItemID":        "ep-...",
      "SplitItemName":      "doubao-pro-4k-prod",
      "Currency":           "CNY",
      "OriginalBillAmount": "123.4500",
      "PreferentialAmount": "10.0000",
      "PayableAmount":      "113.4500",
      "Count":              "1234567",
      "Unit":               "千tokens"
    }
  ],
  "total_records": 87881,
  "returned": 3000,
  "is_truncated": true,
  "partial_failures": [
    {
      "period": "2026-05",
      "total": 87881,
      "returned": 3000,
      "reason": "exceeds 10 pages cap (3000 rows scanned, total 87881 rows). Options: --limit/--offset to sample / page through; --split-dim apikey|endpoint to aggregate server-side; --endpoint/--apikey to narrow scope; --page-limit=N to raise the cap"
    }
  ]
}
```

| 字段 | 含义 |
|------|------|
| `BillPeriod` / `BillingDate` | 账期 / 账单日 (interval=month 时 BillingDate 可能为空) |
| `Product` / `ProductZh` | 产品编码 / 中文名 |
| `SplitItemID` / `SplitItemName` | 拆分项 ID / 名字 (endpoint 或 apikey,取决于 scope) |
| `Currency` | 通常 `CNY` |
| `OriginalBillAmount` | 优惠前金额 |
| `PreferentialAmount` | 优惠 / 资源包抵扣金额 |
| `PayableAmount` | **应付金额** (= Original - Preferential,对账主字段) |
| `Count` / `Unit` | 计量值 / 单位 |
| `total_records` / `returned` / `is_truncated` | summary 头: 服务端真实总数 / 本次返回行数 / 是否截断 |
| `partial_failures` | 软截断 / sample 标记数组 (撞 cap 或 windowed); 每项含 `period` / `resource` / `total` / `returned` / `reason`,**`is_truncated=true` 时必查** |

> 金额字段是字符串,加和 / 比较用 decimal 库,不要 `parseFloat` (JSON 数字精度有损)。

## 软截断 (cap / windowed) 决策

撞 cap 或显式 `--limit/--offset` 时 service **不 hard fail**,返已抓数据 + summary 头 + stderr 强警告。agent 看 `partial_failures` 选下一步:

| `partial_failures[].reason` | 含义 | 推荐处理 |
|---|---|---|
| `exceeds N pages cap (...)` | 默认全量模式撞 service 兜底 cap | 想全量明细 → `--output FILE` (自动放宽 cap 到 300k 行,落盘);只要总金额 → `--split-dim apikey\|endpoint` 服务端聚合;sample 够 → 用现状;缩到单资源 → `--endpoint`/`--apikey` |
| `windowed sample (limit=L, offset=M); returned X of Y rows` | 用户主动开 `--limit/--offset` 拿 sample | 用户已知是 sample;循环 `offset += limit` 客户端拼全量,或换 `--output FILE` 全量落盘 |

跨 fan-out (`--end` 跨月 / `--mine` 多资源) 的 `--limit/--offset` 是 **per-fan-out** 各 N 行 (returned ≤ limit × fanout_count); 每个 fan-out 在 `partial_failures` 各占一项,标 `period` / `resource` 字段。

## --output FILE 模式

落盘大数据避免 stdout 灌爆 agent context。跟 `train finetune logs/metrics/trajectory` 的 `--output` 一致:

```bash
arkcli billing list --start 2026-05 --output bills.json
#   ↳ stdout: {total_records: 87881, returned: 87881, is_truncated: false, output_file: "bills.json"}
#   ↳ bills.json: 完整 JSON (含 87881 行 items + summary + partial_failures)
```

- service 自动把内部 cap 抬高到 1000 pages (300k 行 / ~100 MB 内存),覆盖绝大多数大账户全量
- stdout 只剩 metadata (items=null + output_file 路径) — agent context 安全
- 用户显式 `--page-limit=N` 仍尊重 (用户传的优先)
- 仍超时 partial_failures 会标,再放宽 `--page-limit`

## 客户端循环翻页范式

需要拼全量数据时 (撞 cap 或 windowed sample 都适用):

**单月分页拉全量**:
```bash
total=$(arkcli billing list --start 2026-05 --limit 1 2>/dev/null | jq '.total_records')
offset=0; all="[]"
while [ "$offset" -lt "$total" ]; do
  page=$(arkcli billing list --start 2026-05 --limit 300 --offset "$offset" 2>/dev/null | jq '.items')
  all=$(jq -n --argjson a "$all" --argjson b "$page" '$a + $b')
  offset=$((offset + 300))
done
echo "$all" | jq 'length'   # 应等于 total_records
```

**跨月嵌套翻页拉每月全量** (外层 fan-out + 内层 +offset):
```bash
for m in 2026-03 2026-04 2026-05; do
  total=$(arkcli billing list --start "$m" --limit 1 2>/dev/null | jq '.total_records')
  offset=0; items="[]"
  while [ "$offset" -lt "$total" ]; do
    page=$(arkcli billing list --start "$m" --limit 300 --offset "$offset" 2>/dev/null | jq '.items')
    items=$(jq -n --argjson a "$items" --argjson b "$page" '$a + $b')
    offset=$((offset + 300))
  done
  jq -n --arg m "$m" --argjson t "$total" --argjson it "$items" '{month:$m, total:$t, items:$it}'
done | jq -s '.'
```

客户端翻页**不受** `--page-limit` cap 限制 (cap 只在默认全量模式触发)。

## 常见错误

| 错误 | 处理 |
|------|------|
| `--start is required` / `must be YYYY-MM` | 加或修正 `--start 2026-05` (距今 24 个月内) |
| `--end ... must be >= --start` / `range exceeds 24 months` | 调整范围 |
| `--day ... incompatible with --start/--end range` / `requires --interval=day or detail` | 单月 + 同月 day,或去掉 `--day` |
| `--endpoint, --apikey, --apikey-sid are mutually exclusive` | 三选一 (SplitItemID 单值) |
| `--split-dim ... incompatible with --endpoint/--apikey` | 让 split-dim 跟显式 scope 一致或不传 |
| `--apikey value too short` (<16) | 改完整 ark-*,或如果只有 SID 改 `--apikey-sid` |
| `couldn't resolve API key` | 切到正确 profile,或直接传 `--apikey-sid` |
| `not logged in` / SSO 失效 | `arkcli auth login volc-sso` |

> **撞 cap 不在错误码表** — 软截断走正常返回路径 (ok=true + is_truncated=true + partial_failures + stderr WARN)。详见上面"软截断决策"段。

## 注意事项

- **T+1 出账**: 当月当日数据可能不全,月度对账建议月初拉**前一个完整账期**
- **数据量大**: 单账期 ARK 账单常上万行,默认 cap ~3000 条 (~1 MB JSON);首选 `--split-dim` 服务端聚合,sample 用 `--limit`,要全量明细放宽 `--page-limit`
- **金额是 CNY 字符串**: 用 decimal 库,不要 `parseFloat`
- **默认锁 ARK 7 个产品**: 防 EIP / TOS / VDB 等非 ARK 产品脏返回,**不含 `ark_subscription`** — 看订阅类显式 `--product ark_subscription`

## 参考

- [arkcli-billing](../SKILL.md) — billing skill 概览
- [usage stats](../../arkcli-usage/references/arkcli-usage-stats.md) — 推理用量 (Token / 请求数)
- [arkcli-shared](../../arkcli-shared/SKILL.md) — 认证和全局参数
