package server

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
)

const (
	RequestInterval = time.Minute
)

type RateLimiter struct {
	Counts map[string][]time.Time
}

// 频率限制器
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		Counts: make(map[string][]time.Time),
	}
}

// 闭包,网页后端,图床逻辑,基于gin和www静态文件的简易图床
func UploadBase64ImageHandler(rateLimiter *RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ipAddress := r.RemoteAddr // Get client IP address
		if !rateLimiter.CheckAndUpdateRateLimit(ipAddress) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Ensure method is POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parsing form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		base64Image := r.FormValue("base64Image")
		fmt.Println("Received base64 data length:", len(base64Image), "characters")

		imageBytes, err := base64.StdEncoding.DecodeString(base64Image)
		if err != nil {
			fmt.Println("Error while decoding base64:", err)
			http.Error(w, "invalid base64 data", http.StatusBadRequest)
			return
		}

		// Assume getImageFormat and getFileExtensionFromImageFormat are implemented elsewhere
		imageFormat, err := getImageFormat(imageBytes)
		if err != nil {
			http.Error(w, "undefined picture format", http.StatusBadRequest)
			return
		}

		fileExt := getFileExtensionFromImageFormat(imageFormat)
		if fileExt == "" {
			http.Error(w, "unsupported image format", http.StatusBadRequest)
			return
		}

		fileName := getFileMd5(imageBytes) + "." + fileExt
		directoryPath := "./channel_temp/"
		savePath := directoryPath + fileName

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(directoryPath, 0755); err != nil {
			http.Error(w, "error creating directory", http.StatusInternalServerError)
			return
		}

		// If file exists, skip saving
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			if err := os.WriteFile(savePath, imageBytes, 0644); err != nil {
				http.Error(w, "error saving file", http.StatusInternalServerError)
				return
			}
		} else {
			fmt.Println("File already exists, skipping save.")
		}

		serverAddress := config.GetSelfPath()
		serverPort := config.GetPort()
		protocol := "http"
		if serverPort == 443 {
			protocol = "https"
		}

		imageURL := fmt.Sprintf("%s://%s:%d/channel_temp/%s", protocol, serverAddress, serverPort, fileName)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"url":"%s"}`, imageURL)
	}
}

// 闭包,网页后端,语音床逻辑,基于www静态文件的简易语音床
func UploadBase64RecordHandler(rateLimiter *RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ipAddress := r.RemoteAddr // Get client IP address
		if !rateLimiter.CheckAndUpdateRateLimit(ipAddress) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Ensure method is POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parsing form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		base64Record := r.FormValue("base64Record")
		fmt.Println("Received base64 data length:", len(base64Record), "characters")

		recordBytes, err := base64.StdEncoding.DecodeString(base64Record)
		if err != nil {
			fmt.Println("Error while decoding base64:", err)
			http.Error(w, "invalid base64 data", http.StatusBadRequest)
			return
		}

		fileName := getFileMd5(recordBytes) + ".silk" // Assuming .silk format for the audio file
		directoryPath := "./channel_temp/"
		savePath := directoryPath + fileName

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(directoryPath, 0755); err != nil {
			http.Error(w, "error creating directory", http.StatusInternalServerError)
			return
		}

		// If file exists, skip saving
		if _, err := os.Stat(savePath); !os.IsNotExist(err) {
			fmt.Println("File already exists, skipping save.")
		} else {
			if err := os.WriteFile(savePath, recordBytes, 0644); err != nil {
				http.Error(w, "error saving file", http.StatusInternalServerError)
				return
			}
		}

		serverAddress := config.GetSelfPath()
		serverPort := config.GetPort()
		protocol := "http"
		if serverPort == 443 {
			protocol = "https"
		}

		recordURL := fmt.Sprintf("%s://%s:%d/channel_temp/%s", protocol, serverAddress, serverPort, fileName)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"url":"%s"}`, recordURL)
	}
}

