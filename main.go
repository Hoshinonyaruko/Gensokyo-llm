package main

import (
	"bufio"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3" // 只导入，作为驱动

	"github.com/hoshinonyaruko/gensokyo-llm/applogic"
	"github.com/hoshinonyaruko/gensokyo-llm/config"
	"github.com/hoshinonyaruko/gensokyo-llm/fmtf"
	"github.com/hoshinonyaruko/gensokyo-llm/hunyuan"
	"github.com/hoshinonyaruko/gensokyo-llm/template"
	"github.com/hoshinonyaruko/gensokyo-llm/utils"
)

func main() {
	testFlag := flag.Bool("test", false, "Run the test script,test.txt中的是虚拟信息,一行一条")
	flag.Parse()

	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {

		// 将修改后的配置写入 config.yml
		err = os.WriteFile("config.yml", []byte(template.ConfigTemplate), 0644)
		if err != nil {
			fmtf.Println("Error writing config.yml:", err)
			return
		}

		fmtf.Println("请配置config.yml然后再次运行.")
		fmtf.Print("按下 Enter 继续...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(0)
	}
	// 加载配置
	conf, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// 日志落地
	if config.GetSavelogs() {
		fmtf.SetEnableFileLog(true)
	}

	// 测试函数
	if *testFlag {
		// 如果启动参数包含 -test，则执行脚本
		err := utils.PostSensitiveMessages()
		if err != nil {
			log.Fatalf("Error running PostSensitiveMessages: %v", err)
		}
		return // 退出程序
	}
	// Deprecated
	secretId := conf.Settings.SecretId
	secretKey := conf.Settings.SecretKey
	fmtf.Printf("secretId:%v\n", secretId)
	fmtf.Printf("secretKey:%v\n", secretKey)
	region := config.Getregion()
	client, err := hunyuan.NewClientWithSecretId(secretId, secretKey, region)
	if err != nil {
		fmtf.Printf("创建hunyuanapi出错:%v", err)
	}

	db, err := sql.Open("sqlite3", "file:mydb.sqlite?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &applogic.App{DB: db, Client: client}
	// 在启动服务之前确保所有必要的表都已创建
	err = app.EnsureTablesExist()
	if err != nil {
		log.Fatalf("Failed to ensure database tables exist: %v", err)
	}
	// 确保user_context表存在
	err = app.EnsureUserContextTableExists()
	if err != nil {
		log.Fatalf("Failed to ensure user_context table exists: %v", err)
	}

	// 确保向量表存在
	err = app.EnsureEmbeddingsTablesExist()
	if err != nil {
		log.Fatalf("Failed to ensure EmbeddingsTable table exists: %v", err)
	}

	// 确保 QA缓存表 存在
	err = app.EnsureQATableExist()
	if err != nil {
		log.Fatalf("Failed to ensure EmbeddingsTable table exists: %v", err)
	}

	// 加载基于向量的拦截词 即使文本不同 也能按阈值精准拦截
	err = app.EnsureSensitiveWordsTableExists()
	if err != nil {
		log.Fatalf("Failed to ensure SensitiveWordsTable table exists: %v", err)
	}

	// 加载 拦截词
	err = app.ProcessSensitiveWords()
	if err != nil {
		log.Fatalf("Failed to ProcessSensitiveWords: %v", err)
	}

	apiType := config.GetApiType() // 调用配置包的函数获取API类型

	switch apiType {
	case 0:
		// 如果API类型是0，使用app.chatHandlerHunyuan
		http.HandleFunc("/conversation", app.ChatHandlerHunyuan)
	case 1:
		// 如果API类型是1，使用app.chatHandlerErnie
		http.HandleFunc("/conversation", app.ChatHandlerErnie)
	case 2:
		// 如果API类型是2，使用app.chatHandlerChatGpt
		http.HandleFunc("/conversation", app.ChatHandlerChatgpt)
	default:
		// 如果是其他值，可以选择一个默认的处理器或者记录一个错误
		log.Printf("Unknown API type: %d", apiType)
	}

	http.HandleFunc("/gensokyo", app.GensokyoHandler)
	port := config.GetPort()
	portStr := fmtf.Sprintf(":%d", port)
	fmtf.Printf("listening on %v\n", portStr)
	// 这里阻塞等待并处理请求
	log.Fatal(http.ListenAndServe(portStr, nil))
}
