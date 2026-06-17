---
name: arkcli-onboard
version: 1.0.0
description: "arkcli 接入向导(workflow):把某个模型接入到自己的应用/服务的端到端引导 —— 从'我想用某模型'到拿到可调用的 Endpoint(+ 可选示例代码)。当用户说'我想在我的 app/服务里用豆包/某模型''怎么把方舟模型接进来''帮我接入 XX 模型''想正式用上某模型'这类**不含 deploy/部署关键词、但本质是正式接入**的意图时触发。已明确说'部署/创建 endpoint'的直接走 arkcli-deploy;只想试效果/问一句的走 +chat/+gen。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli --help"
---

# arkcli 接入向导（workflow）

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)（认证闸门、命令选择顺序、写操作安全边界）。**

> **本 skill 只负责编排顺序与分支。** 每一步的**私有命令细节**（如 `+deploy` 的 `--model` / `--name` / JSON 字段，`models search` 的过滤）都在对应能力 skill 里，**本文不重复**。本文只允许出现跨命令通用的安全护栏（如 `--dry-run` 预演）；一旦出现某个命令的私有 flag 说明，就是越界。

## 它解决什么

用户表达的是**意图级**目标（"我想在服务里用豆包"），而不是命令级目标（"部署一个 endpoint"）。这类口语通常不会命中 [`arkcli-deploy`](../arkcli-deploy/SKILL.md) 的关键词，但本质就是"正式接入 = 需要一个可复用的 Endpoint"。本 skill 把这个意图固化成一条有序、有分支、可回归的链路，逐步 delegate 给 owning skill。

## 触发 vs 不触发

- ✅ "我想在我的 app/服务里用豆包" / "怎么把方舟模型接到我的服务" / "帮我接入 seedream" / "想正式用上某模型"
- ❌ 用户已说"**部署 / 创建 endpoint / deploy**" → 直接走 [`arkcli-deploy`](../arkcli-deploy/SKILL.md)（已在 deploy 意图里，不必经本向导）
- ❌ 只想"**试效果 / 问一句 / 生成一张图**" → [`+chat`](../arkcli-chat/SKILL.md) / [`+gen`](../arkcli-gen/SKILL.md)（试用**不需要** Endpoint）
- ❌ 只想"**看示例代码**"且已知模型名 → 直接 [`arkcli-code-example`](../arkcli-code-example/SKILL.md)

## 编排（每步 delegate，不抄细节）

```
Step 0  认证 + 实名闸门
          └─► 见 arkcli-shared「认证闸门」；命中开通/部署意图先做实名检查
              （../arkcli-auth/references/realname-gate.md）

Step 1  选模型
          └─► arkcli-models（search / get）。用户没指定模型就先帮选、确认再继续

Step 2  查是否已有可复用 Endpoint     ★必查★
          └─► arkcli-infer-endpoint：list --mine（"我的"语义见
              ../arkcli-auth/references/identity-resolution.md）
              ├─ 有 EP 且【绑定目标模型/版本】+【状态可用(如 Running)】
              │    多个匹配或不确定时先让用户确认选哪个 ─► 跳到 Step 4（复用其 endpoint-id）
              └─ 无匹配 EP（账号下只有别的模型的 EP 不算匹配）─► Step 3

Step 3  创建 Endpoint                 ★唯一写操作★
          └─► arkcli-deploy：先 --dry-run / 与用户确认，再真建
              （二次确认协议见 arkcli-shared）

Step 4  (可选) 生成调用示例
          └─► arkcli-code-example：按 model-version 提供，覆盖不全；
              缺失时降级到方舟控制台示例页，不当作失败报给用户

Step 5  回执
          └─► 给出 endpoint-id + 一句话"怎么在代码里用它"，回到用户原始目标
```

## 关键约束

- **Step 2 必查、但只复用"匹配"的**：仅当 `list --mine` 里存在**绑定目标模型/版本**且**状态可用(如 Running)**的 endpoint 才跳过 Step 3 复用它；账号下别的模型的 endpoint **不算匹配**，不能拿来顶。多个匹配或拿不准就先让用户确认选哪个，别默默挑一个——否则会给出与目标不一致的 endpoint-id，链路不闭环。无匹配就正常走 Step 3 创建。
- **Step 3 是唯一写操作**：套用 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) 的二次确认 + 实名闸门；用户语气再急也先 `--dry-run` 或确认。
- **Step 4 可选、可降级**：`+code-example` 的示例按 model-version 提供，部分版本后端无 group 会返回 not found；这是预期内的覆盖缺口，降级到控制台示例页即可，不要把它当链路失败。
- 全程是**编排**：每一步把控制权交给 owning skill，不在本文复述其参数。

## 参考

- [arkcli-shared](../arkcli-shared/SKILL.md) — 认证/实名闸门、命令选择顺序、安全与二次确认
- [arkcli-models](../arkcli-models/SKILL.md) — Step 1 选模型
- [arkcli-infer-endpoint](../arkcli-infer-endpoint/SKILL.md) — Step 2 查/复用 Endpoint
- [arkcli-deploy](../arkcli-deploy/SKILL.md) — Step 3 创建 Endpoint
- [arkcli-code-example](../arkcli-code-example/SKILL.md) — Step 4 可选示例代码
