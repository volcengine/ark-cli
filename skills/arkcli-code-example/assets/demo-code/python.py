# Install: pip install openai
#
# 模板说明:
# - model 使用带版本号的合并 ID (模型名-PrimaryVersion);
#   用 `arkcli models get <name> --format json` 查 PrimaryVersion 后拼合
# - $ARK_API_KEY 运行前由环境变量提供, 不要把真实 Key 写进文件
import os

from openai import OpenAI

client = OpenAI(
    api_key=os.getenv("ARK_API_KEY", "$ARK_API_KEY"),
    base_url="https://ark.cn-beijing.volces.com/api/v3",
)

# 非流式对话
chat_completion = client.chat.completions.create(
    model="doubao-seed-2-0-pro-260215",
    messages=[
        {"role": "system", "content": "你是人工智能助手"},
        {"role": "user", "content": "Hello!"},
    ],
    stream=False,
)

print(chat_completion.choices[0].message.content)
