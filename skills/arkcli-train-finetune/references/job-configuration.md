# 任务配置

创建或审查 `job.py`、`job.yaml`、数据设置、超参或提交命令时读取本文件。

## 需要检查的文件

SDK 生成的 workspace 通常包含：

- `job.py`：Python 对象配置，包含 `ModelCustomizationJob(...)` 和 `submit(...)`。
- `job.yaml`：声明式任务配置，通过 `ark create mcj` 提交。
- `arkworkspace.toml`：workspace 元数据。
- `requirements.txt`：远端插件运行环境依赖。
- `plugins/`：rollout、grader 和测试辅助函数。
- `test_faas.py` 或等价的在线插件测试文件。

编辑前必须先检查当前 SDK 版本生成的文件。模板里的 import 和字段名比旧示例更可靠。

## 必要配置区域

多数任务需要：

- `name`：可读任务名。
- `customization_type`：当前 SDK 接受的训练方式枚举或字符串。
- `model_reference`：带名称和版本的 `foundation_model`，或 `custom_model_id`。
- `data`：训练集和可选验证集。
- `hyperparameters`：所选模型版本支持的字段。
- `custom_rl_pipeline`：自定义 RL 任务需要。

## 数据输入

训练集和验证集通常可使用以下模式，具体以当前 SDK 规则为准：

- `local_files`：workspace 内的本地 JSONL 文件。
- `tos_bucket` 加 `tos_paths`：对象存储文件。
- `datasets`：数据集 id、版本 id 和可选混入控制。

验证集可以是显式 `validation_set`，也可以是 `validation_percentage`。互斥规则以生成类型或当前文档为准。

## 超参

不要复制旧文档里的超参表。针对当前模型版本查询：

```bash
ark get foundation-model --model <model-name> --version <model-version> --fields hyperparameters
```

只把返回的受支持字段映射到 `job.py` 或 `job.yaml`。如果生成模板把超参值写成字符串，保持这种风格。

## Python 配置模式

优先改造生成的 `job.py`。稳定结构如下：

```python
from ark_sdk.resources.model_customization_job import ModelCustomizationJob

mcj = ModelCustomizationJob(
    name="sdk-job",
    model_reference=...,
    customization_type=...,
    hyperparameters={...},
    data=...,
    custom_rl_pipeline=...,  # 仅 RL 需要
    enable_trajectory=True,  # RL 且 SDK 支持时使用
)
mcj.submit()
print(mcj.url)
```

RL debug 任务中，给 `submit(...)` 传入 `debug=True`，并用当前 SDK 类型名配置 debug endpoint 和 API key。

## YAML 配置模式

优先改造生成的 `job.yaml`。稳定结构如下：

```yaml
name: sdk-job
customization_type: GRPOLoRA
model_reference:
  foundation_model:
    name: <model-name>
    model_version: "<model-version>"
hyperparameters:
  batch_size: "32"
data:
  training_set:
    local_files:
      - ./data/train.jsonl
custom_rl_pipeline:
  graders:
    - plugin:
        name: reward_fn
        python_func: plugins.reward:reward_fn
      weight: 1.0
  rollout:
    plugin:
      name: rollout_fn
      python_func: plugins.rollout:rollout_fn
enable_trajectory: true
```

提交：

```bash
ark create mcj -f job.yaml
```

仅当 RL 任务具备有效 debug 配置时，才使用：

```bash
ark create mcj --file job.yaml --debug
```

## 资源组

当用户需要稳定或专用资源时，先检查当前资源组支持：

```bash
ark list resource_group --mcj --file job.yaml
ark list resource_group --mcj --file job.yaml --only-matched
```

除非当前 SDK help 明确支持，否则不要把 debug 提交和 resource group 同时使用。
