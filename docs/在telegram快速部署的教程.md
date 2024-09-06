# 在 Telegram 部署 LLM 聊天机器人教程

本教程将指导你使用 `gensokyo-telegram` 和 `gensokyo-llm` 在 Telegram 上部署一个 LLM 聊天机器人。

## 准备工作

### 下载 gensokyo-telegram

首先，下载 `gensokyo-telegram`：
- [gensokyo-telegram GitHub 仓库](https://github.com/Hoshinonyaruko/gensokyo-telegram)

### 创建 Telegram 机器人

1. 打开 Telegram 应用。
2. 搜索 `BotFather`。
3. 创建一个新的机器人，并按照 `BotFather` 提供的指引完成设置。

你将收到如下消息：

```
Done! Congratulations on your new bot. You will find it at t.me/Txxx. You can now add a description, about section, and profile picture for your bot, see /help for a list of commands. By the way, when you've finished creating your cool bot, ping our Bot Support if you want a better username for it. Just make sure the bot is fully operational before you do this.
```

**重要**：保护好你的 HTTP API Token，它将用于接下来的步骤。

### 访问和配置机器人

- 访问你的机器人链接：[t.me/Txxx](t.me/Txxx)。
- 使用收到的 `<<Your Token>>` 作为你的 `botToken`。

## 配置 gensokyo-telegram

1. 首次运行 `.exe` 文件，按提示释放脚本。
2. 运行 `.bat` 文件。
3. 打开 `config.yml` 配置文件，进行以下设置：
   - `botToken`: 填入你的 `botToken`。
   - `useNgrok`: 设置为 `true`。
   - `webHookPath`: 保持为空 (`""`)。
   - `customcert`: 设置为 `false`。

### 配置 ngrok

1. 访问 [ngrok 官网](https://dashboard.ngrok.com/get-started/setup/windows)，并注册或登录。
2. 在获取开始（getting started）部分找到 `your authtoken`。
3. 将 `authtoken` 输入到 `ngrokKey` 配置中。
4. 设置 `highway` 为 `true`。
5. 设置 `sendDirectResponse` 为 `true`。

## 配置 gensokyo-llm

1. 在 `docs/中级-轻松对接豆包大模型.md` 完成豆包模型的配置。
2. 确保 `gensokyo-llm` 的 `iPWhiteList` 包含 `127.0.0.1`。
3. 使用默认端口 `46233`。

### 连接 gensokyo-telegram 和 gensokyo-llm

1. 打开 `gensokyo-telegram` 的 `yml` 配置文件。
2. 添加 `gensokyo-llm` 的反向 WebSocket 地址到 `ws_address` 配置：
   ```yaml
   ws_address: ["ws://127.0.0.1:46233"]
   ```
3. 在 `config.yml` 的 `systemPrompt` 配置项中配置好提示词。

### 运行和测试

1. 双击运行 `gensokyo-llm`。
2. 双击运行你已配置好的 `gensokyo-telegram`。
3. 发送信息给你的 bot，检查是否能成功接收信息。

![](/pic/5.png)