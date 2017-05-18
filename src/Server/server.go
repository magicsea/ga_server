package main

import (
	"GAServer/app"
	"GAServer/gate"
	"Server/center"
	"Server/game"
	"Server/login"
	"Server/session"
	"flag"
	"log"
	//"GAServer/login"
	//logr "github.com/Sirupsen/logrus"
	//logr "GAServer/log"
)

var (
	confPath = flag.String("config", "config.json", "配置文件")
)

func main() {
	flag.Parse()
	conf, err := LoadConfig(*confPath)
	if err != nil {
		log.Println("load config err:", err)
		return
	}
	app.RegisterService(center.Type(), center.Service)
	app.RegisterService(session.Type(), session.Service)
	app.RegisterService(login.Type(), login.Service)
	app.RegisterService(gate.Type(), gate.Service)
	app.RegisterService(game.Type(), game.Service)
	log.Println("===Run===", conf)
	app.Run(&conf.Base)
	log.Println("===GameOver===")

}
