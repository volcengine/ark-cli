# 实名认证闸门（开通 / 写资源类命令）

> 本文从 `arkcli-shared` 拆出。仅在命中"开通模型 / 部署 / 精调 / 激活模型"这类会触发 `OpenModelChargeItem` 的写资源意图时加载。常规读任务不需要读本文。

部分命令会触发**开通模型**（`OpenModelChargeItem`），而开通模型要求账号**已完成实名认证**——未实名时后端必拒（`OperationDenied` / cause `NotVerifiedAccount`）。对这类命令，在「认证闸门」通过之后、执行命令之前，**再追加一道实名检查**：

适用命令（凡会触发开通模型的）：

- `arkcli +deploy`
- `arkcli infer endpoint create`
- `arkcli train finetune create`
- `arkcli models activate`

**执行顺序是硬约束**：命中上述任一意图（开通 / 部署 / 精调）时，**第一步**就做实名检查——在跑 `models get/search`、`+deploy`、`activate`、`infer endpoint create` 等**任何业务命令之前**，先 `auth status` 读 `verified`。**不要**先去试 deploy/activate、撞到"模型未开通"或"model ID invalid"之类报错再回头查实名——那些报错会把你带偏，让你忘了真正的前置是实名。

检查步骤：

1. 读 `arkcli auth status` 输出里的 `volc_sso.identity.verified`（0.1.17+ 字段）：
   - `verified == true` → 已实名，放行，继续目标命令。
   - `verified == false` → 账号未实名。**立即停**：用一句话告诉用户"检测到账号未实名，开通模型 / 部署 / 精调需要先完成实名认证"，并给出实名认证页：`https://console.volcengine.com/user/authentication/detail/`。**在此暂停等待**用户完成实名后回来告知，再继续目标命令；**不要原地轮询、不要改用别的模型 ID 重试、不要继续跑 deploy/activate 去碰运气**（与 SSO 登录失败的"贴回链接 + 等用户"处理一致）。
   - **`verified` 字段缺失**（探测失败 / AK-SK 模式无此字段）→ **不要硬卡**，正常继续执行；实名是火山账号概念，字段缺失视为"未知 / 不适用"，交由命令失败时的 error-hint 兜底，**不要假报**已实名或未实名。
2. 实名是**账号级**前置，与 profile / project / 登录方式无关；它是"开通模型 / 开通云产品 / 创建推理接入点"的共同必要条件。
3. **代码兜底（0.1.17+）**：即便漏了上面的前置检查、直接跑 `+deploy` / `infer endpoint create` / `train finetune create`，`EnsureModelAvailable` 也会在开通前做一次实名预检，未实名时直接返回结构化 `account_not_verified` hint（不再是误导的"操作已取消：模型未开通"）。这是最后一道网，不是免去主动检查的理由——主动在第一步检查能给用户更早、更清晰的引导。
