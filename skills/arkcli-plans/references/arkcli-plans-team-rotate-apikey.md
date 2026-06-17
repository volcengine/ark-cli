# plans team rotate-apikey

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **⚠️ 写操作 / 毁坏性：** 调用成功后，所选席位的**原 APIKey 立即失效**，必须立即替换接入该 APIKey 的模型 / Harness 配置。默认走 [Y/N] 二次确认；`--yes` 跳过。

轮换 Agent Plan / Coding Plan **企业版** 席位的 APIKey。一次调用 `RegenerateEnterpriseCodingPlanApikey`，可批量；服务端 per-item 状态。

支持两种使用场景：

- **self-rotate**：不传 `--seat-ids` 时，自动反查 "我的 agent-plan-team 席位"（`GetSeatInfo` Scene=`agent_plan_enterprise`），轮换该席位 APIKey
- **admin batch**：显式传 `--seat-ids`，按 SeatID 批量轮换，服务端校验当前身份是否有管理权限

> 用户 spec 注释里写的是 "Agent Plan 企业版"，self-rotate 路径只解析 agent-plan-team 自己的席位。Coding Plan 企业版用户自己轮换 APIKey 当前要走 admin 路径（显式传 `--seat-ids`）。

## 命令

```bash
# 自己轮换 agent-plan-team 席位 APIKey（交互式）
arkcli plans team rotate-apikey

# 同上但跳过确认（脚本场景）
arkcli plans team rotate-apikey --yes

# 管理员批量轮换指定席位
arkcli plans team rotate-apikey --seat-ids seat-001,seat-002

# 管理员批量 + 跳过确认
arkcli plans team rotate-apikey --seat-ids seat-001,seat-002,seat-003 --yes
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--seat-ids` | 否 | string | 逗号分隔 SeatID。留空=self-rotate；非空=admin batch |
| `--yes` | 否 | bool | 跳过 [Y/N] 确认 |

## 返回值

```json
{
  "scope": "self",
  "success_count": 1,
  "failed_count": 0,
  "success": [
    {
      "seat_id": "seat-self-42",
      "api_key": "sk-rotated-new-secret",
      "api_key_sid": "apikey-..."
    }
  ],
  "failed": []
}
```

| 字段 | 说明 |
|------|------|
| `scope` | `self` / `admin` —— 提示该次调用走的是反查路径还是显式批量路径 |
| `success[].api_key` | 新明文，立即保存 |
| `success[].api_key_sid` | APIKey 稳定 ID |
| `failed[].reason` | 服务端透传 |

## 二次确认文案

self-rotate：
```
确认更新当前身份的 Agent Plan 企业版席位 APIKey?
更新后原 APIKey 将立即失效，请及时替换接入该 API Key 的模型及 Harness 配置，以避免服务中断。
继续? [y/N]:
```

admin batch（带 N 个席位）：
```
确认批量更新 N 个席位的 APIKey?
更新后原 APIKey 将立即失效，请及时替换接入该 API Key 的模型及 Harness 配置，以避免服务中断。
继续? [y/N]:
```

## 失败语义

**部分失败**是合法终态（跟 `seat-assign` 同套路）：
- stdout 始终是完整 result
- `failed_count > 0` → stderr 追加 `rotate_apikey_partial` envelope + exit code 5

| 错误 | 原因 | 处理 |
|------|------|------|
| `cancelled` envelope（exit 1） | 用户输 N / 空回车 / EOF | 重新跑加 `--yes` 或确认轮换 |
| `no Agent Plan 企业版 seat bound to current identity` (validation) | self-rotate 但当前身份没企业版席位 | 让管理员买套餐 + 用 [`seat-assign`](arkcli-plans-team-seat-assign.md) 给你绑席位；或者你自己有管理权限就走 `--seat-ids` admin 路径 |
| `--seat-ids must contain at least one non-empty seat ID` | `--seat-ids ""` 或全是逗号 | 检查输入 |
| `--seat-ids exceeds single-call limit 1000` | 超上限 | 拆批调用 |
| `RotateTeamSeatAPIKeys: duplicate SeatID %q` | 重复 ID | 客户端先拦 |
| `failed_count > 0` + `AccessDenied` | 当前身份对该 SeatID 没管理权限 | 检查 SeatID 归属；普通成员只能 self-rotate |

## 注意事项

- self-rotate 当前**只解析 agent-plan-team 自己的席位**（Scene=agent_plan_enterprise）；如果你是 coding-plan-team 用户想自己轮换，先 `plans team seat-list --plan coding-plan-team --user-name <你>` 拿 SeatID，再用 admin 路径
- 服务端按 SeatID 反推套餐线，所以管理员**可以同时混合 agent / coding plan 的 SeatID** 在一个 `--seat-ids` 里
- 跟 `plans personal rotate-apikey` 区别：那个是**个人订阅**的 APIKey；这个是**企业版席位**的 APIKey，互不影响
- 可能撞 `failed_count > 0` 但 `success_count > 0` 的部分成功 —— stdout 仍包含成功项的新 APIKey，**不要因为 exit code 非零就丢掉 success 那部分**

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [`plans personal rotate-apikey`](arkcli-plans-personal-rotate-apikey.md) -- 个人版 APIKey 轮换
- [`plans team seat-list`](arkcli-plans-team-seat-list.md) -- 列出可轮换的席位
- [`plans team seat-assign`](arkcli-plans-team-seat-assign.md) -- 绑定席位
- [arkcli-shared](../../arkcli-shared/SKILL.md)
