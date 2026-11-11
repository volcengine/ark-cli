# 精调创建回归用例

使用用户明确授权的目标模型、版本、类型和数据；以下用例不是提交线上任务的授权。

| 用例 | 输入场景 | 期望行为 |
|---|---|---|
| control-plane-auth | 用户拿 `.env` / `--api-key` 要求提交 `train finetune create` | 说明精调是控制面，只走 SSO/STS；不把 API Key 传给训练创建；不切 Agent Plan 或改用 `+chat`/`+gen`。 |
| sft-route | 用户说“我要做 SFT / 精调这个模型” | 进入本 skill 的 `train finetune` 控制面流程；先查能力/价格，再确认数据与创建，不按 Agent Plan 模型调用选型。 |
| mine-list | 用户说“查我创建的精调任务” | 第一条列表命令包含 `--mine`；不先拉全账号再本地过滤；降级身份明确报不支持。 |
| job-data-download | 用户要下载某精调任务的训练集/验证集 | 使用 `train finetune download <job-id>`；按 `TosBucket/TosPaths` 或 Dataset 版本下载，preset/local-only 来源列为 skipped，不猜路径。 |
| local-train-validation-estimate | 已授权上传训练和验证两个本地 JSONL，要求在线估算 | 使用 `create --estimate`；两组均上传并显示 TOS 引用，`uploaded_files` 包含两组，训练 token 不包含验证集；不创建任务。 |
| remote-train-local-validation | 训练集为已存在的 TOS URI，验证集为本地文件 | 验证集仍上传并转换为 TOS 引用；缺失文件或上传失败时停止提交，不重放未知结果的创建。 |
| volc-create-url | 用户确认提交，创建返回任务 ID | 只按返回 ID 和 `console.volcengine.com` URL 报告创建结果，不声称训练已完成。 |
| unsupported-local-preview | 要求 `create --dry-run` | 说明当前命令不支持本地 Client Preview；在线 `--estimate` 有上传副作用，不静默替换执行。 |
| export-auto-quantization | 用户在精调提交时选择 W4A8C16 | 按精确模型版本/训练方法能力使用 --output-compression-strategies，保留必选项，估算展示 output_config；非法选择在上传前停止，不放入 hyperparameters 或替换为事后 quantize。 |
