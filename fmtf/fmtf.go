package fmtf

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var logPath string

func init() {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exeDir := filepath.Dir(exePath)
	logPath = filepath.Join(exeDir, "log")
}

// 全局变量，用于存储日志启用状态
var enableFileLogGlobal bool

// SetEnableFileLog 设置 enableFileLogGlobal 的值
func SetEnableFileLog(value bool) {
	enableFileLogGlobal = value
}

// 独立的文件日志记录函数
func LogToFile(level, message string) {
	if !enableFileLogGlobal {
		return
	}
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err := os.Mkdir(logPath, 0755)
		if err != nil {
			panic(err)
		}
	}
	filename := time.Now().Format("2006-01-02") + ".log"
	filepath := logPath + "/" + filename

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	logEntry := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("2006-01-02T15:04:05"), level, message)
	if _, err := file.WriteString(logEntry); err != nil {
		Println("Error writing to log file:", err)
	}
}

func Print(v ...interface{}) {
	log.Println(v...)
	message := fmt.Sprint(v...)
	LogToFile("INFO", message)
}

func Println(v ...interface{}) {
	log.Println(v...)
	message := fmt.Sprint(v...)
	LogToFile("INFO", message)
}

func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
	message := fmt.Sprintf(format, v...)
	LogToFile("INFO", message)
}

// Fprintf 包装了fmt.Fprintf，并将格式化的消息同时记录到日志。
// w 是写入对象，应该实现了io.Writer接口（如http.ResponseWriter）。
// format 是格式字符串，v 是对应的参数列表。
func Fprintf(w io.Writer, format string, v ...interface{}) {
	// 将格式化的消息写入到w中
	fmt.Fprintf(w, format, v...)

	// 生成格式化的字符串消息
	message := fmt.Sprintf(format, v...)

	// 将生成的消息记录到日志
	LogToFile("INFO", message)
}

func Sprintf(format string, v ...interface{}) (str string) {
	log.Printf(format, v...)
	message := fmt.Sprintf(format, v...)
	LogToFile("INFO", message)
	return message
}

// Errorf 创建一个错误消息并记录该消息到日志文件。
// 它返回一个包含格式化错误信息的error，如果格式字符串中包含了 %w 指令，则还会包装一个错误。
func Errorf(format string, v ...interface{}) error {
	// 直接使用fmt.Errorf来构造错误，这样可以保留对原始错误的引用（如果使用了%w）
	err := fmt.Errorf(format, v...)

	// 将错误信息记录到日志
	LogToFile("ERROR", err.Error())

	// 返回构造的错误
	return err
}

func Fatalf(format string, v ...interface{}) {
	log.Printf(format, v...)
	message := fmt.Sprintf(format, v...)
	LogToFile("Fatal", message)
}
