package config

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Target 单个代理目标配置
type Target struct {
	URL       string  `mapstructure:"url"`
	Host      *string `mapstructure:"host"`
	SizeLimit int64   `mapstructure:"size_limit"`
}

// MySQLConfig MySQL 配置结构
type MySQLConfig struct {
	Enable bool   `mapstructure:"enable"`
	DSN    string `mapstructure:"dsn"`
}

// ConfigStruct 配置结构体
type ConfigStruct struct {
	ProxyUrl       string            `mapstructure:"proxy_url"`
	Host           string            `mapstructure:"host"`
	Port           int               `mapstructure:"port"`
	NoQuery        bool              `mapstructure:"no_query"`
	LogRequest     bool              `mapstructure:"log_request"`
	UploadTask     int64             `mapstructure:"upload_task"`
	MaxTimeout     time.Duration     `mapstructure:"max_timeout"`
	ValidTimeout   time.Duration     `mapstructure:"valid_timeout"`
	Invalid        int               `mapstructure:"invalid"`
	UploadEndpoint string            `mapstructure:"upload_endpoint"`
	MySQL          MySQLConfig       `mapstructure:"mysql"`
	Targets        map[string]Target `mapstructure:"targets"`
	Exts           []string          `mapstructure:"exts"`
}

//go:embed config.yaml
var embeddedConfig []byte

// Config 全局配置变量
var Config = initConfig()

// 初始化配置
func initConfig() *ConfigStruct {
	const configFile = "config.yaml"

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := os.WriteFile(configFile, embeddedConfig, 0644); err != nil {
			log.Fatalf("写入默认配置失败: %v", err)
		}
		fmt.Println("默认配置已生成:", configFile)
	}
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	var cfg ConfigStruct
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	return &cfg
}

func ShouldCacheByExt(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range Config.Exts {
		if ext == strings.ToLower(e) {
			return true
		}
	}
	return false
}

// MatchTarget 根据 URL 自动解析 host，匹配 target 返回对应代理目标
func MatchTarget(host string) Target {
	// 将 host 中的 '.' 替换为 '_'，匹配 YAML 配置中的 key
	key := strings.ReplaceAll(host, ".", "_")

	if target, ok := Config.Targets[key]; ok {
		return target
	}

	return Config.Targets["_default"]
}
