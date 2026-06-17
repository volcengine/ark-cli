from openai import OpenAI

client = OpenAI(
    api_key="$ARK_API_KEY",
    base_url="https://ark.cn-beijing.volces.com/api/v3",
)

# 非流式对话
chat_completion = client.chat.completions.create(
    model="doubao-seed-2-0-pro",
    messages=[
        {"role": "system", "content": "你是人工智能助手"},
        {"role": "user", "content": "Hello!"},
    ],
    stream=False,
)

print(chat_completion.choices[0].message.content)

