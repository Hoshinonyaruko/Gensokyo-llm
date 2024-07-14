package relay

import (
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor/ali"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor/baidu"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor/openai"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor/tencent"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor/zhipu"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/apitype"
)

func GetAdaptor(apiType int) adaptor.Adaptor {
	switch apiType {
	case apitype.Ali:
		return &ali.Adaptor{}
	case apitype.Baidu:
		return &baidu.Adaptor{}
	case apitype.OpenAI:
		return &openai.Adaptor{}
	case apitype.Tencent:
		return &tencent.Adaptor{}
	case apitype.Zhipu:
		return &zhipu.Adaptor{}
	case apitype.OpenAI2:
		return &openai.Adaptor{}
	case apitype.OpenAI3:
		return &openai.Adaptor{}
	}
	return nil
}
