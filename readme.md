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

- prompts文件夹需要有一个默认的keyboard.yml用于生成气泡.其系统提示词需要遵循json气泡生成器的prompts规则.

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

## 请求 `/gensokyo` 端点

当向 `/gensokyo` 端点发起请求时，系统支持附加 `prompt` 参数和 `api` 参数。`api` 参数允许指定如 `/conversation_ernie` 这类的完整端点。启用此功能需在配置中开启 `allapi` 选项。

示例请求：
```http
GET /gensokyo?prompt=example&api=conversation_ernie
```

## 请求 `/conversation` 端点

与 `/gensokyo` 类似，`/conversation` 端点支持附加 `prompt` 参数。

示例请求：
```http
GET /conversation?prompt=example
```

## `prompt` 参数解析

提供的 `prompt` 参数将引用可执行文件目录下的 `/prompts` 文件夹中相应的 YAML 文件（例如 `xxxx.yml`，其中 `xxxx` 是 `prompt` 参数的值）。

通过编写大量的prompts的yml文件，你可以实现角色卡切换，同一角色下，你可以实现故事情节和不同场景切换。

## YAML 配置文件

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
- [x] EnhancedQA（bool）

对于不在上述列表中的配置项，如果需要支持覆盖，请[提交 issue](#)。

所有的bool值在配置文件覆盖的yml中必须指定,否则将会被认为是false.

动态配置覆盖是一个我自己构思的特性,利用这个特性,可以实现配置文件之间的递归,举例,你可以在自己的中间件传递prompt=a,在a.yml中指定Lotus为调用自身,并在lotus地址中指定下一个prompt参数为b,b指定c,c指定d,以此类推.

---

## 故事模式(测试中 设计中)

本项目实践了一个基础的，实验性的，提示词和上下文构造方式，基于本项目实现的多配置文件，本项目设计了几个额外的参数，结合数据库储存每个用户的状态，

实现了一种简易的文本控制流,用于生成分支故事,交互式故事格式，参考了一些知名项目的设计思路，（参考部分）

实现了让用户在多个提示词中，按照一些条件，有顺序，可选择的，在多套提示词中进行流转，实现类似文字恋爱游戏的非连续性多支线的故事剧情。

有关的参数，思路来自**Inklewriter**的**ink**语言，一种用于描述非ai分支式故事的语言。将其与大模型ai进行了结合和简化。需要一定的学习，掌握后可以编写ai故事。

- [x]  promptMarkType : 0
- [x]  promptMarksLength : 1
- [x]  promptMarks : ["去逛街路上","在家准备"] #当promptMarkType==0 比较简单,达到promptMarksLength就会随机一个分支进行跳转
- [x]  promptMarks : ["去逛街路上:坐车-走路-触发","在家准备:等一下-慢慢-准备"] #当promptMarkType==1 :后 是关键词,Q和A包含任意关键词就会跳转到 去逛街路上.yml 这个分支
- [x]  enhancedQA : true
- [x]  promptChoicesQ: ["1:回家吧/我累了/不想去了/-我们打车去/快点去/想去/早点-我们走着去/不着急-等下-拉手"] #当用户本轮包含了我累了、不想去了，本轮用户Q会被叠加(回家吧)
- [x]  promptChoicesA: ["1:饿/我想吃饭/-难受/哄哄我"] #当AI本轮回复包含了，我想吃饭，本轮AI回复会被附加（饿），如果希望无条件附加，可以在末尾多加一个/，饿这个分支就是多加了一个/
- [x]  switchOnQ : ["1:故事退出分支/不想/累了-下一个分支/想/不累"]  #和promptMarks一样，可选,比promptMarks更具体，区分了Q和A，以及轮次1:
- [x]  switchOnA : ["1:晚上分支/时间不早了"]
- [x]  exitOnQ : ["1:退出/忘了吧/重置/无聊"] #捕获关键字来实现退出剧情,也可以使用全局的退出指令词。
- [x]  exitOnA : ["1:退出/我是一个AI/我是一个人工/我是一个基于"]
- [x]  enhancedPromptChoices: true  #promptChoicesQA  switchOnQA exitOnQA的语法，false时是随机模式 1:回家吧-不回家-原地休息 没有后方的/，随机一个分支跳转。true是具有关键词条件 1:回家吧/a/b/c-不回家/a/b/c
 
含义解释，以上参数均位于多配置文件的settings部分，你可以决定每个场景的提示词长度，每个场景的长度promptMarksLength,来控制剧情的颗粒度。

故事模式触发方式一，中间件控制,需要使用支持onebotv11标准的机器人框架和ob11插件应用端，以及本项目（3者联用），本项目面向的是有一定开发和试错能力的对话机器人开发者。

使用反向ws，使用ob11插件应用端与obv11机器人框架连接，当ob11插件应用端收到反向ws事件时，自行编写插件拦截ws的json，

通过json中的用户id，message内容，调用gsk-llm（本项目）的http /gensokyo端口，在这个环节，自行判断用户条件（如好感度），在/gensokyo端口附加不同的prompt参数，

故事模式触发方式二,通过配置默认配置文件config.yml的switchOnQ和switchOnA,可以根据关键词切换分支,

结合prompt参数中配置文件的自己推进故事走向的能力，可以实现基础的，以提示词为主的ai故事情节，此外还需要为每一个prompt.yml设计对应的-keyboard.yml，生成气泡。

之后，在gsk-llm设置ob11机器人框架的http api地址，ob11插件应用端不负责发信息，只是根据信息内容进行条件判断，作为控制中间件，给开发者自己控制条件的开发自由度。

promptMarkType=0代表按promptMarksLength来切换提示词文件，promptMarksLength代表本提示词文件维持的上下文长度，

当promptMarksLength小于0时，会从promptMarks中读取之后的分支，并从中随机一个切换，当promptMarkType=1时， 

1=按条件触发,promptMarksLength达到时也触发.条件格式aaaa:xxx-xxx-xxxx-xxx,aaa是promptmark中的yml,xxx是标记,

识别到用户和模型说出标记就会触发这个支线(需要自行写好提示词,让llm能根据条件说出.)

你可以使用当前故事片段的，系统提示词，QA，来引导AI输出与你约定的切换词，从而实现为每个目标分支设计多个触发词，让大模型自行决定故事的发展方向。

当enhancedQA为false时，会将配置文件中的预定义的QA加入到用户QA的顶部，存在于llm的记忆当中（不影响整体对话走向）形成弱影响

当enhancedQA为true时，我尝试将配置文件中预定义QA的位置从顶部下移到用户当前对话的前方，但效果不理想，

目前是会与当前用户的历史QA进行混合和融合，实现对用户输入进行一定程度的引导，从而左右故事进程的走向。

引入了“配置控制流”参数，这是一种相比ai-agent灵活性更低,但剧情可控性更高,生成速度和成本更低的方式。

promptChoicesQ: []                             #当enhancedQA为true时,若数组为空。将附加配置覆盖yml中最后一个Q到用户的当前输入中,格式Q:xxx(yml最后一个Q)。如果数组不为空,且格式需为"轮次编号:文本1-文本2-文本3"，例如"1:hello-goodbye-hi",会在符合的对话轮次中随机选择一个文本添加。所设置的promptChoices数量不能大于当前yml的promptMarksLength。

promptChoicesA: []                            #规则同上,对llm的A生效.我用于追加LLM的情绪和做一个补充的引导,比如llm的a回复包含了饿,可补充(想去吃饭,带我去吃饭...),会追加到当前A,对剧情起到推动和修饰.

enhancedPromptChoices: false                  #当设为true时,promptChoices的格式变化为"轮次编号:附加文本/触发词1/触发词2/触发词3-附加文本/触发词4/触发词5/触发词6"，如"1:hello/aaa/bbb/ccc-goodbye/ddd/eee/fff"。在指定轮次，根据触发词的匹配数量选择最适合的文本添加，匹配越多触发词的组合附加的文本越优先被选择。

switchOnQ代表在Q中寻找到匹配文本时切换当前分支,A同理,其语法和promptChoices一致.

exitOnQ需要enhancedPromptChoices=true,其实enhancedPromptChoices最好就是true的,其/左侧固定为退出(这里任意,右侧是触发词,退出没有具体作用)

promptMarks和switchOnQ、switchOnA在功能上是相同的，都是根据关键字跳转分支，promptMarks先执行，不分轮次不分QA，switchOnQ和switchOnA更具体，区分Q和A，区分轮次，实现细节跳转。

## 为什么采用文本控制流而不是ai-agent？

配置控制流简单直观，通过配置文件来管理对话逻辑，配置文件易于维护，非技术人员，如剧情编写者，可以直接学习配置文件规则，修改配置文件来更新对话逻辑，不需要编程知识。

剧情确定性高：给定相同的输入和配置，剧情走向是大体一致的，这对于确保对话剧情的连贯性和可预测性非常重要。

成本低，对上下文进行巧妙组合和替换，而不是多个ai同时处理，与普通对话消耗几乎等量的token，省钱。

速度快，像生成普通对话QA一样生成结果，像写游戏脚本一样编写剧情。

适用于个体开发者和小型开发团队的低成本ai故事、小说方案,低成本,高速度,高可控,效果随模型和提示词效果提升而直接提升。

对于对话剧情聊天场景，如果剧情较为固定，对话路径预设，且更新频率不高，使用配置控制流更适合，因为它提供了高度的可控性和易于理解的管理方式。

如果对话系统需要高度的互动性和个性化，或者剧情变化复杂，需要根据用户的具体反馈和行为动态调整，那么使用基于AI的agent方案可能更合适，它需要更高的技术投入和维护成本。

---

## 终结点

本节介绍了与API通信的具体终结点信息。

| 属性  | 详情                                    |
| ----- | --------------------------------------- |
| URL   | `http://localhost:46230/conversation`   |
| 方法  | `POST`                                  |

## 请求参数

客户端应向服务器发送的请求体必须为JSON格式，以下表格详细列出了每个字段的数据类型及其描述。

| 字段名            | 类型   | 描述                               |
| ----------------- | ------ | ---------------------------------- |
| `message`         | String | 用户发送的消息内容                 |
| `conversationId`  | String | 当前对话会话的唯一标识符            |
| `parentMessageId` | String | 与此消息关联的上一条消息的标识符    |

## 请求示例

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

## 返回值示例

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


## 思路参考

本项目参考了以下知名项目的思路,实现了一个简化的AI文本控制流配置格式.

1. **Rasa**
   - **项目主页**：[Rasa](https://rasa.com/)
   - **GitHub地址**：[Rasa on GitHub](https://github.com/RasaHQ/rasa)
   - Rasa是一个开源的机器学习框架，用于自动化文本和语音的对话。它让开发者能够构建复杂的聊天机器人，处理多轮对话，并支持自定义的对话管理策略。
   - 它使用了一种名为“故事”的格式来定义对话的可能路径，这些故事基于用户的意图和前置条件来控制对话的流程。

2. **Twine**
   - **项目主页**：[Twine](http://twinery.org/)
   - **GitHub地址**：[Twine on GitHub](https://github.com/klembot/twinejs)
   - Twine是用于创建交互式故事的开源工具，它非常适合编写分支故事和视觉小说。它提供了一个直观的视觉编辑界面，允许创作者无需编程知识即可创作故事。
   - 它允许作者编写基于选择的故事，其中故事的发展依赖于读者的决策。这是一种文本控制流的实现，用于叙述管理。

3. **Inklewriter**
    - **项目主页**，[Ink on GitHub](https://github.com/inkle/ink).
    - 它是一个允许用户创建交互式故事的网页平台。它设计了**ink**语言，它是一个开源项目，用于编写交互式叙述和游戏。