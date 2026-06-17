# plans personal rotate-apikey

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **⚠️ 写操作 / 毁坏性：** 调用成功后**原 APIKey 立即失效**，需要立即替换接入该 API Key 的模型 / Harness 配置以避免服务中断。默认走 [Y/N] 二次确认；脚本场景可加 `--yes` 跳过。

轮换 Agent Plan **个人版** 的专属 APIKey。一次调用走 `ListApiKeys` (Filter.Scene=`RealAgentPlanPersonal`) → `UpdateApiKey {Id, Regen:true}` → `GetRawApiKey` 三步，把新明文返回。

## 命令

```bash
# 交互式：会先弹 [Y/N] 确认
arkcli plans personal rotate-apikey

# 非交互（脚本 / CI）：跳过确认
arkcli plans personal rotate-apikey --yes
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--yes` | 否 | bool | 跳过 [Y/N] 二次确认（脚本/CI 友好） |

## 返回值

成功：
```json
{
  "status": "success",
  "key_id": 114788,
  "api_key": "sk-rotated-new-secret"
}
```

`api_key` 是**新明文**，跟 frontend Modal 弹出的 "新 APIKey" 同一份。务必立即保存。

## 失败语义

| 场景 | exit code | 输出 |
|---|---|---|
| 用户 [N] / EOF / 空回车取消 | 1（ExitGeneral） | `{"ok":false,"error":{"type":"cancelled","message":"operation cancelled"}}` |
| 当前账号没买 Agent Plan 个人版 | 2（ExitValidation） | `{"ok":false,"error":{"type":"validation","message":"no Agent Plan 个人版 API Key found","hint":"Subscribe Agent Plan 个人版 first via \`arkcli plans buy --plan agent-plan ...\`"}}` |
| transport / 服务端错 | 1（ExitGeneral） | 普通 error 上抛 |

## 二次确认文案

CLI 输出（stderr）：
```
确认更新专属 APIKey?
更新后原 APIKey 将立即失效，请及时替换接入该 API Key 的模型及 Harness 配置，以避免服务中断。
继续? [y/N]:
```

输入 `y` / `Y` 才放行；其它（含空回车 / EOF / 任何字符）都按取消。

## 注意事项

- 这条命令仅影响 Agent Plan **个人版** scoped APIKey（`Filter.Scene=RealAgentPlanPersonal`），**不会影响**普通 Platform / Coding 池的 APIKey
- 服务端按账号粒度有最多 1 把 personal-scope key；多了走第一个
- 跟 `arkcli auth` 持久化的 SSO STS / IDToken 无关 —— 那是 CLI 自己的鉴权链路，APIKey 是给数据面 SDK 用的
- 没订阅 Agent Plan 个人版时**不会**报 transport 错，会走 `ErrNoAgentPlanPersonalKey` 哨兵 + ExitValidation envelope，方便脚本判定

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [`plans get`](arkcli-plans-get.md) -- 先确认是否有 Agent Plan 个人版订阅
- [`plans team rotate-apikey`](arkcli-plans-team-rotate-apikey.md) -- 团队版席位 APIKey 轮换
- [arkcli-shared](../../arkcli-shared/SKILL.md)
