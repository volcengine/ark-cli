# arkcli infer endpoint list

List infer endpoint endpoints

## 全量盘点与导出

- “我的”保留对应产品支持的本人范围；“账号所有在用”才按账号范围列 `--status Running --page-all --page-size 100`。不要因为本人范围空而自动改成全账号。
- JSON 数组是 `Items`，不是 `items`。保存完整 stdout 后解析，stderr 单独保留；检查分页上限/告警和实际返回数量，不能把第一页或 `head` 截取结果当总清单。`--page-all` 可以与 `--page-size` 同用。
- 交付 CSV/表格至少包含精确 ID、Name、实际绑定模型、Status、CreateTime、Description；用户要求 QPM/TPM 等列表未返回的详情时，才对已筛选的目标补 `get`。缺失字段与真实 0 分开表示。
- 批量详情默认串行并遵守接口 QPS；需要并行时先确认限流，不照抄固定 8/10 并发。每个目标分别记录成功/失败，不将 `2>&1` 混入 JSON，不因部分失败输出“全量导出成功”。
- 时间保留原始值，并明确展示时区；已带 offset 的时间只转换一次，不能再手动加八小时。汇总报告写明身份/范围、筛选条件、匹配数量、失败项及是否完整。
- ID 与文件名分别处理：不要把外部返回值直接拼进 shell 代码或输出路径；解析完整结构化结果后，以引用参数传给 `get`。

### 统计、逐个补详情与 CSV 模板

下面假设用户已要求“账号所有在用 EP”；如要求“我的”，第一条必须保留 `--mine`，其余步骤不改变身份和范围。在允许脚本的宿主中执行；只允许单条 CLI 的宿主按相同步骤逐条调用工具。目录由 `mktemp` 创建，不覆盖已有交付。

```bash
ep_report_dir=$(mktemp -d)
arkcli infer endpoint list --status Running --page-all --page-size 100 --page-delay 800 --format json > "$ep_report_dir/list.json" 2> "$ep_report_dir/list.stderr"
# 仅在上述命令退出码为 0 后继续；先读 stderr 的分页告警。
jq -e '(.Items | type == "array") and (.TotalCount | type == "number")' "$ep_report_dir/list.json"
jq '{server_total: .TotalCount, captured: (.Items | length),
     by_status: ([.Items[] | (.Status // "未返回")] | group_by(.) | map({(.[0]): length}) | add // {})}' "$ep_report_dir/list.json"
```

`captured` 与 `server_total` 不一致时先核对分页/服务端过滤及查询过程中资源变化；不能宣称完整。截断先读宿主保存的完整文件；确无完整捕获才同范围落盘补查一次。以下只对列表缺少所需详情的对象补查，已有完整详情时直接复用。不能把 RateLimit 缺失当成 0，也不能假定所有版本的 list 都不返回限流字段。

```bash
jq -er '.Items[] | select(.RateLimit.Rpm == null or .RateLimit.Tpm == null) | .Id | select(type == "string" and length > 0)' "$ep_report_dir/list.json" > "$ep_report_dir/detail-ids.txt"
# 空待查集合正常（jq -e 此时可返回 4）；JSON 解析失败则停止，不执行循环。
ep_seq=0
while IFS= read -r ep_id; do
  ep_seq=$((ep_seq + 1))
  if arkcli infer endpoint get "$ep_id" --format json > "$ep_report_dir/$ep_seq.json" 2> "$ep_report_dir/$ep_seq.stderr"; then
    jq -n --arg id "$ep_id" --arg file "$ep_seq.json" '{id:$id,ok:true,file:$file}'
  else
    jq -n --arg id "$ep_id" --arg file "$ep_seq.stderr" '{id:$id,ok:false,error_file:$file}'
  fi
  sleep 1
done < "$ep_report_dir/detail-ids.txt" > "$ep_report_dir/details-manifest.jsonl"
```

每个成功响应先校验为对应 ID 的 Endpoint 对象，再按 `Id` 合并回列表；失败行保留并标注“详情未获取”，不得把错误对象或 stderr 合并进 `Items`。批量脚本的退出码不能代替逐项结果。

生成展示字段时用标准时区转换，保留 UTC 原值；以下以 CN 用户的北京时间为例，其他用户使用其明确时区：

