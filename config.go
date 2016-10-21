package main

import (
	"fmt"
	"os"

	"github.com/naoina/toml"
	"github.com/zhexuany/judge-proxy/client"
	"github.com/zhexuany/judge-proxy/httpd"
)

type Config struct {
	Downstreams client.Config `toml:"judge"`
	Httpd       httpd.Config  `toml:"httpd"`
}

func ParseConfig(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s:%v", path, err)
	}

	defer f.Close()

	cfg := &Config{}
	return cfg, toml.NewDecoder(f).Decode(cfg)
}
