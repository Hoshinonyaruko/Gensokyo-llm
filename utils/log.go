package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Tidylogs() {
	logDir := "./log"
	files, err := os.ReadDir(logDir)
	if err != nil {
		fmt.Println("Error reading log directory:", err)
		return
	}

	for _, file := range files {
		fileName := file.Name()
		if filepath.Ext(fileName) == ".log" && !strings.Contains(fileName, "-tidy") {
			processLogFile(filepath.Join(logDir, fileName))
		}
	}
}

func processLogFile(filePath string) {
	outputFilePath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + "-tidy.log"

	// Check if the tidy file already exists
	if _, err := os.Stat(outputFilePath); err == nil {
		fmt.Println("Skipping as tidy file already exists:", outputFilePath)
		return // File exists, skip processing
	} else if !os.IsNotExist(err) {
		fmt.Println("Error checking output file:", err)
		return // Some other error occurred when checking the file
	}

	// Read the entire file
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Define newline sequences and placeholder
	crlf := []byte{'\r', '\n'}
	lf := []byte{'\n'}
	placeholder := []byte{0xFF, 0xFE} // Safe placeholder for double newlines

	// Handle different newline formats
	doubleCRLF := append(crlf, crlf...)
	doubleLF := append(lf, lf...)

	// Replace double newlines with a placeholder
	data = bytes.ReplaceAll(data, doubleCRLF, placeholder)
	data = bytes.ReplaceAll(data, doubleLF, placeholder)

	// Remove standalone newlines
	data = bytes.ReplaceAll(data, crlf, []byte{})
	data = bytes.ReplaceAll(data, lf, []byte{})

	// Replace placeholders with a single newline (LF)
	data = bytes.ReplaceAll(data, placeholder, lf)

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// Scan through the modified content
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		// Process each line based on specific patterns
		if strings.Contains(line, "实际请求conversation端点内容:") {
			formatAndWriteQuestionLine(line, outputFile)
		}
		if strings.Contains(line, "A完整信息:") {
			formatAndWriteAnswerLine(line, outputFile)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning content:", err)
	}
}

func formatAndWriteQuestionLine(line string, outputFile *os.File) {
	prefix := "实际请求conversation端点内容:"
	startIndex := strings.Index(line, prefix)
	if startIndex != -1 {
		// 找到前缀后，提取从这个位置开始直到行尾的内容
		messageStart := startIndex + len(prefix)
		message := line[messageStart:]                  // 从"实际请求conversation端点内容:"后的内容开始提取到行尾
		message = strings.TrimSpace(message)            // 去除前后空格
		formattedLine := fmt.Sprintf("Q：%s\n", message) // 格式化行

		// 写入到输出文件
		_, err := outputFile.WriteString(formattedLine)
		if err != nil {
			fmt.Println("Error writing to output file:", err)
		}
	}
}

func formatAndWriteAnswerLine(line string, outputFile *os.File) {
	prefix := "A完整信息:"
	startIndex := strings.Index(line, prefix) // 查找"A完整信息:"的开始位置
	if startIndex != -1 {
		// 找到"A完整信息:"后，提取从这个位置开始直到行尾的内容
		messageStart := startIndex + len(prefix)
		message := line[messageStart:]                                     // 从"A完整信息:"后的内容开始提取到行尾
		formattedLine := fmt.Sprintf("A：%s\n", strings.TrimSpace(message)) // 格式化并去除前后空白字符

		// 写入到输出文件
		_, err := outputFile.WriteString(formattedLine)
		if err != nil {
			fmt.Println("Error writing to output file:", err)
		}
	}
}