// 检查是否超过调用频率限制
func (rl *RateLimiter) CheckAndUpdateRateLimit(ipAddress string) bool {
	// 获取 MaxRequests 的当前值
	maxRequests := 600

	now := time.Now()
	rl.Counts[ipAddress] = append(rl.Counts[ipAddress], now)

	// Remove expired entries
	for len(rl.Counts[ipAddress]) > 0 && now.Sub(rl.Counts[ipAddress][0]) > RequestInterval {
		rl.Counts[ipAddress] = rl.Counts[ipAddress][1:]
	}

	return len(rl.Counts[ipAddress]) <= maxRequests
}

// 获取图片类型
func getImageFormat(data []byte) (format string, err error) {
	// Print the size of the data to check if it's being read correctly
	fmtf.Println("Received data size:", len(data), "bytes")

	_, format, err = image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		// Print additional error information
		fmtf.Println("Error while trying to decode image config:", err)
		return "", fmt.Errorf("error decoding image config: %w", err)
	}

	// Print the detected format
	fmtf.Println("Detected image format:", format)

	if format == "" {
		return "", errors.New("undefined picture format")
	}
	return format, nil
}

// 判断并返回图片类型
func getFileExtensionFromImageFormat(format string) string {
	switch format {
	case "jpeg":
		return "jpg"
	case "gif":
		return "gif"
	case "png":
		return "png"
	default:
		return ""
	}
}

// 生成随机md5图片名,防止碰撞
func getFileMd5(base64file []byte) string {
	md5Hash := md5.Sum(base64file)
	return hex.EncodeToString(md5Hash[:])
}

func OriginalUploadBehavior(base64Image string) (string, error) {
	// 原有的UploadBase64ImageToServer函数的实现
	protocol := "http"
	serverPort := config.GetPort()
	if serverPort == 443 {
		protocol = "https"
	}

	serverDir := config.GetSelfPath()
	if serverPort == 443 {
		protocol = "http"
		serverPort = 444
	}

	if isPublicAddress(serverDir) {
		url := fmt.Sprintf("%s://127.0.0.1:%d/uploadpic", protocol, serverPort)

		resp, err := postImageToServer(base64Image, url)
		if err != nil {
			return "", err
		}
		return resp, nil
	}
	return "", errors.New("local server uses a private address; image upload failed")
}

// 将base64语音通过lotus转换成url
func OriginalUploadBehaviorRecord(base64Image string) (string, error) {
	// 根据serverPort确定协议
	protocol := "http"
	serverPort := config.GetPort()
	if serverPort == 443 {
		protocol = "https"
	}

	serverDir := config.GetSelfPath()
	// 当端口是443时，使用HTTP和444端口
	if serverPort == 443 {
		protocol = "http"
		serverPort = 444
	}

	if isPublicAddress(serverDir) {
		url := fmt.Sprintf("%s://127.0.0.1:%d/uploadrecord", protocol, serverPort)

		resp, err := postRecordToServer(base64Image, url)
		if err != nil {
			return "", err
		}
		return resp, nil
	}
	return "", errors.New("local server uses a private address; image record failed")
}

// 请求语音床api(图床就是lolus为false的gensokyo)
func postRecordToServer(base64Image, targetURL string) (string, error) {
	data := url.Values{}
	data.Set("base64Record", base64Image) // 修改字段名以与服务器匹配

	resp, err := http.PostForm(targetURL, data)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response from server: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if value, ok := responseMap["url"]; ok {
		return fmt.Sprintf("%v", value), nil
	}

	return "", fmt.Errorf("URL not found in response")
}

// 请求图床api(图床就是lolus为false的gensokyo)
func postImageToServer(base64Image, targetURL string) (string, error) {
	data := url.Values{}
	data.Set("base64Image", base64Image) // 修改字段名以与服务器匹配

	resp, err := http.PostForm(targetURL, data)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response from server: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if value, ok := responseMap["url"]; ok {
		return fmt.Sprintf("%v", value), nil
	}

	return "", fmt.Errorf("URL not found in response")
}

// 判断是否公网ip 填写域名也会被认为是公网,但需要用户自己确保域名正确解析到gensokyo所在的ip地址
func isPublicAddress(addr string) bool {
	if strings.Contains(addr, "localhost") || strings.HasPrefix(addr, "127.") || strings.HasPrefix(addr, "192.168.") {
		return false
	}
	if net.ParseIP(addr) != nil {
		return true
	}
	// If it's not a recognized IP address format, consider it a domain name (public).
	return true
}
