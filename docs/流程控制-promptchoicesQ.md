#  `promptChoicesQ` 配置格式文档

## 配置格式概述

新的 `promptChoicesQ` 配置使用 `PromptChoice` 结构体来替代原有的字符串数组格式。根据是否包含关键字 `Keywords` 决定了匹配方式。

## `PromptChoice` 结构体

### 结构体定义

```go
type PromptChoice struct {
    Round       int      // 轮次编号
    ReplaceText []string   // 替换词列表
    Keywords    []string // 匹配词列表
}
```

### 字段说明

- `Round`: 轮次编号，表示在第几轮对话中应用此配置。
- `ReplaceText`: 替换词列表，符合条件时添加到用户当前输入中的文本。多个将会随机一个。
- `Keywords`: 匹配词列表，用户输入中包含任意一个匹配词时，将替换词添加到用户当前输入中。如果 `Keywords` 为空，则表示在该轮次中随机选择一个 `ReplaceText`。

## 配置示例

### 配置格式

```yaml
promptChoicesQ:
  - Round: 1
    ReplaceText: ["回家吧"]
    Keywords: ["我累了", "不想去了"]
  - Round: 2
    ReplaceText: ["我们打车去"]
    Keywords: ["快点去", "想去", "早点"]
  - Round: 3
    ReplaceText: ["我们走着去"]
    Keywords: ["不着急", "等下"]
  - Round: 1
    ReplaceText: ["放松一下"]
    Keywords: [] # 相当于 enhancedChoices = false
```

### 示例解释

1. **第一轮对话**：
   - 如果用户输入包含 "我累了" 或 "不想去了"，则在用户输入中添加 "回家吧"。
   - 如果 `Keywords` 为空（例如 "放松一下" 这条配置），则随机选择符合当前轮次的 `ReplaceText` 之一。
2. **第二轮对话**：
   - 如果用户输入包含 "快点去"、"想去" 或 "早点"，则在用户输入中添加 "我们打车去"。
3. **第三轮对话**：
   - 如果用户输入包含 "不着急" 或 "等下"，则在用户输入中添加 "我们走着去"。

## 重要说明
- 根据 `Keywords` 是否为空来决定匹配方式。
- 如果 `Keywords` 为空，则在符合当前轮次的所有 `choice` 中随机选择一个 `ReplaceText`。
- 如果 `Keywords` 不为空，则根据用户输入的匹配词数量选择最适合的 `ReplaceText`。