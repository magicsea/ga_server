package main

import (
	"GAServer/network"
	//	"io/ioutil"
	"log"
	//	"net/http"

	//	"GAServer/network/protobuf"

	//	"github.com/gogo/protobuf/proto"
)

func newAgent(conn *network.TCPConn) network.Agent {
	Client := new(Agent)
	Client.conn = conn
	return Client
}

type Agent struct {
	conn      *network.TCPConn
	msgHandle func(channel byte, msgId byte, data []byte)
}

func (a *Agent) Run() {
	log.Println("Agent.run")
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Println("read message: ", err)
			break
		}

		a.msgHandle(data[0], data[1], data[2:])
	}
}

func (a *Agent) OnClose() {}

func (a *Agent) WriteMsg(channel byte, msgId byte, msg []byte) {

	data := []byte{channel, msgId}
	data = append(data, msg...)
	err := a.conn.WriteMsg(data)
	if err != nil {
		log.Println("write message error:", err)
	}

}

//func (a *Agent) LocalAddr() net.Addr {
//return a.conn.LocalAddr()
//}

//func (a *Agent) RemoteAddr() net.Addr {
//	return a.conn.RemoteAddr()
//}

func (a *Agent) Close() {
	a.conn.Close()
}

func (a *Agent) Destroy() {
	a.conn.Destroy()
}
