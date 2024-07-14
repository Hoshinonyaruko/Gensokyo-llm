package baidu

import (
	"errors"
	"io"
	"net/http"

	"github.com/hoshinonyaruko/gensokyo-llm/config"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/model"
)

type Adaptor struct {
}

func (a *Adaptor) Init() {

}

func (a *Adaptor) GetRequestURL() (string, error) {

	fullRequestURL := config.GetWenxinApiPath()
	accessToken := config.GetWenxinAccessToken()

	fullRequestURL += "?access_token=" + accessToken
	return fullRequestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request) error {
	adaptor.SetupCommonRequestHeader(c, req)
	accessToken := config.GetWenxinAccessToken()
	req.Header.Set("Authorization", "Bearer "+accessToken)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	baiduRequest := ConvertRequest(*request)
	return baiduRequest, nil
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
		err, usage = StreamHandler(c, resp)
	} else {
		// switch meta.Mode {
		// case relaymode.Embeddings:
		// 	err, usage = EmbeddingHandler(c, resp)
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
	return "baidu"
}
