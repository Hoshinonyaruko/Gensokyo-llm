package template

const ConfigTemplate = `
version: 1
settings:

  #通用配置项
  allApi : false                                #以conversation_ernie conversation_hunyuan形式同时开启全部api,请设置好iPWhiteList避免被盗用.
  useSse : 0                                    #智能体场景开启,其他场景,比如普通onebotv11不开启 0、1=false 2=true
  port : 46233                                  #本程序监听端口,支持gensokyo http上报,   请在gensokyo的反向http配置加入 post_url: ["http://127.0.0.1:port/gensokyo"] 
  selfPath : ""                                 #本程序监听地址,不包含http头,请放通port到公网,仅发图场景需要填写,可以是域名,暂不支持https.
  path : "http://123.123.123.123:11111"         #调用gensokyo api的地址,填入 gensokyo 的 正向http地址   http_address: "0.0.0.0:46231"  对应填入 "http://127.0.0.1:46231"
  paths : []                                    #当要连接多个onebotv11的http正向地址时,多个地址填入这里.
  conversationPath : "/conversation"            #所请求的conversation端口,在当前配置下不需要修改,在prompts文件夹中的promptstr=xxx xxx.yml中,可选用conversation_xxx n种不同的api(main.go 172行查看) 需打开allApi=true
  lotus : ""                                    #当填写另一个gensokyo-llm的http地址时,将请求另一个的conversation端点,实现多个llm不需要多次配置,简化配置,单独使用请忽略留空.例:http://192.168.0.1:12345(包含http头和端口)
  pathToken : ""                                #gensokyo正向http-api的access_token(是onebotv11标准的)
  apiType : 0                                   #0=混元 1=文心(文心平台包含了N种模型...) 2=gpt 3=rwkv 4=通义千问 5=智谱AI 6=腾讯元器
  stringob11 : false                            #兼容string模式ob11

  oneApi : false                                #内置了一个简化版的oneApi
  oneApiPort : 50052                            #内置简化版oneApi所监听的地址 :50052/v1
  modelInterceptor : false                      #用gsk=llm的model名字覆盖简化版oneApi的model配置.可以传入任意model名.会自动覆盖为gsk=llm的配置.

  iPWhiteList : ["192.168.0.101","127.0.0.1"]               #接口调用,安全ip白名单,gensokyo的ip地址,或调用api的程序的ip地址
  accessKey : ""                                #白名单ip未符合时,校验url参数&access_token=xxxx是否匹配

  systemPrompt : ["我是一个助手."]                           #人格提示词,或多个随机
  firstQ : [""]                                 #强化思想钢印,在每次对话的system之前固定一个QA,需都填写内容,会增加token消耗,可一定程度提高人格提示词效果,或抵抗催眠
  firstA : [""]                                 #强化思想钢印,在每次对话的system之前固定一个QA,需都填写内容,会增加token消耗,可一定程度提高人格提示词效果,或抵抗催眠
  secondQ : [""]                                #可空
  secondA : [""]                                #可空
  thirdQ : [""]                                 #可空
  thirdA : [""]                                 #可空

  groupMessage : true                         	#是否响应群信息
  splitByPuntuations : 40                       #截断率,仅在sse时有效,100则代表每句截断
  splitByPuntuationsGroup : 10                  #截断率(群),仅在sse时有效,100则代表每句截断
  sensitiveMode : false                         #是否开启敏感词替换
  sensitiveModeType : 0                         #0=只过滤用户输入 1=输出也进行过滤
  defaultChangeWord : "*"                       #默认的屏蔽词替换,你可以在sensitive_words.txt的####后修改为自己需要,可以用记事本批量替换

  ignoreExtraTips : false                       #自用,无视[[]]的消息不检查是否是注入[[]]内的内容只能来自自己数据库,向量数据库,不能是用户输入.可能有安全问题.被审核端开启.
  proxy : ""                                    #proxy设定,如http://127.0.0.1:7890 请仅在出海业务使用代理,如discord机器人
  saveResponses: [""]                           #安全拦截时的回复.
  restoreCommand : ["重置"]                     #重置指令关键词.
  restoreResponses : [""]                       #重置时的回复.

  usePrivateSSE : false                         #QQ开放平台的stream回复
  promptkeyboard : [""]                         #临时的promptkeyboard超过3个则随机,后期会增加一个ai生成的方式,也会是ai-agent

  savelogs : false                              #本地落地日志.
  noContext : false                             #不开启上下文     
  withdrawCommand : ["撤回"]                    #撤回指令
  memoryCommand : ["记忆"]                      #记忆指令
  memoryLoadCommand : ["载入"]                  #载入指令
  newConversationCommand : ["新对话"]           #新对话指令
  memoryListMD : 0                              #记忆列表使用md按钮(qq开放平台) 0=不用 1=按钮 2=inlinecmd(文字链)
  hideExtraLogs : false                         #忽略流信息的log,提高性能
  urlSendPics : false                           #自己构造图床加速图片发送.需配置公网ip+放通port+设置正确的selfPath

  groupHintWords : []                           #当机器人位于群内时,需满足包含groupHintWords数组任意内容如[CQ:at,qq=2] 机器人的名字 等
  groupHintChance : 0                           #需与groupHintWords联用,代表不满足hintwords时概率触发,不启用groupHintWords相当于百分百概率回复.
  groupContext : 0                              #群上下文 在智能体在群内时,以群为单位处理上下文. 0=默认 1=一个人一个上下文 2=群聊共享上下文
  groupAddNicknameToQ : 0                       #群上下文增加message.sender.nickname到上下文(昵称)让模型能知道发送者名字 0=默认 1=false 2=true
  groupAddCardToQ : 0                           #群上下文增加message.sender.card到上下文(群名片)让模型能知道发送者名字 0=默认 1=false 2=true
  noEmoji : 0                                   #0=默认,正常发emoji 1=正常发emoji 2=不发任何emoji
  superSafe : 0                                 #0=默认,1=正常,2=超级安全性

  specialNameToQ:                               #开启groupAddNicknameToQ和groupAddCardToQ时有效,应用特殊规则,让模型对某个id产生特殊称谓
  - id: 12345
    name: ""
  replacementPairsIn:                           #每个不同的yml文件,都可以有自己独立的替换词规则,IN OUT分离
  - originalWord: "hello"
    targetWord: "hi"
  replacementPairsOut:                          #每个不同的yml文件,都可以有自己独立的替换词规则,IN OUT分离
  - originalWord: "hello"
    targetWord: "hi"

  #Ws服务器配置
  wsServerToken : ""                            #ws密钥 可以由onebotv11反向ws接入
  wsPath : "nil"                                #设置了ws就不用设置path了,可以连接多个机器人.

  functionMode : false                          #是否指定本agent使用func模式(目前仅支持千帆平台),效果不好,暂时不用.
  functionPath : ""                             #调用另一个启用了func模式的gsk-llm联合工作的/conversation地址,效果不好,暂时不用.
  useFunctionPromptkeyboard : false             #使用func生成气泡,效果不好,暂时不用.

  AIPromptkeyboardPath : ""                     #调用另一个(可以是自身,规则,当使用中间件指定prompt参数时,配置位于prompts文件夹,其格式xxx-keyboard.yml,若未使用中间件,请在path中指定prompts参数,并将相应的xxx.yml放在prompts文件夹下)设置系统提示词的gsk-llm联合工作的/conversation地址,约定系统提示词需返回文本json数组(3个).
  useAIPromptkeyboard : false                   #使用ai生成气泡.
  mdPromptKeyboardAtGroup : false               #QQ智能体 群内mdPromptKeyboard
  mdPromptKeyboardAtGroupCMDs : []               #QQ智能体 固定指令
  groupNoKeyboard : false                        #群内不使用按钮

  #语言过滤
  allowedLanguages : ["cmn"]                    #根据自身安全实力,酌情过滤,cmn代表中文,小写字母,[]空数组代表不限制. /gensokyo api 可传参数skip_lang_check=true让某些信息跳过检查
  langResponseMessages : ["抱歉，我不会**这个语言呢","我不会**这门语言,请使用中文和我对话吧"]   #定型文,**会自动替换为检测到的语言
  questionMaxLenth : 100                        #最大问题字数. 0代表不限制
  qmlResponseMessages : ["问题太长了,缩短问题试试吧"]  #最大问题长度回复.
  blacklistResponseMessages : ["目前正在维护中...请稍候再试吧"]   #黑名单回复,将userid丢入blacklist.txt 一行一个

  #向量缓存(省钱-酌情调整参数)(进阶!!)需要有一定的调试能力,数据库调优能力,计算和数据测试能力.
  #不同种类的向量,维度和模型不同,所以请一开始决定好使用的向量,或者自行将数据库备份\对应,不同种类向量没有互相检索的能力。

  embeddingType : 0                             #0=混元向量 1=文心向量,需设置wenxinEmbeddingUrl 2=chatgpt向量,需设置gptEmbeddingUrl
  useCache : 1                              #使用缓存省钱.
  cacheThreshold : 100                          #阈值,以汉明距离单位. hunyuan建议250-300 文心v1建议80-100,越小越精确.
  cacheChance : 100                             #使用缓存的概率,前期10,积攒缓存,后期酌情增加,测试时100
  printHanming : true                           #输出汉明距离,还有分片基数(norm*CacheK)等完全确认下来汉明距离、分片数后，再关闭这个选项。
  cacheK : 10000000000                          #计算分片基数所用的值,请根据向量的实际情况和公式计算适合的值。默认值效果不错。
  cacheN : 256                                  #分片数量=256个 计算公式 (norm*CacheK) mod cacheN = 分组id 分组越多,分类越精确,数据库越快,cacheN不能大于(norm*CacheK)否则只分一组。
  printVector : false                           #直接输出向量的内容,根据经验判断和设置向量二值化阈值.
  vToBThreshold : 0                             #默认0效果不错,浮点数,向量二值化阈值,这里二值化是为了加速,损失了向量的精度,请根据输出的向量特征,选择具有中间特性的向量二值化阈值.
  vectorSensitiveFilter : false                 #是否开启向量拦截词,请放在同目录下的vector_sensitive.txt中 一行一个，可以是句子。 命令行参数 -test 会用test.exe中的内容跑测试脚本。
  vertorSensitiveThreshold : 200                #汉明距离,满足距离代表向量含义相近,可给出拦截.

  #多配置覆盖,切换条件等设置 该类配置比较绕,可咨询QQ2022717137
  promptMarksLength : 99999                        #未设置keywords时,多少轮开始切换上下文.
  enhancedQA : false                            #默认是false,用于在故事支线将firstQA的位置从顶部移动到用户之前,增强权重和效果.
  promptMarks:
  - branchName: "分支yml名"
    keywords: ["触发词", "触发词2"]

  promptChoicesQ:
  - round: 1
    replaceText: ["附加内容"]
    keywords: ["触发词"]

  promptChoicesA:
  - round: 1
    replaceText: ["附加内容"]
    keywords: ["触发词"]

  switchOnQ:
  - round: 1
    switch: ["分支", "分支"]
    keywords: ["触发词"]

  switchOnA:
  - round: 1
    switch: ["分支"]
    keywords: ["触发词"]

  exitOnQ:
    - round: 1
      keywords: ["触发词", "触发词", "触发词", "触发词"]

  exitOnA:
    - round: 1
      keywords: ["退出"]

  #混元配置项
  secretId : ""                                 #腾讯云账号(右上角)-访问管理-访问密钥，生成获取
  secretKey : ""
  region : ""                                   #留空
  maxTokensHunyuan : 4096                       #最大上下文
  hunyuanType : 0                               #0=高级版 1=标准版std 2=hunyuan-lite 3=hunyuan-standard 4=hunyuan-standard-256K 5=hunyuan-pro
  hunyuanStreamModeration : false               #是否采用流式审核
  topPHunyuan : 1.0                             #累积概率最高的令牌进行采样的界限
  temperatureHunyuan : 1.0                      #生成的随机性控制

  #文心配置项
  wenxinAccessToken : ""                        #请求百度access_token接口获取到的,有效期一个月,需要自己请求获取
  wenxinApiPath : "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant"    #在百度文档有，填啥就是啥模型，计费看文档
  wenxinEmbeddingUrl : "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/embeddings/embedding-v1"                       #百度的几种embedding接口url都可以用
  maxTokenWenxin : 4096
  wenxinTopp : 0.7                              #影响输出文本的多样性，取值越大，生成文本的多样性越强,默认0.7,范围0.1~1.0
  wenxinPenaltyScore : 1.0                      #通过对已生成的token增加惩罚,减少重复生成的现象。值越大表示惩罚越大,默认1.0
  wenxinMaxOutputTokens : 1024                  #指定模型最大输出token数,2~1024

  #chatgpt配置项 (这里我适配的是api2d.com的api)
  #chatgpt类接口仅适用于对接gensokyo-discord、gensokyo-telegram等平台,国内请符合相应的api要求.

  gptModel : "gpt-3.5-turbo"
  gptApiPath : ""
  gptEmbeddingUrl : ""                          #向量地址,和上面一样,基于标准的openai格式.哎哟..api2d这个向量好贵啊..暂不支持。
  gptToken : ""
  maxTokenGpt : 4096
  gptSafeMode : false                           #额外走腾讯云检查安全,但是会额外消耗P数(会给出回复,但可能跑偏)仅api2d支持
  gptModeration : false                         #额外走腾讯云检查安全,不合规直接拦截.(和上面一样但是会直接拦截.)仅api2d支持
  gptSseType : 0                                #gpt的sse流式有两种形式,0是只返回新的 你 好 呀 , 我 是 一 个,1是递增 你好呀，我是一个人类 你 你好 你好呀 你好呀， 你好呀，我 你好呀，我是
  standardGptApi : false                        #标准的gptApi,openai和groq需要开启.

  # RWKV 模型配置文件 仅适用于对接gensokyo-discord、gensokyo-telegram等平台,国内请遵守并符合相应的api资质要求.
  rwkvApiPath: "https://api.example.com/rwkv"       # 符合 RWKV 标准的 API 地址 是否以流形式取决于UseSSE配置
  rwkvMaxTokens: 1024                              # 最大的输出 Token 数量
  rwkvTemperature: 0.7                             # 生成的随机性控制
  rwkvTopP: 0.9                                    # 累积概率最高的令牌进行采样的界限
  rwkvPresencePenalty: 0.0                         # 当前上下文中令牌出现的频率惩罚
  rwkvFrequencyPenalty: 0.0                        # 全局令牌出现的频率惩罚
  rwkvPenaltyDecay: 0.99                           # 惩罚值的衰减率
  rwkvTopK: 25                                     # 从概率最高的K个令牌中采样
  rwkvSseType : 0                                  # 同gptSseType
  rwkvGlobalPenalty: false                         # 是否在全局上应用频率惩罚
  rwkvStop:                                        # 停止生成的标记列表
    - "\n\nUser"
  rwkvUserName: "User"                             # 用户名称
  rwkvAssistantName: "Assistant"                   # 助手名称
  rwkvSystemName: "System"                         # 系统名称
  rwkvPreSystem: false                             # 是否在系统层面进行预处理

  # TYQW 模型配置文件，适用于对接您的平台。请遵守并符合相应的API资质要求。
  tyqwApiPath: "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"       # 符合 TYQW 标准的 API 地址，是否以流形式取决于UseSSE配置
  tyqwMaxTokens: 1500                               # 最大的输出 Token 数量
  tyqwModel : ""                                    # 指定用于对话的通义千问模型名，目前可选择qwen-turbo、qwen-plus、qwen-max、qwen-max-0403、qwen-max-0107、qwen-max-1201和qwen-max-longcontext。
  tyqwApiKey : ""                                   # api的key
  tyqwWorkspace : ""                                # 指明本次调用需要使用的workspace；需要注意的是，对于子账号Apikey调用，此参数为必选项，子账号必须归属于某个workspace才能调用；对于主账号Apikey此项为可选项，添加则使用对应的workspace身份，不添加则使用主账号身份。
  tyqwTemperature: 0.85                             # 生成的随机性控制
  tyqwTopP: 0.9                                     # 累积概率最高的令牌进行采样的界限
  tyqwPresencePenalty: 0.2                          # 当前上下文中令牌出现的频率惩罚
  tyqwFrequencyPenalty: 0.2                         # 全局令牌出现的频率惩罚
  tyqwPenaltyDecay: 0.99                            # 惩罚值的衰减率
  tyqwRepetitionPenalty : 1.1                       # 用于控制模型生成时的重复度。提高repetition_penalty时可以降低模型生成的重复度。1.0表示不做惩罚，默认为1.1。没有严格的取值范围。
  tyqwTopK: 40                                      # 从概率最高的K个令牌中采样
  tyqwSeed : 1234                                   # 生成时使用的随机数种子，用户控制模型生成内容的随机性。seed支持无符号64位整数，默认值为1234。在使用seed时，模型将尽可能生成相同或相似的结果，但目前不保证每次生成的结果完全相同。
  tyqwSseType: 1                                    # 1=默认,sse发新内容 2=sse内容递增(不推荐)
  tyqwGlobalPenalty: false                          # 是否在全局上应用频率惩罚
  tyqwStop:                                         # 停止生成的标记列表
    - "\n\nUser"                                    
  tyqwUserName: "User"                              # 用户名称
  tyqwAssistantName: "Assistant"                    # 助手名称
  tyqwSystemName: "System"                          # 系统名称
  tyqwPreSystem: false                              # 是否在系统层面进行预处理
  tyqwEnableSearch : false                          # 是否使用网络搜索

  # GLM 模型配置文件，为确保与API接口兼容，请符合相应的API资质要求。
  glmApiPath: "https://open.bigmodel.cn/api/paas/v4/chat/completions"  # GLM API的地址，用于调用模型生成文本
  glmApiKey : ""                                   # glm的api密钥          
  glmModel: ""                                     # 指定用于调用的模型编码，根据您的需求选择合适的模型,可选 glm-3-turbo glm-4
  glmRequestID: ""                                 # 请求的唯一标识，用于追踪和调试请求
  glmDoSample: true                                # 是否启用采样策略，默认为true，采样开启
  glmTemperature: 0.95                             # 控制输出随机性的采样温度，值越大输出越随机
  glmTopP: 0.9                                     # 采用核取样策略，从概率最高的令牌中选择top P的比例
  glmMaxTokens: 1024                               # 模型输出的最大token数，控制输出长度
  glmStop:                                         # 模型输出时遇到以下标记将停止生成
    - "stop_token"                                 # 可以列出多个停止标记
  glmTools:                                        # 列出模型可以调用的工具列表
    - "web_search"                                 # 默认启用网络搜索工具
  glmToolChoice: "auto"                            # 工具选择策略，目前支持auto，自动选择最合适的工具
  glmUserID: ""                                    # 用户唯一标识，用于跟踪和分析用户行为

  # Yuanqi 助手配置文件，确保按业务需求配置。
  yuanqiApiPath: "https://open.hunyuan.tencent.com/openapi/v1/agent/chat/completions"
  yuanqiChatType: "published"   # 聊天类型，默认为published，支持preview模式下使用草稿态智能体
  yuanqiMaxToken: 4096
  yuanqiConfs:
  - yuanqiAssistantID: "123"
    yuanqiToken: "123"
    yuanqiName: "123"
  - yuanqiAssistantID: "123"
    uanqiToken: "123"
    yuanqiName: "123"

`

const Logo = `
'                                                                                                      
'    ,hakurei,                                                      ka                                  
'   ho"'     iki                                                    gu                                  
'  ra'                                                              ya                                  
'  is              ,kochiya,    ,sanae,    ,Remilia,   ,Scarlet,    fl   and  yu        ya   ,Flandre,   
'  an      Reimu  'Dai   sei  yas     aka  Rei    sen  Ten     shi  re  sca    yu      ku'  ta"     "ko  
'  Jun        ko  Kirisame""  ka       na    Izayoi,   sa       ig  Koishi       ko   mo'   ta       ga  
'   you.     rei  sui   riya  ko       hi  Ina    baI  'ran   you   ka  rlet      komei'    "ra,   ,sa"  
'     "Marisa"      Suwako    ji       na   "Sakuya"'   "Cirno"'    bu     sen     yu''        Satori  
'                                                                                ka'                   
'                                                                               ri'                    
`
