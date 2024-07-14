package ali

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/model"
)

// https://help.aliyun.com/zh/dashscope/developer-reference/api-details

type Adaptor struct {
}

func (a *Adaptor) Init() {
}

func (a *Adaptor) GetRequestURL() (string, error) {
	fullRequestURL := config.GetTyqwApiPath()
	return fullRequestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request) error {
	adaptor.SetupCommonRequestHeader(c, req)
	if config.GetuseSse() == 2 {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("X-DashScope-SSE", "enable")
	}
	req.Header.Set("Authorization", "Bearer "+config.GetTyqwKey())

	// if meta.Mode == relaymode.ImagesGenerations {
	// 	req.Header.Set("X-DashScope-Async", "enable")
	// }
	// if a.meta.Config.Plugin != "" {
	// 	req.Header.Set("X-DashScope-Plugin", a.meta.Config.Plugin)
	// }
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	aliRequest := ConvertRequest(*request)
	return aliRequest, nil

}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	aliRequest := ConvertImageRequest(*request)
	return aliRequest, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if config.GetuseSse() == 2 {
		err, usage = StreamHandler(c, resp)
	} else {
		// switch meta.Mode {
		// case relaymode.Embeddings:
		// 	err, usage = EmbeddingHandler(c, resp)
		// case relaymode.ImagesGenerations:
		// 	err, usage = ImageHandler(c, resp)
		// default:
		err, usage = Handler(c, resp)
		// }
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "ali"
}
