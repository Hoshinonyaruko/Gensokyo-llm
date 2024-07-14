package openai

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/model"
)

type Adaptor struct {
	ChannelType int
}

func (a *Adaptor) Init() {
}

func (a *Adaptor) GetRequestURL() (string, error) {
	return GetFullRequestURL(), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request) error {
	adaptor.SetupCommonRequestHeader(c, req)

	req.Header.Set("Authorization", "Bearer "+config.GetGptToken())

	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if config.GetuseSse() == 2 {
		err, _, usage = StreamHandler(c, resp)

	} else {
		err, usage = Handler(c, resp)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	_, modelList := GetCompatibleChannelMeta()
	return modelList
}

func (a *Adaptor) GetChannelName() string {
	channelName, _ := GetCompatibleChannelMeta()
	return channelName
}
