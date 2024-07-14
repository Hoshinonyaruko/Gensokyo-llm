## 开始使用 Discord 开发者平台

### 成为开发者

首先，访问 [Discord 开发者主页](https://discord.com/developers/applications) 申请成为开发者。

### 进入 Discord 开发者门户

1. 访问 [Discord 开发者门户](https://discord.com/developers/applications) 并选择或创建一个应用。
2. 在应用详情页面，创建一个机器人：
   - 转到“Bot”部分。
   - 点击“Add Bot”并确认。
   - 在机器人页面找到 Token，复制并留作备用。

### 下载 gensokyo-discord

接下来，下载 `gensokyo-discord`：
- [gensokyo-discord GitHub 仓库](https://github.com/Hoshinonyaruko/gensokyo-discord)

## 配置 gensokyo-discord

1. 首次运行 `.exe` 文件，按提示释放脚本。
2. 运行 `.bat` 文件。
3. 打开 `config.yml` 配置文件，进行以下设置：

```yaml
token: "你的机器人Token"
```

## 配置 gensokyo-llm

1. 在 `docs/中级-轻松对接豆包大模型.md` 完成豆包模型的配置。
2. 确保 `gensokyo-llm` 的 `iPWhiteList` 包含 `127.0.0.1`。
3. 使用默认端口 `46233`。

### 连接 gensokyo-discord 和 gensokyo-llm

1. 打开 `gensokyo-discord` 的 `yml` 配置文件。
2. 添加 `gensokyo-llm` 的反向 WebSocket 地址到 `ws_address` 配置：
   ```yaml
   ws_address: ["ws://127.0.0.1:46233"]
   ```
3. 在 `config.yml` 的 `systemPrompt` 配置项中配置好提示词。

### 运行和测试

1. 双击运行 `gensokyo-llm`。
2. 双击运行你已配置好的 `gensokyo-discord`。
3. 发送信息给你的机器人，检查是否能成功接收信息。