# auth status / logout

## 查看状态

```bash
arkcli auth status
```

输出通常包含：

- `volc_sso` 或 `aksk`（取决于当前 profile 的 tenant 和登录方式）
- `volc_sso.identity`：当前火山账号身份事实（`name` / `account_id` / `trn` / `is_root`），并附带账号实名认证状态：
  - `verified`：是否完成实名认证（`true` / `false`）
  - `verify_type`：实名主体类型，`individual`（个人）或 `enterprise`（企业）；**仅在已实名时出现**
  - 实名探测失败（网络 / STS / 权限）时 `verified` 与 `verify_type` **两个字段都省略**（与"未探测"语义一致，绝不假报已实名/未实名）
  - 实名是账号级事实，与 profile / project / 登录方式无关；实名是"开通模型 / 开通云产品 / 创建推理接入点"的必要前提
- `ark_api_key`：当前缓存的 ARK API Key
- `ark_api_key.key / status`：当前 Key 掩码值，以及与远端列表对齐后的状态
- `status` 取值为 `active`、`disabled`、`notfound`，远端不可判断时为 `unknown`
- `project_name`：当前生效的 ARK Project Name

敏感字段会做掩码处理，不会直接打印完整 token / secret。

## Project Name 排障

- 如果 `project_name` 不是预期值：
  1. 优先看是否被 `--project-name` flag 或 `ARK_PROJECT_NAME` 环境变量临时覆盖
  2. 其次看 active profile.yaml 的 `profile.project` 字段 (0.1.16 起持久化路径; `arkcli profile create --project <name>` 写入, cfg 装配映射到 `cfg.ProjectName`)
  3. 再看 identity store apikey 关联的 project (SSO 登录 / `auth apikey` 选 key 时写入)
  4. 最后看 `.env` (老用户兼容路径)
  5. 都不设置时，火山默认值为 `"default"`
- 想把 Project Name 固定为某个值：优先 `arkcli profile create --project <name>` 走 profile.yaml; 临时覆盖 `export ARK_PROJECT_NAME=<name>` 或 `arkcli auth apikey` 选中对应项目下的 Key
- 老 `arkcli config init --project-name ...` 已废弃, 不应再建议

## 退出登录

```bash
arkcli auth logout
```

该命令会删除本地存储的认证信息，执行前应确认用户确实要清理凭证。
