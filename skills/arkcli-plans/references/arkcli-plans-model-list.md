# plans model-list

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

列出指定套餐支持的全部模型，并标出**当前选中的 ark-latest-model**（被 alias 到 `<latest-name>` 的真实模型 ID）。读操作。

## 命令

```bash
# 默认看 agent-plan-team 的模型清单
arkcli plans model-list

# 看 Agent Plan 个人版模型
arkcli plans model-list --plan agent-plan

# 看 Coding Plan 团队版模型
arkcli plans model-list --plan coding-plan-team
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--plan` | 否 | string | `agent-plan` / `coding-plan` / `agent-plan-team` / `coding-plan-team`，默认 `agent-plan-team` |

## 返回值

```json
{
  "plan": "agent-plan-team",
  "ark_latest_model_id": "doubao-seed-2-0-pro-260215",
  "models": [
    {
      "model_id": "doubao-seed-2-0-pro-260215",
      "is_ark_latest": true
    },
    {
      "model_id": "doubao-pro-32k-260215",
      "is_ark_latest": false
    }
  ]
}
```

| 字段 | 说明 |
|------|------|
| `plan` | 入参回显 |
| `ark_latest_model_id` | 当前 alias 到 `ark-latest-model` 的具体模型 ID；个人版按订阅服务端 `ListAgentPlanLatestModel` / `ListArkCodeLatestModel` 的 Enabled 字段判断；团队版从用户席位 `SeatInfo.ExtraConfig.ArkCodeLatestMappingModelID` 取 |
| `models[].is_ark_latest` | 该模型是否就是当前 ark-latest 指向 |

## 注意事项

- 个人版与团队版的 ark-latest 数据源不同：
  - 个人版：服务端按订阅维度返回模型清单 + Enabled flag
  - 团队版：从用户席位 `ExtraConfig.ArkCodeLatestMappingModelID` 字段读
- 没订阅 / 没席位时返回空数组 + 空 `ark_latest_model_id`，**不会报错**
- 数据是订阅期内允许调用的模型，不等同于全局可调用模型；试用模型用 [`../arkcli-models/`](../../arkcli-models/SKILL.md)

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [`plans get`](arkcli-plans-get.md) -- 先看自己持有哪些套餐
- [arkcli-models](../../arkcli-models/SKILL.md) -- 不限于套餐的模型查询
- [arkcli-shared](../../arkcli-shared/SKILL.md)