```bash
node - "$ep_report_dir/list.json" <<'JS' > "$ep_report_dir/display.json"
const fs = require('node:fs');
const data = JSON.parse(fs.readFileSync(process.argv[2], 'utf8'));
const fmt = new Intl.DateTimeFormat('sv-SE', {timeZone:'Asia/Shanghai',year:'numeric',month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit',second:'2-digit',hourCycle:'h23'});
for (const item of data.Items) {
  const time = Date.parse(item.CreateTime);
  item.CreateTimeBeijing = Number.isFinite(time) ? fmt.format(new Date(time)) + ' +08:00' : '时间缺失或无效';
}
process.stdout.write(JSON.stringify(data));
JS
jq -r '
  def binding:
    if .EndpointModelType == "CustomModel" then .ModelReference.CustomModelId // "待解析"
    elif .EndpointModelType == "FoundationModel" then
      .ModelReference.FoundationModel as $m |
      if ($m.Name // "") == "" then "待解析"
      else [$m.Name, $m.ModelVersion | select(. != null and . != "")] | join("-") end
    else "待 resources resolve 解析" end;
  ["序号","EP-ID","Name","绑定模型","Status","CreateTime 原值","创建时间（北京时间）","RPM","TPM","Description"],
  (.Items | to_entries[] | .key as $i | .value |
   [$i+1,.Id,.Name,binding,.Status,.CreateTime,.CreateTimeBeijing,
    (.RateLimit.Rpm // "未返回"),(.RateLimit.Tpm // "未返回"),(.Description // "")]) | @csv
' "$ep_report_dir/display.json" > "$ep_report_dir/endpoints.csv"
```

当前字段叫 `Rpm` / `Tpm`，不是臆造的 `Qpm`；若用户称 QPM，说明对应请求/分钟口径。CustomModel 即便同时带 FoundationModel lineage 也只导出 `cm-*` 真实绑定。CSV 用 `@csv` 正确处理逗号、引号、换行；未知绑定单列待解析项，不能当成已完成。最终附文件路径、查询范围/条件、完整性、详情失败清单。

## Usage

```bash
arkcli infer endpoint list [flags]
```

## Flags

| Flag | Type | Description | Required |
|------|------|-------------|----------|
| `--format` | string | Output format: table \| json | No |
| `--model` | string | Filter by model: a custom model ID (`cm-...`) or a foundation model name (`doubao-seedream-5-0-pro`, or the dotted DisplayName form `doubao-seedream-5.0-pro`) | No |
| `--status` | string | Filter by endpoint status | No |
| `--mine` | bool | 仅列出当前 SSO 子账号创建的接入点（服务端 `sys:ark:createdBy` Tag 过滤） | No |
| `--page-number` | int | Page number (>=1) | No |
| `--page-size` | int | Page size | No |
| `-h`, `--help` | | help for list | No |

## `--model` 语义

`--model` 接受两类取值，由 CLI 分流：

- `cm-` 前缀 → 按**自定义模型 ID** 过滤（`Filter.CustomModelIds`，大小写敏感）。
- 其它 → 按**基础模型名**过滤：先归一化到 connector Name，再走
  `Filter.FoundationModelNames`（服务端精确匹配）。

归一化只做两件事，**不做机械字符替换**：先按 Name 全串匹配，再按 DisplayName
大小写不敏感全串匹配。展示名与 Name 不总是同一变形 —— 例如模型
`doubao-seedream-5-0` 的 DisplayName 是 `Doubao-Seedream-5.0-lite`（展示名带
`lite`，connector Name 不带），所以把点号换成连字符会拼出一个不存在的名字。

名字既不是 `cm-` ID、也匹配不上任何基础模型名时，命令**报错退出（非 0）**并给出
提示，**不会**返回空表。不要因为拿到空结果就断言"这个账号没有该模型的接入点"。

```bash
arkcli infer endpoint list --model doubao-seedream-5-0-pro --page-all --page-size 100 --format json
```

## "我的接入点" 语义（覆盖 shared default）

本命令**已内建 `--mine`**，按 [`../../arkcli-auth/references/identity-resolution.md`](../../arkcli-auth/references/identity-resolution.md) 的"决策顺序"必须**优先**走它，
不要再退回 `whoami + jq` 客户端过滤。

```bash
arkcli infer endpoint list --mine --page-all --page-size 100 --page-delay 800 --format json
```

行为：

- `--mine` 在 shortcut 层把当前 `cfg.UserID` / `cfg.UserName` 拼成 `IAMUser/<UserID>/<UserName>`，
  作为 `TagFilters.{Key: "sys:ark:createdBy", Values: [...]}` 下发到 `ListEndpoints`。
- 与 `--status` / `--model` / `--page-*` 可自由叠加（例如 `--mine --status Running`）。
- 仅在 **SSO 子账号**登录态下可用。其它态会**直接报错**，不静默退化：
  - 非 SSO（AK/SK / APIKey）：`--mine requires an SSO sub-user login`，运行 `arkcli auth login volc-sso` 重新建立 Volc 身份
  - SSO root：`--mine is not supported for root logins`，引导改用 sub-user 登录
  - SSO 但 UserID/UserName claim 缺失：引导重登刷新身份

## Global Flags

| Flag | Type | Description |
|------|------|-------------|
| `--debug` | | Print request and response debug details to stderr |
| `--page-all` | | Automatically fetch all pages when supported |
| `--page-delay` | int | Delay in milliseconds between pages (default 200) |
| `--page-limit` | int | Maximum pages to fetch with --page-all (default 10) |
| `--profile` | string | Active config profile |
| `--transform` | string | Transform output with a GJSON-style path expression |
