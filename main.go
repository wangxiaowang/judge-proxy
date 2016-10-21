package main

import (
	"flag"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"

	"github.com/zhexuany/judge-proxy/client"
	"github.com/zhexuany/judge-proxy/httpd"
)

var (
	configPath string
	logFile    string
)

//Initialize flag option
func init() {
	flag.StringVar(&configPath, "config", "judge.toml", "judge-proxy config file path")
	flag.StringVar(&logFile, "logfile", "", "log file")
	flag.Parse()
}

//Ccre
func initLog() {
	if logFile != "" {
		rotateOutput := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     7,
		}
		log.SetOutput(rotateOutput)
	}

}
func main() {
	initLog()
	config, err := ParseConfig(configPath)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	c, err := client.NewClient(config.Downstreams)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	server, err := httpd.NewHttpServer(c, config.Httpd)
	if err != nil {
		log.Fatalf("failed to create http server: %v", err)
	}

	fmt.Println("start web server: ", config.Httpd)
	log.Fatal(server.Start())
}
