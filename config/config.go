package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Server struct {
		Port           string `yaml:"port"`
		Network        string `yaml:"network"`
		TickerInterval int    `yaml:"ticker-interval"`
	}
	Kafka struct {
		ServerList []string `yaml:"server-list"`
	}
}

var config *Config

var configPath string

func GetConfig() *Config {
	return config
}

func init() {
	flag.StringVar(&configPath, "conf", "./config.yml", "configuration path")
	flag.Parse()

	//读取配置文件
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
}
