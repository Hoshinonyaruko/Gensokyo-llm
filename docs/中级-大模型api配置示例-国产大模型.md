# 国产大模型配置实例

### 混元配置项
配置混元模型需要以下参数：
```yaml
secretId: "xxx"
secretKey: "xxx"
region: ""
maxTokensHunyuan: 1024
hunyuanType: 3  # 可选类型：0=高级版, 1=标准版std, 2=hunyuan-lite, 3=hunyuan-standard, 4=hunyuan-standard-256K, 5=hunyuan-pro
hunyuanStreamModeration: true  # 启用流式内容审核
topPHunyuan: 1.0
temperatureHunyuan: 1.0  # 控制生成内容的随机性
```

### 文心配置项
配置文心模型时需要以下参数，这里涵盖了接口的调用以及特殊参数的设置：
```yaml
wenxinAccessToken: "xxx"
wenxinApiPath: "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant"
wenxinEmbeddingUrl: "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/embeddings/embedding-v1"  # 百度的embedding接口URL
maxTokenWenxin: 1024
wenxinTopp: 0.7  # 控制输出文本的多样性，取值范围0.1~1.0，默认0.7
wenxinPenaltyScore: 1.0  # 增加生成token的惩罚，以减少重复，值越大惩罚越重
wenxinMaxOutputTokens: 100  # 模型最大输出token数，范围2~1024
```

### RWKV 模型配置(rwkv runner)
```yaml
rwkvApiPath: "https://"  # 符合RWKV标准的API地址
rwkvMaxTokens: 100
rwkvTemperature: 1.1
rwkvTopP: 0.5
rwkvPresencePenalty: 0.5
rwkvFrequencyPenalty: 1.1
rwkvPenaltyDecay: 0.99
rwkvTopK: 25
rwkvSseType: 0
rwkvGlobalPenalty: false
rwkvStop:
  - "\n\nUser"
rwkvUserName: "User"
rwkvAssistantName: "Assistant"
rwkvSystemName: "System"
rwkvPreSystem: false
```

### TYQW 模型配置
```yaml
tyqwApiPath: "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
tyqwMaxTokens: 1500
tyqwModel: "qwen-turbo"
tyqwApiKey: "sk-"
tyqwWorkspace: ""
tyqwTemperature: 0.85
tyqwTopP: 0.9
tyqwPresencePenalty: 0.2
tyqwFrequencyPenalty: 0.2
tyqwRepetitionPenalty: 1.1
tyqwPenaltyDecay: 0.99
tyqwTopK: 40
tyqwSeed: 1234
tyqwSseType: 1
tyqwGlobalPenalty: false
tyqwStop:
  - "\n\nUser"
tyqwUserName: "User"
tyqwAssistantName: "Assistant"
tyqwSystemName: "System"
tyqwPreSystem: false
tyqwEnableSearch: false
```

### GLM 模型配置
```yaml
glmApiPath: "https://open.bigmodel.cn/api/paas/v4/chat/completions"
glmModel: "glm-3-turbo"
glmApiKey: ".xxx"
glmRequestID: ""
glmDoSample: true
glmTemperature: 0.95
glmTopP: 0.9
glmMaxTokens: 1024
glmStop:
  - "stop_token"
glmTools:
  - ""
glmToolChoice: "auto"
glmUserID: ""
```