package template

const ConfigTemplate = `
version: 1
settings:
	#混元配置项
	secretId : ""          #腾讯云账号(右上角)-访问管理-访问密钥，生成获取
	secretKey : ""
	region : ""            #留空
	#通用配置
	useSse : false         #流式发送
	port : 46230           #接口的端口   请在gensokyo将反向配置加入 post_url: ["http://127.0.0.1:port/gensokyo"] 
	path : ""              #填入 gensokyo 的 正向http地址   http_address: "0.0.0.0:46231"  对应填入 "http://127.0.0.1:46231"
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
