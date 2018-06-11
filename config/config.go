package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Server []Server `toml:"server"`
	Php    Php
}

type Server struct {
	Port       string
	WebRoot    string `toml:"web_root"`
	Index      string
	SslEnabled bool    `toml:"ssl_enabled"`
	Key        string  `toml:"ssl_key"`
	Cert       string  `toml:"ssl_cert"`
	Vhosts     []Vhost `toml:"vhost"`
}

type Vhost struct {
	Name    string `toml:"server_name"`
	Port    string
	WebRoot string `toml:"web_root"`
	Index   string
}

type Php struct {
	Enabled bool
	FpmSock string `toml:"fpm_sock"`
	Index   string
}

func LoadConfig() Config {
	var conf Config
	_, err := toml.DecodeFile("config.tml", &conf)
	if err != nil {
		panic(err)
	}
	for i := range conf.Server {
		for j := range conf.Server[i].Vhosts {
			if conf.Server[i].Vhosts[j].Port == "" {
				conf.Server[i].Vhosts[j].Port = conf.Server[i].Port
			}
			if conf.Server[i].Vhosts[j].WebRoot == "" {
				conf.Server[i].Vhosts[j].WebRoot = conf.Server[i].WebRoot
			}
			if conf.Server[i].Vhosts[j].Index == "" {
				conf.Server[i].Vhosts[j].Index = conf.Server[i].Index
			}
		}
	}
	return conf
}
