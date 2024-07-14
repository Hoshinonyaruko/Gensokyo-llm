package adaptor

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-llm/relay/model"
)

type Adaptor interface {
	Init()
	GetRequestURL() (string, error)
	SetupRequestHeader(c *gin.Context, req *http.Request) error
	ConvertRequest(c *gin.Context, request *model.GeneralOpenAIRequest) (any, error)
	ConvertImageRequest(request *model.ImageRequest) (any, error)
	DoRequest(c *gin.Context, requestBody io.Reader) (*http.Response, error)
	DoResponse(c *gin.Context, resp *http.Response) (usage *model.Usage, err *model.ErrorWithStatusCode)
	GetModelList() []string
	GetChannelName() string
}
