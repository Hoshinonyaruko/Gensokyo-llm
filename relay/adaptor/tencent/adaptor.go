package tencent

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-llm/common/helper"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/adaptor"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/model"
)

// https://cloud.tencent.com/document/api/1729/101837

type Adaptor struct {
	Sign      string
	Action    string
	Version   string
	Timestamp int64
}

func (a *Adaptor) Init() {
	a.Action = "ChatCompletions"
	a.Version = "2023-09-01"
	a.Timestamp = helper.GetTimestamp()
}

func (a *Adaptor) GetRequestURL() (string, error) {
	region := config.Getregion()
	var url string

	if region == "0" || region == "" {
		url = "https://hunyuan.tencentcloudapi.com"
	} else {
		url = fmt.Sprintf("https://hunyuan.%s.tencentcloudapi.com", region)
	}

	return url, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request) error {
	adaptor.SetupCommonRequestHeader(c, req)
	req.Header.Set("Authorization", a.Sign)
	req.Header.Set("X-TC-Action", a.Action)
	req.Header.Set("X-TC-Version", a.Version)
	req.Header.Set("X-TC-Timestamp", strconv.FormatInt(a.Timestamp, 10))
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	secretId, secretKey, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	tencentRequest := ConvertRequest(*request)
	// we have to calculate the sign here
	a.Sign = GetSign(*tencentRequest, a, secretId, secretKey)
	return tencentRequest, nil
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
		err, _ = StreamHandler(c, resp)
	} else {
		err, usage = Handler(c, resp)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "tencent"
}
