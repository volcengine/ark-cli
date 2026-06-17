# models activate

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

主动开通指定基础模型的服务（对应 URL2 章节「三、模型激活」中的 `OpenModelChargeItem` Action，API 显示名「开通模型服务」）。

## 与 deploy / infer-create 隐式开通的区别

| 入口 | 触发方式 | 适用场景 |
|------|---------|---------|
| `arkcli models activate <name>` | 用户**主动**开通 | 提前为业务做好模型可用性准备；想试用新模型；测 dry-run |
| `arkcli deploy ...` / `arkcli infer endpoint create ...` | **被动**触发：检测到模型未开通时自动 prompt | 用户的真实目标是 deploy / 创建端点，开通只是前置依赖 |

`activate` 不会先 GetModelChargeItem 检测，而是直接发起 OpenModelChargeItem。重复对已开通模型调用是幂等的（后端处理）。

## 命令

```bash
# 默认仅开通基础服务（推理 + 精调）— 等价于 SubServices=["base"]
arkcli models activate doubao-seed-1-6-flash

# 同时开通基础服务与低延迟推理子服务（URL2 完整请求示例）
arkcli models activate doubao-seed-1-6-flash --sub-services base,fast-infer

# dry-run 校验（仅校验请求合法性，不实际开通；自动跳过 [Y/N] 提示）
arkcli models activate doubao-seed-1-6-flash --sub-services base,fast-infer --dry-run

# 非交互（CI / 脚本场景，跳过 [Y/N] 确认提示）
arkcli models activate doubao-seed-1-6-flash --yes
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<foundation-model-name>` | 是 | string | 待开通的基础模型名，可通过 `arkcli models list` / `search` 查询 |
| `--sub-services` | 否 | string list | 待开通的子服务，逗号分隔。可选值：`base` / `context-cache` / `fast-infer`。**不传时后端默认仅开通 `base`** |
| `--yes` | 否 | bool | 跳过 `[Y/N]` 二次确认。CI / 脚本中必传 |

### `--sub-services` 取值（URL2 权威定义）

| 值 | 含义 |
|----|------|
| `base` | 基础服务（推理 + 精调） |
| `context-cache` | 上下文缓存 |
| `fast-infer` | 低延迟推理 |

不传时后端等价于 `["base"]`。CLI 先做本地校验：传非枚举值会立刻拒绝，避免发出会被后端 `InvalidParameter.SubServices` 反弹的请求。

### 全局参数

`--dry-run`（继承自全局）会让后端**仅校验请求合法性，不真正开通**。dry-run 模式自动跳过本地 `[Y/N]` 确认，且 service 层不再 polling 状态（后端不会真的 Available）。

## 返回值

成功返回（JSON）：

```json
{
  "foundation_model": "doubao-seed-1-6-flash",
  "state": "Available",
  "sub_services": ["base", "fast-infer"]
}
```

dry-run 返回：

```json
{
  "dry_run": true,
  "validated": true,
  "foundation_model": "doubao-seed-1-6-flash",
  "sub_services": ["base", "fast-infer"]
}
```

## 常见错误

| 错误 | 原因 | 处理方式 |
|------|------|---------|
| `invalid --sub-services value "foo"` | 子服务名拼写错误 | 用 `base` / `context-cache` / `fast-infer` |
| `操作已取消` | 在 `[Y/N]` 提示中输入了非 Y | 重跑命令并确认，或加 `--yes` |
| 模型不存在 | `<foundation-model-name>` 错误或模型已下线 | `arkcli models search` 找正确名字 |
| `timed out waiting for model to become available` | 开通成功但 5 秒内未变 Available | 通常无害，稍后重试 `arkcli models get <name>` 看真实状态 |

## 注意事项

- 子服务可以增量开通：先 `--sub-services base` 再 `--sub-services context-cache`，互不冲突
- `--dry-run` 不计入用量，可放心使用以校验请求合法性
- 已开通的模型再次 activate 是幂等行为（后端约定）

## 参考

- [arkcli-models](../SKILL.md) — models skill 概览
- [arkcli-shared](../../arkcli-shared/SKILL.md) — 认证和全局参数
- URL2 文档章节「三、模型激活」 → `OpenModelChargeItem`
