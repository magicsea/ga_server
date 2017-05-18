package main

import (
	"GAServer/example/messages"
	. "GAServer/service"
	"fmt"

	"github.com/AsynkronIT/goconsole"
	"github.com/AsynkronIT/protoactor-go/actor"
)

type MyServer struct {
	BaseServer
}

func (s *MyServer) OnReceive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *messages.SayRequest:
		fmt.Println("OnREC:", msg.UserName, msg.Message)
	}
}
func (s *MyServer) OnInit() {

}
func (s *MyServer) OnStart() {

}

func main() {
	s := &MyServer{}
	s.Init("127.0.0.1:8090", "server")
	console.ReadLine()
}
