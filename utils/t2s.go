package utils

import (
	"log"
	"sync"

	"github.com/longbridgeapp/opencc"
)

// Global converter instance
var converter *opencc.OpenCC
var once sync.Once

// init function to initialize the global converter
func init() {
	var err error
	once.Do(func() {
		// Initialize the converter with the appropriate conversion configuration
		converter, err = opencc.New("t2s")
		if err != nil {
			log.Fatalf("Failed to initialize OpenCC converter: %v", err)
		}
	})
}

// ConvertTraditionalToSimplified converts traditional Chinese to simplified Chinese.
func ConvertTraditionalToSimplified(text string) (string, error) {
	return converter.Convert(text)
}
