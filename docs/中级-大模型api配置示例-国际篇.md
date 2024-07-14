### 国际模型配置实例

#### api2d 配置项
```yaml
gptModel: "gpt-3.5-turbo"
gptApiPath: "https://openai.api2d.net/v1/chat/completions"
gptToken: "fk207628-***********"
maxTokenGpt: 1024
gptModeration: false  # 额外走腾讯云检查安全,不合规直接拦截. 仅api2d支持
gptSafeMode: false
gptSseType: 0
```

#### openai 测试配置
```yaml
gptModel: "gpt-3.5-turbo"
gptApiPath: "https://api.openai.com/v1/chat/completions"
gptToken: "sk_8*******"
maxTokenGpt: 1024
gptModeration: false  # 额外走腾讯云检查安全,不合规直接拦截. 仅api2d支持
gptSafeMode: false
gptSseType: 0
standardGptApi: true  # 标准的gptApi, OpenAI 和 Groq 需要开启
```

#### Groq 测试配置
```yaml
gptModel: "llama3-70b-8192"
gptApiPath: "https://api.groq.com/openai/v1/chat/completions"
gptToken: "gsk_8*******"
maxTokenGpt: 1024
gptModeration: false  # 额外走腾讯云检查安全,不合规直接拦截. 仅api2d支持
gptSafeMode: false
gptSseType: 0
standardGptApi: true  # 标准的gptApi, OpenAI 和 Groq 需要开启
```

#### One-API 测试配置(该项目也支持国内多个大模型,具体请参考one-api接入教程)
[one-api](https://github.com/hoshinonyaruko/gensokyo-llm)
```yaml
gptModel: "chatglm_turbo"
gptApiPath: "http://127.0.0.1:3000/v1/chat/completions"
gptToken: "sk-d*****"
maxTokenGpt: 1024
gptModeration: false  # 额外走腾讯云检查安全,不合规直接拦截. 仅api2d支持
gptSafeMode: false
gptSseType: 0
standardGptApi: true  # 标准的gptApi, OpenAI 和 Groq 需要开启
```