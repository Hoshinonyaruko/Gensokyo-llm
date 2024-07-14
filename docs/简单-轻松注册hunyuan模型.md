# 如何配置腾讯混元大模型API

本教程将指导你如何在腾讯云控制台创建访问密钥，并配置混元大模型的API。请按照以下步骤操作。

## 第一步：访问腾讯混元大模型控制台

首先，打开腾讯混元大模型控制台：

[腾讯混元大模型控制台](https://console.cloud.tencent.com/hunyuan/start)

## 第二步：创建秘钥

1. 在控制台页面，点击“创建秘钥”按钮。
2. 点击界面左侧的“继续”按钮，完成秘钥的创建过程。
3. 获取并记录下你的 `SecretID`。
4. 同时，将 `SecretKey` 保存到一个安全且不易丢失的地方，例如加密的备忘录。

## 第三步：配置 `config.yml`

在项目的 `config.yml` 文件中填写你刚刚获取的 `SecretID` 和 `SecretKey`。请参考以下配置，填写相应的字段：

```yaml
secretId: ""                                   # 腾讯云账号(右上角)-访问管理-访问密钥，生成获取
secretKey: ""
region: ""                                     # 留空
maxTokensHunyuan: 4096                         # 最大上下文
hunyuanType: 0                                 # 选择使用的混元版本
hunyuanStreamModeration: false                 # 是否采用流式审核
topPHunyuan: 1.0                               # 累积概率最高的令牌进行采样的界限
temperatureHunyuan: 1.0                        # 生成的随机性控制
```

### 混元类型选择说明：

- `0`: 高级版
- `1`: 标准版 std
- `2`: hunyuan-lite
- `3`: hunyuan-standard
- `4`: hunyuan-standard-256K
- `5`: hunyuan-pro

确保 `apiType` 保持为 `0`，这样你的设置将正确应用：

```yaml
apiType: 0
```

完成以上步骤后，你就成功设置了混元的 API。现在，你可以将混元的 API 转化为通用 API，供各种服务和 app 使用。