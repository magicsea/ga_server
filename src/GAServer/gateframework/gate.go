package gateframework

import (
	"GAServer/network"
	_ "net"
	_ "reflect"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type IGateService interface {
	GetAgentActor(Agent) *actor.PID
}

type Gate struct {
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
	Processor       network.Processor
	//AgentChanRPC    *chanrpc.Server

	// websocket
	WSAddr      string
	HTTPTimeout time.Duration
	CertFile    string
	KeyFile     string

	// tcp
	TCPAddr      string
	LenMsgLen    int
	LittleEndian bool

	//实例
	wsServer  *network.WSServer
	tcpServer *network.TCPServer
}

func (gate *Gate) Run(gs IGateService) {

	var wsServer *network.WSServer
	if gate.WSAddr != "" {
		wsServer = new(network.WSServer)
		wsServer.Addr = gate.WSAddr
		wsServer.MaxConnNum = gate.MaxConnNum
		wsServer.PendingWriteNum = gate.PendingWriteNum
		wsServer.MaxMsgLen = gate.MaxMsgLen
		wsServer.HTTPTimeout = gate.HTTPTimeout
		wsServer.CertFile = gate.CertFile
		wsServer.KeyFile = gate.KeyFile
		wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
			a := &GFAgent{conn: conn, gate: gate}
			//if gate.AgentChanRPC != nil {
			//	gate.AgentChanRPC.Go("NewAgent", a)
			//}
			return a
		}
	}

	var tcpServer *network.TCPServer
	if gate.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.PendingWriteNum = gate.PendingWriteNum
		tcpServer.LenMsgLen = gate.LenMsgLen
		tcpServer.MaxMsgLen = gate.MaxMsgLen
		tcpServer.LittleEndian = gate.LittleEndian
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := &GFAgent{conn: conn, gate: gate}
			//ab := NewAgentActor(a, pid)
			//gs.Pid.Tell(new(messages.NewChild)) //请求一个actor
			//a.agentActor = <-gs.actorchan
			//a.agentActor.bindAgent = a
			a.agentActor = gs.GetAgentActor(a)
			return a
		}
	}

	if wsServer != nil {
		wsServer.Start()
	}
	if tcpServer != nil {
		tcpServer.Start()
	}

	gate.tcpServer = tcpServer
	gate.wsServer = wsServer
}

func (gate *Gate) OnDestroy() {
	if gate.wsServer != nil {
		gate.wsServer.Close()
	}
	if gate.tcpServer != nil {
		gate.tcpServer.Close()
	}
}
