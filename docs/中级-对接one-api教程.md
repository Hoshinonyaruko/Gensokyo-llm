### 开始使用 One-API对接gsk-llm

**步骤 1: 下载 One-API**
- One-API 是一个轻便易用的项目，包含一个可执行文件，无需其他环境支持，且带有 Web UI。
- 下载链接：[one-api](https://github.com/hoshinonyaruko/gensokyo-llm)

**步骤 2: 运行 One-API**
- 执行下载的 One-API 可执行文件。
- 在浏览器中打开 `http://localhost:3000/`。

**步骤 3: 登录**
- 使用默认用户名 `root` 和密码 `123456` 登录。

**步骤 4: 创建 API 渠道**
- 在网页控制台的顶栏选择“渠道”-“添加新的渠道”。
- 为你的模型渠道命名，并在模型栏中输入你申请的模型名称，该栏支持自动补全。
- 输入你从模型所在平台（如腾讯云、智谱、通义等）获取的 API access token。
- 点击“提交”以创建 API 渠道。

**步骤 5: 生成令牌**
- 点击顶栏的“令牌”并创建一个新令牌。
- 选择要使用的模型，创建令牌后点击绿色的“复制”按钮复制生成的令牌。

**步骤 6: 配置 gsk-llm**
- 在 gsk-llm 配置文件中更新以下配置以连接到你的 one-api 平台。
```yaml
# One-API 测试配置
gptModel: "chatglm_turbo"  # 使用的模型名称
gptApiPath: "http://127.0.0.1:3000/v1/chat/completions"  # One-API 服务的端口号
gptToken: "sk-dbmr0Oxxxxxxxxxxxxxxxxxxxxxxx"  # 生成的密钥
maxTokenGpt: 1024
gptModeration: false
gptSafeMode: false
gptSseType: 0
standardGptApi: true  # 启用标准 GPT API
```

这样配置后，你就可以灵活地管理用量和使用的模型了。本项目的配置文件是热更新的,你不需要重启来应用配置.

