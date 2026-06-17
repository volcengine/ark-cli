---
name: ark-finetune-sdk
description: 用于创建、配置、测试、调试、提交、监控或排查模型精调 SDK 项目，尤其是包含 rollout 和 grader 插件的强化学习任务。当请求涉及 `ark` SDK 命令入口、`ark_sdk` Python 包、`job.py`、`job.yaml`、`arkworkspace.toml`、`custom_rl_pipeline`、`ModelCustomizationJob`、rollout/grader 函数、插件测试、RL debug 任务、轨迹分析或精调产物时使用。
---

# Ark Finetune SDK

这是通过精调 SDK 执行模型定制任务的操作指南。优先相信当前 SDK 行为和最新生成的模板，而不是记忆中的旧示例。

## 信息优先级

1. 当前已安装 SDK 和命令输出。
2. 使用 `ark init workspace` 新生成的 SDK 模板。
3. 需要用户可见引用时，使用官方公开产品文档。
4. 本 skill 附带的 reference。

不要把访问标签、环境专属说明、准入清单、变化较快的模型支持列表或敏感来源链接写进用户可见输出或生成文件。

## 起步检查

处理任何实质性任务前：

1. 确认 SDK 版本和是否存在更新。
   ```bash
   ark upgrade --dry-run
   python -m pip show ark-sdk
   ```
2. 依赖参数前先查看当前命令面。
   ```bash
   ark --help
   ark init workspace --help
   ark create mcj --help
   ark test pipeline_plugin --help
   ```
3. 如果要创建或改造项目，先在临时目录用当前 SDK 初始化模板，并阅读生成文件后再改用户项目。
   ```bash
   ark init workspace <workspace-dir> --template <template-name>
   ```
4. 判断任务类型：SFT/DPO/CPT 类监督精调、带自定义 rollout/grader 的 RL、Agent RL，或产物管理。
5. 涉及写入或可能产生费用的操作时，如果用户没有明确要求提交，先总结最终配置并等待明确意图。

## Reference 路由

只加载需要的文件：

- 任务与数据配置：[references/job-configuration.md](references/job-configuration.md)
- RL rollout/grader 实现：[references/rl-plugins.md](references/rl-plugins.md)
- 测试、调试、监控与排障：[references/testing-debug-monitoring.md](references/testing-debug-monitoring.md)

## 核心流程

执行 SDK 精调任务时：

1. 使用 `ark login` 或当前 SDK 支持的命令级凭证完成环境和鉴权准备。
2. 初始化或检查 SDK workspace。
3. 配置 `job.py` 或 `job.yaml`：模型引用、训练方式、数据、超参、可选验证集、可选产物配置。
4. 对 RL 任务，先实现或改造 rollout，再实现 grader，最后接入 `custom_rl_pipeline`。
5. 在正式测试前补齐可观测性：轨迹、函数日志、指标，以及可用的 tracing。
6. 用一条预期通过样本和一条预期失败样本做本地测试。
7. 用小批量数据测试，并发接近计划 batch size。
8. 模板支持时，执行在线插件或类 FaaS 测试。
9. RL 正式训练前提交 debug 任务，除非用户明确接受跳过风险。
10. 提交正式任务后，监控状态、日志、轨迹、reward 指标和产物。

## 稳定规则

- 不硬编码超参表。针对所选模型和版本实时查询。
  ```bash
  ark get foundation-model --model <model-name> --version <model-version> --fields hyperparameters
  ```
- 除非当前 SDK 文档或模板明确允许，否则 `foundation_model` 和 `custom_model_id` 互斥。
- 本地数据路径尽量保持为项目内相对路径；生成模板通常按相对路径工作。
- RL 任务中，SDK 支持且有排障价值时设置 `enable_trajectory=True`。
- 不在代码、日志或最终答复中暴露 API key、AK/SK、token、endpoint JWT 或 Authorization header。优先使用环境变量或本地鉴权。
- 不凭记忆编造 SDK 类名或装饰器。名称不一致时，检查生成模板 import 和已安装的 `ark_sdk`。
- 不依赖本 skill 获取变化较快的支持矩阵。使用当前 SDK 输出、生成模板和官方文档确认。

## 常用命令

```bash
ark upgrade --dry-run
ark login
ark init workspace <workspace-dir> --template rl_demo
ark init workspace <workspace-dir> --template rl_search_mcp_demo
ark get foundation-model --model <model-name> --version <model-version> --fields hyperparameters
python job.py
ark create mcj -f job.yaml
ark create mcj --file job.yaml --debug
ark get mcj <mcj-id>
ark list mcj
ark pull mcj <mcj-id> --include-plugin
```

命令参数不确定时，优先执行 `ark <command> --help`。
