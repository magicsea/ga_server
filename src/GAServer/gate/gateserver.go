package gate

import (
	"GAServer/cluster"
	"GAServer/config"
	"GAServer/log"
	"GAServer/messages"
	"GAServer/service"
	_ "encoding/json"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type GateService struct {
	service.ServiceData
	agents map[uint64]*actor.PID
}

//Service 获取服务对象
func Service() service.IService {
	return new(GateService)
}

func Type() string {
	return "gate"
}

//以下为接口函数
func (s *GateService) OnReceive(context service.Context) {
	log.Println("GateService.OnReceive:", context.Message())
	switch msg := context.Message().(type) {

	case *messages.AddAgentToParent:
		log.Info("messages.AddAgentToParent%v", msg.Uid)
		//子对象注册
		s.agents[msg.Uid] = msg.Sender
	case *messages.RemoveAgentFromParent:
		log.Info("messages.RemoveAgentFromParent%v", msg.Uid)
		delete(s.agents, msg.Uid)
	case *messages.NewChild:
		//创建子节点
		ab := &AgentActor{verified: false}
		pid := context.Spawn(actor.FromInstance(ab))
		context.Sender().Tell(&messages.NewChildResult{Pid: pid})

	case *messages.UnicastFrameMsg:
		//单todo:...
		log.Println("gate.UnicastFrameMsg:", msg)

	case *messages.MulticastFrameMsg:
		//组todo:...
		log.Println("gate.UnicastFrameMsg:", msg)

	case *messages.BroadcastFrameMsg:
		//广播
		log.Println("gate.BroadcastFrameMsg:", msg, " child:", len(s.agents))
		//children := context.Children()
		for _, child := range s.agents {
			log.Println("send agent:", child)
			child.Tell(msg.FrameMsg)
		}
	}
}
func (s *GateService) OnInit() {
	s.agents = make(map[uint64]*actor.PID)
}
func (s *GateService) OnStart(as *service.ActorService) {
	//as.RegisterMsg(reflect.TypeOf(&messages.UserLogin{}), s.OnUserLogin) //注册登录
	log.Println("gate start")
	gate := &Gate{
		MaxConnNum:      config.GetServiceConfigInt(s.Name, "MaxConnNum"),
		PendingWriteNum: 1024,
		MaxMsgLen:       65535,
		WSAddr:          config.GetServiceConfigString(s.Name, "WsAddr"),
		HTTPTimeout:     5,
		CertFile:        "",
		KeyFile:         "",
		TCPAddr:         config.GetServiceConfigString(s.Name, "TcpAddr"),
		LenMsgLen:       2,
		LittleEndian:    false,
		Processor:       nil, //msg.Processor,
		//AgentChanRPC:    nil, //game.ChanRPC,
	}

	gate.Run(s.GetPID())

	val := &messages.ServiceValue{"TcpAddr", config.GetServiceConfigString(s.Name, "TcpAddr")}
	cluster.RegServerWork(&s.ServiceData, []*messages.ServiceValue{val})
}

func (s *GateService) OnDestory() {

}
