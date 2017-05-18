package game

import (
	"GAServer/cluster"
	"GAServer/log"
	"GAServer/messages"
	"GAServer/service"
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

type Player struct {
	baseInfo *messages.UserBaseInfo                 //基础信息
	selfPID  *actor.PID                             //本地地址
	agentPID *actor.PID                             //gate的agent地址
	modules  []IPlayerModule                        //所有player模块
	rounter  map[messages.ChannelType]IPlayerModule //消息路由到模块

	isDataDirty bool
}

func NewPlayer(baseInfo *messages.UserBaseInfo, agentpid *actor.PID, context service.Context) (*Player, error) {
	p := &Player{baseInfo: baseInfo, agentPID: agentpid}
	p.rounter = make(map[messages.ChannelType]IPlayerModule)
	props := actor.FromInstance(p)
	pid := context.Spawn(props)
	//pid, err := actor.SpawnWithParent(props, parent)
	//if err != nil {
	//	return nil, err
	//}
	p.selfPID = pid
	log.Info("NewPlayer:%v, agent=%v", p.baseInfo.Uid, p.agentPID)
	return p, nil
}

func (p *Player) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here", msg)
		p.Start()
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down", msg)
		p.OnDestory()
		//case *actor.Stopped:
		//	fmt.Println("Stopped, actor and its children are stopped")
	case *messages.Tick:
		p.OnTick()
	case *messages.PlayerOutline:
		p.OnOutline(msg.Reason)
		context.Self().Stop() //kill!
	case *messages.FrameMsg: //客户端消息
		mod := p.rounter[msg.Channel]
		if mod != nil {
			if !mod.route(msg) {
				log.Error("player recv unknow cmd:%v,%v,%v", p.GetID(), msg.Channel, msg.MsgId)
			}
		} else {
			log.Error("player recv unknow channel:%v,%v", p.GetID(), msg.Channel)
		}
	}
}

func (p *Player) GetID() uint64 {
	return p.baseInfo.Uid
}

func (p *Player) GetName() string {
	return p.baseInfo.Name
}

func (p *Player) String() string {
	return fmt.Sprintf("%v:%v", p.baseInfo.Name, p.baseInfo.Uid)
}

func (p *Player) Start() {
	p.InitModules()
	p.LoadData()
	p.StartTimer()

	log.Info("player.start...")
	for _, mod := range p.modules {
		mod.OnStart()
	}
}

func (p *Player) InitModules() {
	//添加module
	p.CreateModules()

	//触发init
	for _, mod := range p.modules {
		mod.initData(p)
		mod.OnInit()
	}
}

func (p *Player) CreateModules() {
	//添加module here ...
	p.AddModule(messages.Shop, new(PlayerShopModule))
	p.AddModule(messages.Chat, new(PlayerChatModule))
}

func (p *Player) AddModule(ch messages.ChannelType, mod IPlayerModule) {
	p.modules = append(p.modules, mod)
	p.rounter[ch] = mod
}

//StartTimer 计时器(new goroutine!)
func (p *Player) StartTimer() {
	var timeTicker = time.NewTicker(time.Minute)

	go func() {
		for {
			select {
			case <-timeTicker.C:
				p.selfPID.Tell(&messages.Tick{}) //转player主线程执行
			}
		}
	}()
}

//OnTick 定时帧
func (p *Player) OnTick() {
	p.AutoSave()
	for _, mod := range p.modules {
		mod.OnTick()
	}
}

//LoadData 从db加载数据，长时阻塞操作
func (p *Player) LoadData() {
	log.Println("player loaddata:", p.baseInfo)
	for _, mod := range p.modules {
		mod.OnLoad()
	}
}

func (p *Player) OnOutline(reason string) {
	log.Info("player outline,%v,=> %v", p.String(), reason)
	p.AutoSave()
}

func (p *Player) OnDestory() {
	p.AutoSave()
	for _, mod := range p.modules {
		mod.OnDestory()
	}
}

func (p *Player) AutoSave() {
	if !p.isDataDirty {
		return
	}
	p.isDataDirty = false
	//save data
	for _, mod := range p.modules {
		mod.OnSave()
	}
}

//设置脏数据,等待异步写
func (p *Player) SetDataDirty() {
	p.isDataDirty = true
}

//发送消息到自己客户端
func (p *Player) SendClientMsg(c messages.ChannelType, msgId byte, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Error("SendClientMsg.Marshal error:%v", err)
	}
	frame := &messages.FrameMsg{Channel: c, MsgId: uint32(msgId), RawData: data}
	p.agentPID.Tell(frame)
}

//发送到其他玩家
func SendPlayerClientMsg(gatePID *actor.PID, c messages.ChannelType, msgId byte, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Error("SendPlayerClientMsg.Marshal error:%v", err)
	}
	frame := &messages.FrameMsg{Channel: c, MsgId: uint32(msgId), RawData: data}
	gatePID.Tell(frame)
}

//发送到其他玩家
func SendWorldMsg(c messages.ChannelType, msgId byte, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Error("SendPlayerClientMsg.Marshal error:%v", err)
	}
	frame := &messages.FrameMsg{Channel: c, MsgId: uint32(msgId), RawData: data}

	result := AskCenter(&messages.GetTypeServices{ServiceType: "gate"})
	if result != nil {
		resultIns := result.(*messages.GetTypeServicesResult)
		if resultIns.Pids != nil {
			for _, pid := range resultIns.Pids {
				pid.Tell(&messages.BroadcastFrameMsg{FrameMsg: frame})
			}
		}

	}
}

func AskSession(msg proto.Message) interface{} {
	ss := cluster.GetServicePID("session")
	result, err := ss.Ask(msg)
	if err != nil {
		log.Error("player.AskSession error:%v", err)
	}
	return result
}

func AskCenter(msg proto.Message) interface{} {
	ss := cluster.GetServicePID("center")
	result, err := ss.Ask(msg)
	if err != nil {
		log.Error("player.AskCenter error:%v", err)
	}
	return result
}
