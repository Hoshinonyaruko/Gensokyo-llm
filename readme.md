<p align="center">
  <a href="https://www.github.com/hoshinonyaruko/gensokyo-llm">
    <img src="pic/2.jpg" width="200" height="200" alt="gensokyo-llm">
  </a>
</p>

<div align="center">
# gensokyo-llm

_✨ 适用于Gensokyo以及Onebotv11的大模型一键端 ✨_  
</div> 

---

## 特性

支持所有Onebotv11标准框架.支持http-api和反向ws,支持流式发送,多配置文件(多提示词)

超小体积,内置sqlite维护上下文,支持proxy,

可一键对接[Gensokyo框架](https://gensokyo.bot) 仅需配置反向http地址用于接收信息，正向http地址用于调用发送api

基于sqlite数据库自动维系上下文，对话模式中，使用重置 命令即可重置

可设置system，角色卡,上下文长度,内置多种模型,混元,文心,chatgpt

同时对外提供带有自动上下文的openai原始风味api(经典3参数,id,parent id,messgae)

可作为api运行，也可一键接入QQ频道机器人[QQ机器人开放平台](https://q.qq.com)

可转换gpt的sse类型，递增还是只发新增的sse

并发环境下的sse内存安全，支持维持多用户同时双向sse传输

---

## 安全性

多重完备安全措施,尽可能保证开发者和llm应用安全.

可设置多轮模拟QA强化角色提示词，可自定义重置回复，安全词回复,第一重安全措施

支持多gsk-llm互联,形成ai-agent类应用,如一个llm为另一个llm整理提示词,审核提示词,第二重安全措施

向量安全词列表,基于向量相似度的敏感拦截词列表,先于文本替换进行,第三重安全措施

AhoCorasick算法实现的超高效文本IN-Out替换规则，可大量替换n个关键词到各自对应的新关键词,第四重安全措施

结果可再次通过百度-腾讯,文本审核接口,第五重安全措施

日志全面记录,命令行参数-test 从test.txt快速跑安全自测脚本,

命令行 -mlog 将当前储存的所有日志进行QA格式化,每日审验,从实际场景提炼新安全规则,不断增加安全性,第六重安全措施

语言过滤,允许llm只接受所指定的语言,自动将繁体转换为简体应用安全规则,在自己擅长的领域进行防守,第七重安全措施

提示词长度限制,用最原始的方式控制安全,阻止恶意用户构造长提示词,第八重安全措施

通过这些方法,打造尽可能安全llm对话机器人。

文本IN-OUT双层替换,可自行实现内部提示词动态替换,修改,更安全强大

基于sqlite设计的向量数据表结构,可使用缓存来省钱.自定义缓存命中率,精准度.

针对高效高性能高QPS场景优化的专门场景应用，没有冗余功能和指令,全面围绕数字人设计.

---

## 使用方法
使用命令行运行gensokyo-llm可执行程序

配置config.yml 启动，后监听 port 端口 提供/conversation api

支持中间件开发,在gensokyo框架层到gensokyo-llm的http请求之间,可开发中间件实现向量拓展,数据库拓展,动态修改用户问题.

---

# API接口调用说明

本文档提供了关于API接口的调用方法和配置文件的格式说明，帮助用户正确使用和配置。

---

## 接口支持的查询参数

本系统的 `conversation` 和 `gensokyo` 端点支持通过查询参数 `?prompt=xxx` 来指定特定的配置。

- `prompt` 参数允许用户指定位于执行文件（exe）的 `prompts` 文件夹下的配置YAML文件。使用该参数可以动态地调整API行为和返回内容。

---

## YAML配置文件格式

配置文件应遵循以下YAML格式。这里提供了一个示例配置文件，展示了如何定义不同角色的对话内容：

```yaml
Prompt:
  - role: "system"
    content: "Welcome to the system. How can I assist you today?"
  - role: "user"
    content: "I need help with my account."
  - role: "assistant"
    content: "I can help you with that. What seems to be the problem?"
  - role: "user"
    content: "aaaaaaaaaa!"
  - role: "assistant"
    content: "ooooooooo?"
settings:
  # 以下是通用配置项 和config.yml相同
  useSse: true
  port: 46233
```

---

## 多配置文件支持

### 请求 `/gensokyo` 端点

当向 `/gensokyo` 端点发起请求时，系统支持附加 `prompt` 参数和 `api` 参数。`api` 参数允许指定如 `/conversation_ernie` 这类的完整端点。启用此功能需在配置中开启 `allapi` 选项。

示例请求：
```http
GET /gensokyo?prompt=example&api=/conversation_ernie
```

### 请求 `/conversation` 端点

与 `/gensokyo` 类似，`/conversation` 端点支持附加 `prompt` 参数。

示例请求：
```http
GET /conversation?prompt=example
```

### `prompt` 参数解析

提供的 `prompt` 参数将引用可执行文件目录下的 `/prompts` 文件夹中相应的 YAML 文件（例如 `xxxx.yml`，其中 `xxxx` 是 `prompt` 参数的值）。

### YAML 配置文件

YAML 文件的配置格式请参考 **YAML配置文件格式** 部分。以下列出的配置项支持在请求中动态覆盖：

实现了配置覆盖的函数
- [x] GetWenxinApiPath
- [x] GetGptModel
- [x] GetGptApiPath
- [x] GetGptToken
- [x] GetMaxTokenGpt
- [x] GetUseCache（bool）
- [x] GetProxy
- [x] GetRwkvMaxTokens
- [x] GetLotus
- [x] GetuseSse（bool）
- [x] GetAIPromptkeyboardPath

对于不在上述列表中的配置项，如果需要支持覆盖，请[提交 issue](#)。

所有的bool值在配置文件覆盖的yml中必须指定,否则将会被认为是false.

动态配置覆盖是一个我自己构思的特性,利用这个特性,可以实现配置文件之间的递归,举例,你可以在自己的中间件传递prompt=a,在a.yml中指定Lotus为调用自身,并在lotus地址中指定下一个prompt参数为b,b指定c,c指定d,以此类推.

---

### 终结点

本节介绍了与API通信的具体终结点信息。

| 属性  | 详情                                    |
| ----- | --------------------------------------- |
| URL   | `http://localhost:46230/conversation`   |
| 方法  | `POST`                                  |

### 请求参数

客户端应向服务器发送的请求体必须为JSON格式，以下表格详细列出了每个字段的数据类型及其描述。

| 字段名            | 类型   | 描述                               |
| ----------------- | ------ | ---------------------------------- |
| `message`         | String | 用户发送的消息内容                 |
| `conversationId`  | String | 当前对话会话的唯一标识符            |
| `parentMessageId` | String | 与此消息关联的上一条消息的标识符    |

#### 请求示例

下面的JSON对象展示了向该API终结点发送请求时，请求体的结构：

```json
{
    "message": "我第一句话说的什么",
    "conversationId": "07710821-ad06-408c-ba60-1a69bf3ca92a",
    "parentMessageId": "73b144d2-a41f-4aeb-b3bb-8624f0e54ba6"
}
```

该示例展示了如何构造一个包含消息内容、当前对话会话的唯一标识符以及上一条消息的标识符的请求体。这种格式确保了请求的数据不仅符合服务器的处理规则，同时也便于维护对话上下文的连贯性。

---

#### 返回值示例

成功响应将返回状态码 `200` 和一个JSON对象，包含以下字段：

| 字段名            | 类型                    | 描述                        |
| ----------------- | ----------------------- | --------------------------- |
| `response`        | String                  | 接口的响应消息内容          |
| `conversationId`  | String                  | 当前对话的唯一标识符        |
| `messageId`       | String                  | 当前消息的唯一标识符        |
| `details`         | Object                  | 包含额外的使用详情          |
| `usage`           | Object (在 `details` 中) | 使用详情，如令牌计数        |

---

#### 响应示例

```json
{
    "response": "回答内容",
    "conversationId": "c9b8746d-aa8c-44b3-804a-bb5ad27f5b84",
    "messageId": "36cc9422-da58-47ec-a25e-e8b8eceb47f5",
    "details": {
        "usage": {
            "prompt_tokens": 88,
            "completion_tokens": 2
        }
    }
}
```
---

## 兼容性
可在各种架构运行
（原生android暂不支持，sqlitev3需要cgo）
由于cgo编译比较复杂，arm平台，或者其他架构，可试图在对应系统架构下，自行本地编译

---

## 场景支持

API方式调用
QQ频道直接接入

---

## 约定参数

审核员请求参数

当需要将请求发给另一个 GSK LLM 作为审核员时，应该返回的 JSON 格式如下：

```json
{"result":%s}
```

这里的 `%s` 代表一个将被替换为具体浮点数值的占位符。

气泡生成请求结果

当请求另一个 GSK LLM 生成气泡时，应该返回的 JSON 格式如下：

```json
["","",""]
```

这表示气泡生成的结果是一个包含三个字符串的数组。这个格式用于在返回结果时指明三个不同的气泡，也可以少于或等于3个.

现已不再需要开多个gsk-llm实现类agent功能,基于新的多配置覆盖,prompt参数和lotus特性,可以自己请求自己实现气泡生成,故事推进等复杂特性.

GetAIPromptkeyboardPath可以是自身地址,可以带有prompt参数

当使用中间件指定prompt参数时,配置位于prompts文件夹,其格式xxx-keyboard.yml,若未使用中间件,请在path中指定prompts参数,并将相应的xxx.yml放在prompts文件夹下)

设置系统提示词的gsk-llm联合工作的/conversation地址,约定系统提示词需返回文本json数组(3个).