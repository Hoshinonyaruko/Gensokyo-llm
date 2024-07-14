package zhipu

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor/openai"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/model"
)

type Adaptor struct {
	APIVersion string
}

func (a *Adaptor) Init() {
}

func (a *Adaptor) SetVersionByModeName(modelName string) {
	if strings.HasPrefix(modelName, "glm-") {
		a.APIVersion = "v4"
	} else {
		a.APIVersion = "v3"
	}
}

func (a *Adaptor) GetRequestURL() (string, error) {
	return config.GetGlmApiPath(), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request) error {
	adaptor.SetupCommonRequestHeader(c, req)
	token := config.GetGlmApiKey()
	req.Header.Set("Authorization", token)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return ConvertRequest(*request), nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	newRequest := ImageRequest{
		Model:  request.Model,
		Prompt: request.Prompt,
		UserId: request.User,
	}
	return newRequest, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, requestBody)
}

func (a *Adaptor) DoResponseV4(c *gin.Context, resp *http.Response) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if config.GetuseSse() == 2 {
		err, _, usage = openai.StreamHandler(c, resp)
	} else {
		err, usage = openai.Handler(c, resp)
	}
	return
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if config.GetuseSse() == 2 {
		err, usage = StreamHandler(c, resp)
	} else {
		err, usage = Handler(c, resp)
	}
	return
}

func ConvertEmbeddingRequest(request model.GeneralOpenAIRequest) (*EmbeddingRequest, error) {
	inputs := request.ParseInput()
	if len(inputs) != 1 {
		return nil, errors.New("invalid input length, zhipu only support one input")
	}
	return &EmbeddingRequest{
		Model: request.Model,
		Input: inputs[0],
	}, nil
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "zhipu"
}
