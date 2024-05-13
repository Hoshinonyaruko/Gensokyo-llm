package applogic

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
)

func ParseMessageContent(message interface{}) string {
	messageText := ""

	switch message := message.(type) {
	case string:
		fmtf.Printf("params.message is a string\n")
		messageText = message
	case []interface{}:
		//多个映射组成的切片
		fmtf.Printf("params.message is a slice (segment_type_koishi)\n")
		for _, segment := range message {
			segmentMap, ok := segment.(map[string]interface{})
			if !ok {
				continue
			}

			segmentType, ok := segmentMap["type"].(string)
			if !ok {
				continue
			}

			segmentContent := ""
			switch segmentType {
			case "text":
				segmentContent, _ = segmentMap["data"].(map[string]interface{})["text"].(string)
			case "image":
				fileContent, _ := segmentMap["data"].(map[string]interface{})["file"].(string)
				segmentContent = "[CQ:image,file=" + fileContent + "]"
			case "voice":
				fileContent, _ := segmentMap["data"].(map[string]interface{})["file"].(string)
				segmentContent = "[CQ:record,file=" + fileContent + "]"
			case "record":
				fileContent, _ := segmentMap["data"].(map[string]interface{})["file"].(string)
				segmentContent = "[CQ:record,file=" + fileContent + "]"
			case "at":
				qqNumber, _ := segmentMap["data"].(map[string]interface{})["qq"].(string)
				segmentContent = "[CQ:at,qq=" + qqNumber + "]"
			case "markdown":
				mdContent, ok := segmentMap["data"].(map[string]interface{})["data"]
				if ok {
					if mdContentMap, isMap := mdContent.(map[string]interface{}); isMap {
						// mdContent是map[string]interface{}，按map处理
						mdContentBytes, err := json.Marshal(mdContentMap)
						if err != nil {
							fmtf.Printf("Error marshaling mdContentMap to JSON:%v", err)
						}
						encoded := base64.StdEncoding.EncodeToString(mdContentBytes)
						segmentContent = "[CQ:markdown,data=" + encoded + "]"
					} else if mdContentStr, isString := mdContent.(string); isString {
						// mdContent是string
						if strings.HasPrefix(mdContentStr, "base64://") {
							// 如果以base64://开头，直接使用
							segmentContent = "[CQ:markdown,data=" + mdContentStr + "]"
						} else {
							// 处理实体化后的JSON文本
							mdContentStr = strings.ReplaceAll(mdContentStr, "&amp;", "&")
							mdContentStr = strings.ReplaceAll(mdContentStr, "&#91;", "[")
							mdContentStr = strings.ReplaceAll(mdContentStr, "&#93;", "]")
							mdContentStr = strings.ReplaceAll(mdContentStr, "&#44;", ",")

							// 将处理过的字符串视为JSON对象，进行序列化和编码
							var jsonMap map[string]interface{}
							if err := json.Unmarshal([]byte(mdContentStr), &jsonMap); err != nil {
								fmtf.Printf("Error unmarshaling string to JSON:%v", err)
							}
							mdContentBytes, err := json.Marshal(jsonMap)
							if err != nil {
								fmtf.Printf("Error marshaling jsonMap to JSON:%v", err)
							}
							encoded := base64.StdEncoding.EncodeToString(mdContentBytes)
							segmentContent = "[CQ:markdown,data=" + encoded + "]"
						}
					}
				} else {
					fmtf.Printf("Error marshaling markdown segment to interface,contain type but data is nil.")
				}
			}

			messageText += segmentContent
		}
	case map[string]interface{}:
		//单个映射
		fmtf.Printf("params.message is a map (segment_type_trss)\n")
		messageType, _ := message["type"].(string)
		switch messageType {
		case "text":
			messageText, _ = message["data"].(map[string]interface{})["text"].(string)
		case "image":
			fileContent, _ := message["data"].(map[string]interface{})["file"].(string)
			messageText = "[CQ:image,file=" + fileContent + "]"
		case "voice":
			fileContent, _ := message["data"].(map[string]interface{})["file"].(string)
			messageText = "[CQ:record,file=" + fileContent + "]"
		case "record":
			fileContent, _ := message["data"].(map[string]interface{})["file"].(string)
			messageText = "[CQ:record,file=" + fileContent + "]"
		case "at":
			qqNumber, _ := message["data"].(map[string]interface{})["qq"].(string)
			messageText = "[CQ:at,qq=" + qqNumber + "]"
		case "markdown":
			mdContent, ok := message["data"].(map[string]interface{})["data"]
			if ok {
				if mdContentMap, isMap := mdContent.(map[string]interface{}); isMap {
					// mdContent是map[string]interface{}，按map处理
					mdContentBytes, err := json.Marshal(mdContentMap)
					if err != nil {
						fmtf.Printf("Error marshaling mdContentMap to JSON:%v", err)
					}
					encoded := base64.StdEncoding.EncodeToString(mdContentBytes)
					messageText = "[CQ:markdown,data=" + encoded + "]"
				} else if mdContentStr, isString := mdContent.(string); isString {
					// mdContent是string
					if strings.HasPrefix(mdContentStr, "base64://") {
						// 如果以base64://开头，直接使用
						messageText = "[CQ:markdown,data=" + mdContentStr + "]"
					} else {
						// 处理实体化后的JSON文本
						mdContentStr = strings.ReplaceAll(mdContentStr, "&amp;", "&")
						mdContentStr = strings.ReplaceAll(mdContentStr, "&#91;", "[")
						mdContentStr = strings.ReplaceAll(mdContentStr, "&#93;", "]")
						mdContentStr = strings.ReplaceAll(mdContentStr, "&#44;", ",")

						// 将处理过的字符串视为JSON对象，进行序列化和编码
						var jsonMap map[string]interface{}
						if err := json.Unmarshal([]byte(mdContentStr), &jsonMap); err != nil {
							fmtf.Printf("Error unmarshaling string to JSON:%v", err)
						}
						mdContentBytes, err := json.Marshal(jsonMap)
						if err != nil {
							fmtf.Printf("Error marshaling jsonMap to JSON:%v", err)
						}
						encoded := base64.StdEncoding.EncodeToString(mdContentBytes)
						messageText = "[CQ:markdown,data=" + encoded + "]"
					}
				}
			} else {
				fmtf.Printf("Error marshaling markdown segment to interface,contain type but data is nil.")
			}
		}
	default:
		fmtf.Println("Unsupported message format: params.message field is not a string, map or slice")
	}
	return messageText
}
