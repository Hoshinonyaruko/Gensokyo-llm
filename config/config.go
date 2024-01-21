package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	instance *Config
	mu       sync.Mutex
)

type Config struct {
	Version  int      `yaml:"version"`
	Settings Settings `yaml:"settings"`
}

type Settings struct {
	SecretId  string `yaml:"secretId"`
	SecretKey string `yaml:"secretKey"`
	Region    string `yaml:"region"`
	UseSse    bool   `yaml:"useSse"`
	Port      int    `yaml:"port"`
	HttpPath  string `yaml:"path"`
}

// LoadConfig 从文件中加载配置并初始化单例配置
func LoadConfig(path string) (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	// 如果单例已经被初始化了，直接返回
	if instance != nil {
		return instance, nil
	}

	configData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	err = yaml.Unmarshal(configData, conf)
	if err != nil {
		return nil, err
	}

	// 设置单例实例
	instance = conf
	return instance, nil
}

// 获取secretId
func GetsecretId() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.SecretId
	}
	return "0"
}

// 获取secretKey
func GetsecretKey() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.SecretKey
	}
	return "0"
}

// 获取region
func Getregion() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.Region
	}
	return "0"
}

// 获取useSse
func GetuseSse() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.UseSse
	}
	return false
}

// 获取GetPort
func GetPort() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.Port
	}
	return 46230
}

// 获取getHttpPath
func GetHttpPath() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.HttpPath
	}
	return "0"
}
