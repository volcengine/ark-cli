# 测试、调试与监控

提交任务前，或任务失败需要排查时读取本文件。

## RL 必要顺序

自定义 RL 任务按以下顺序推进：

1. 本地单样本 rollout 测试。
2. 本地 rollout 加 grader 联调。
3. 使用代表性数据做本地小批量测试。
4. 使用接近计划 batch size 的并发做压测。
5. 生成项目支持时，执行在线插件或类 FaaS 测试。
6. 使用 `debug=True` 或 `--debug` 提交 RL debug 任务。
7. 提交正式训练任务。

只有用户明确接受风险时，才跳过 debug 任务。

## 非 RL 任务

不要为 SFT、DPO、CPT 或其他非 RL 任务编造 RL debug 流程。按当前 SDK 文档或模板使用其正常校验和提交路径。

## 本地测试

优先使用生成模板中的测试工具。建议覆盖：

- 一条预期通过样本
- 一条预期失败样本
- 畸形工具参数
- 下游 API 超时或报错
- 小规模代表性数据集
- 接近目标 batch size 的并发

API key 使用环境变量：

```bash
export ARK_API_KEY=...
PYTHONPATH=. python plugins/<rollout_or_test>.py
```

如果项目使用 `uv`，遵循生成 README 或模板命令。

## 依赖冻结

本地测试通过后，在线插件测试前冻结依赖：

```bash
python -m pip freeze > requirements.txt
```

或在 `uv` 项目中：

```bash
uv pip freeze > requirements.txt
```

如果模板期望轻量远端环境，清理无关依赖。

## 在线插件测试

使用生成的 `test_faas.py`，或查看当前命令：

```bash
ark test pipeline_plugin --help
```

常见请求形态包含 `context`、`proxy` 和 `sample`，但准确结构以当前模板为准。

## RL Debug 提交

Python 模式：

```python
mcj.submit(debug=True)
```

YAML/命令模式：

```bash
ark create mcj --file job.yaml --debug
```

Debug 任务需要 endpoint/API-key 配置，让 rollout 能调用推理 endpoint。密钥放在环境变量或本地 secret 处理中。

## 正式提交前检查

正式提交前总结：

- SDK 版本和已检查的模板
- 模型引用或自定义模型 id
- 训练方式
- 数据路径和验证策略
- 超参来源
- rollout 和 grader 标识
- 可观测性设置
- 本地、批量、在线、debug 测试结果
- 已知费用或资源影响

## 监控

提交后检查：

```bash
ark list mcj
ark get mcj <mcj-id>
ark pull mcj <mcj-id> --include-plugin
```

关注任务阶段、失败原因、插件日志、轨迹分析、reward 指标、自定义指标、时间线和产物输出。

RL 任务中，在可用时同时观察训练 reward 和验证 reward。不要只看聚合 reward；要检查高 reward 与低 reward 轨迹，确认没有 reward hacking 或 grader 盲区。

## 排障模式

1. 读取任务状态和失败原因。
2. 检查 rollout/grader 日志和轨迹。
3. 用最小失败样本本地复现。
4. 增加有针对性的日志或指标。
5. 修复插件、配置或数据。
6. 重新跑本地、批量、在线和 debug 检查，再开始正式重训。

常见失败点：

- 所选模型版本不支持某个超参
- 数据形态错误或 JSONL 非法
- 缺少环境变量
- 插件依赖不匹配
- 下游工具/API 超时或过慢
- rollout 循环不退出
- 工具参数 JSON 错误
- grader reward 形状不符合预期
- rollout 吞掉了本应扣分的错误
- grader 过于宽松导致 reward hacking
