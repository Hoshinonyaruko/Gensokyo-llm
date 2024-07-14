package validator

import (
	"errors"
	"math"

	"github.com/hoshinonyaruko/gensokyo-llm/relay/model"
)

func ValidateTextRequest(textRequest *model.GeneralOpenAIRequest) error {
	if textRequest.MaxTokens < 0 || textRequest.MaxTokens > math.MaxInt32/2 {
		return errors.New("max_tokens is invalid")
	}
	if textRequest.Model == "" {
		return errors.New("model is required")
	}

	return nil
}
