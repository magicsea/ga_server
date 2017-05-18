package gate

import (
	"GAServer/log"
	"GAServer/network"
	"net"
	"reflect"
)

type Agent interface {
	WriteMsg(msg interface{})
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
	Destroy()
	UserData() interface{}
	SetUserData(data interface{})
}

type agent struct {
	conn       network.Conn
	gate       *Gate
	agentActor *AgentActor
	userData   interface{}
}

func (a *agent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		//log.Info("agent.read msg:", len(data))
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		if a.gate.Processor != nil {
			msg, err := a.gate.Processor.Unmarshal(data)
			if err != nil {
				log.Debug("unmarshal message error: %v", err)
				break
			}
			err = a.gate.Processor.Route(msg, a)
			if err != nil {
				log.Debug("route message error: %v", err)
				break
			}

		} else {
			err := a.agentActor.ReceviceClientMsg(data)
			if err != nil {
				log.Error("ReceviceClientMsg message error: %v", err)
				break
			}
		}
	}
}

func (a *agent) OnClose() {
	if a.agentActor != nil {
		a.agentActor.Stop()
	}

}

func (a *agent) WriteMsg(data []byte) {
	err := a.conn.WriteMsg(data)
	if err != nil {
		log.Error("write message %v error: %v", reflect.TypeOf(data), err)
	}

}

func (a *agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *agent) Close() {
	a.conn.Close()
}

func (a *agent) Destroy() {
	a.conn.Destroy()

}

func (a *agent) UserData() interface{} {
	return a.userData
}

func (a *agent) SetUserData(data interface{}) {
	a.userData = data
}
