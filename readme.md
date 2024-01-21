<p align="center">
  <a href="https://www.github.com/hoshinonyaruko/gensokyo-hunyuan">
    <img src="pic/2.png" width="200" height="200" alt="gensokyo-hunyuan">
  </a>
</p>

<div align="center">

# Gensokyo-hunyuan

_✨ 基于tencentcloud/hunyuan 的一键混元api连接器 ✨_  

## 特点

可一键对接[Gensokyo框架](https://gensokyo.bot) 仅需配置反向http地址用于接收信息，正向http地址用于调用发送api

基于sqlite数据库自动维系上下文，对话模式中，使用重置 命令即可重置

可作为api运行，也可一键接入QQ频道机器人[QQ机器人开放平台](https://q.qq.com)

## 使用方法
使用命令行运行gensokyo-hunyuan可执行程序

配置config.yml 启动，后监听 port 端口 提供/conversation api

## 接口调用说明

### 终结点

| 属性  | 详情                                   |
| ----- | -------------------------------------- |
| URL   | http://localhost:46230/conversation    |
| 方法  | POST                                   |

### 请求参数

请求体应为JSON格式，包含以下字段：

| 字段名            | 类型   | 描述                           |
| ----------------- | ------ | ------------------------------ |
| `message`         | String | 要发送的消息内容               |
| `conversationId`  | String | 当前对话的唯一标识符           |
| `parentMessageId` | String | 上一条消息的唯一标识符         |

#### 请求示例

```json
{
    "message": "我第一句话说的什么",
    "conversationId": "07710821-ad06-408c-ba60-1a69bf3ca92a",
    "parentMessageId": "73b144d2-a41f-4aeb-b3bb-8624f0e54ba6"
}
```

#### 返回值示例

成功响应将返回状态码 `200` 和一个JSON对象，包含以下字段：

| 字段名            | 类型                    | 描述                        |
| ----------------- | ----------------------- | --------------------------- |
| `response`        | String                  | 接口的响应消息内容          |
| `conversationId`  | String                  | 当前对话的唯一标识符        |
| `messageId`       | String                  | 当前消息的唯一标识符        |
| `details`         | Object                  | 包含额外的使用详情          |
| `usage`           | Object (在 `details` 中) | 使用详情，如令牌计数        |

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


## 兼容性
可在各种架构运行
（原生android暂不支持，sqlitev3需要cgo）
由于cgo编译比较复杂，arm平台，或者其他架构，可试图在对应系统架构下，自行本地编译



## 场景支持

API方式调用
QQ频道直接接入