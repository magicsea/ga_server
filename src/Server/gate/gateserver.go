package gate

import (
	"GAServer/config"
	gfw "GAServer/gateframework"
	"GAServer/log"
	"GAServer/service"
	"GAServer/util"
	"Server/cluster"
	"gameproto/msgs"

	"reflect"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type GateService struct {
	service.ServiceData
	agents    map[uint64]*actor.PID
	actorchan chan *AgentActor //传说创建actor
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
	log.Debug("GateService.OnReceive:", context.Message())
	switch msg := context.Message().(type) {

	case *msgs.AddAgentToParent:
		log.Info("msgs.AddAgentToParent%v", msg.Uid)
		//子对象注册
		s.agents[msg.Uid] = msg.Sender
	case *msgs.RemoveAgentFromParent:
		log.Info("msgs.RemoveAgentFromParent%v", msg.Uid)
		delete(s.agents, msg.Uid)
	case *msgs.NewChild:
		//创建子节点
		ab := &AgentActor{verified: false}
		pid := context.Spawn(actor.FromInstance(ab))
		ab.pid = pid
		ab.parentPid = context.Self()
		//context.Sender().Tell(&msgs.NewChildResult{Pid: pid})
		s.actorchan <- ab
	case *msgs.UnicastFrameMsg:
		//单todo:...
		log.Println("gate.UnicastFrameMsg:", msg)

	case *msgs.MulticastFrameMsg:
		//组todo:...
		log.Println("gate.UnicastFrameMsg:", msg)

	case *msgs.BroadcastFrameMsg:
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
	s.actorchan = make(chan *AgentActor)
}

func (s *GateService) OnStart(as *service.ActorService) {
	//as.RegisterMsg(reflect.TypeOf(&msgs.UserLogin{}), s.OnUserLogin) //注册登录
	as.RegisterMsg(reflect.TypeOf(&msgs.Tick{}), s.OnTick) //定时任务
	as.RegisterMsg(reflect.TypeOf(&msgs.Kick{}), s.OnKick) //踢人

	log.Println("gate start")
	gate := &gfw.Gate{
		MaxConnNum:      config.GetServiceConfigInt(s.Name, "MaxConnNum"),
		PendingWriteNum: 1024,
		MaxMsgLen:       65535,
		WSAddr:          config.GetServiceConfigString(s.Name, "WsAddr"),
		HTTPTimeout:     5,
		CertFile:        "",
		KeyFile:         "",
		TCPAddr:         config.GetServiceConfigString(s.Name, "TcpAddr"),
		LenMsgLen:       2,
		LittleEndian:    true,
		Processor:       nil, //msg.Processor,
		//AgentChanRPC:    nil, //game.ChanRPC,
	}

	gate.Run(s)

	//注册
	val := &msgs.ServiceValue{"TcpAddr", config.GetServiceConfigString(s.Name, "TcpAddr")}
	cluster.RegServerWork(&s.ServiceData, []*msgs.ServiceValue{val})
	//定时任务
	util.StartLoopTask(time.Second*5, func() {
		s.Pid.Tell(&msgs.Tick{}) //转主线程执行
	})
}

func (s *GateService) OnTick(context service.Context) {
	load := len(s.agents)
	cluster.UpdateServiceLoad(s.Name, uint32(load), msgs.ServiceStateFree)
}

func (s *GateService) OnKick(context service.Context) {
	msg := context.Message().(*msgs.Kick)
	log.Info("GateService.OnKick:%v", msg)
	if agent, ok := s.agents[msg.Uid]; ok {
		agent.Tell(&msgs.Kick{Uid: msg.Uid})
	}
}

func (s *GateService) OnDestory() {

}

//创建agentactor,外部线程调用.....not very safe...
func (s *GateService) GetAgentActor(a gfw.Agent) *actor.PID {
	s.Pid.Tell(new(msgs.NewChild)) //请求一个actor
	agentActor := <-s.actorchan
	agentActor.bindAgent = a
	return agentActor.pid
}
