package template

const ConfigTemplate = `	
version: 1
settings:

  #通用配置项
  useSse : true
  port : 46233                                  #本程序监听端口,支持gensokyo http上报,   请在gensokyo的反向http配置加入 post_url: ["http://127.0.0.1:port/gensokyo"] 
  path : "http://123.123.123.123:11111"         #调用gensokyo api的地址,填入 gensokyo 的 正向http地址   http_address: "0.0.0.0:46231"  对应填入 "http://127.0.0.1:46231"
  apiType : 0                                   #0=混元 1=文心(文心平台包含了N种模型...) 2=gpt
  iPWhiteList : ["192.168.0.102"]               #接口调用,安全ip白名单,gensokyo的ip地址,或调用api的程序的ip地址
  systemPrompt : [""]                           #人格提示词,或多个随机
  groupMessage : true                       	#是否响应群信息
  splitByPuntuations : 40                       #截断率,仅在sse时有效,100则代表每句截断

  #混元配置项
  secretId : ""                                 #腾讯云账号(右上角)-访问管理-访问密钥，生成获取
  secretKey : ""
  region : ""                                   #留空
  maxTokensHunyuan : 4096                       #最大上下文
  hunyuanType : 0                               #0=高级版 1=标准版 价格差异10倍

  #文心配置项
  wenxinAccessToken : ""                        #请求百度access_token接口获取到的,有效期一个月,需要自己请求获取
  wenxinApiPath : "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant"    #在百度文档有，填啥就是啥模型，计费看文档
  maxTokenWenxin : 4096

  #chatgpt配置项 (这里我适配的是api2d.com的api)
  gptModel : "gpt-3.5-turbo"
  gptApiPath : ""
  ptToken : ""
  maxTokenGpt : 4096
  gptSafeMode : false                           #额外走腾讯云检查安全,但是会额外消耗P数
  gptSseType : 0                                #gpt的sse流式有两种形式,0是只返回新的 你 好 呀 , 我 是 一 个,1是递增 你好呀，我是一个人类 你 你好 你好呀 你好呀， 你好呀，我 你好呀，我是
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
