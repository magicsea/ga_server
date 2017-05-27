package game

import (
	db2 "GAServer/db"
	"GAServer/log"
	"GAServer/service"
	"GAServer/util"
	"Server/cluster"
	"Server/db"
	"fmt"
	"gameproto"
	"gameproto/msgs"
	"strconv"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

type Player struct {
	UID         uint64
	baseInfo    *msgs.UserBaseInfo                 //基础信息
	selfPID     *actor.PID                         //本地地址
	agentPID    *actor.PID                         //gate的agent地址
	parentPID   *actor.PID                         //管理
	modules     []IPlayerModule                    //所有player模块
	rounter     map[msgs.ChannelType]IPlayerModule //消息路由到模块
	msgHandler  map[uint32]MessageReqFunc
	timer       *time.Ticker
	isDataDirty bool

	_transData *msgs.CreatePlayer
}

//MessageFunc 消息绑定函数
type MessageFunc func(data []byte)
type MessageReqFunc func(data []byte) msgs.GAErrorCode

func NewPlayer(uid uint64, agentpid *actor.PID, trans *msgs.CreatePlayer, context service.Context) (*Player, error) {
	p := &Player{UID: uid, agentPID: agentpid}
	p.msgHandler = make(map[uint32]MessageReqFunc)
	p.rounter = make(map[msgs.ChannelType]IPlayerModule)
	p._transData = trans
	p.parentPID = context.Self()
	props := actor.FromInstance(p)
	pid := context.Spawn(props)
	//pid, err := actor.SpawnWithParent(props, parent)
	//if err != nil {
	//	return nil, err
	//}
	p.selfPID = pid
	log.Info("NewPlayer:%v, agent=%v", p.GetID(), p.agentPID)
	return p, nil
}

func (p *Player) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here", msg)
		result := p.Start()
		p.parentPID.Tell(&PlayerInitEnd{Result: result, BaseInfo: p.baseInfo,
			Sender: context.Self(), TransData: p._transData})
		p._transData = nil
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down", msg)
		p.OnDestory()
		//case *actor.Stopped:
		//	fmt.Println("Stopped, actor and its children are stopped")
	case *msgs.Tick:
		p.OnTick()
	case *msgs.Kick:
		p.OnOutline(msg.Reason)
		context.Self().Stop() //kill!
	case *msgs.FrameMsg: //客户端消息
		err := msgs.UNKNOWN_ERROR
		mod := p.rounter[msg.Channel]
		if mod != nil {
			if !mod.route(msg) {
				log.Error("player recv unknow cmd:%v,%v,%v", p.GetID(), msg.Channel, msg.MsgId)
			}
		} else {
			if fun, ok := p.msgHandler[msg.MsgId]; ok {
				err = fun(msg.RawData)
			} else {
				log.Error("player recv unknow channel:id=%v,c=%v,code=%v", p.GetID(), msg.Channel, msg.MsgId)
			}
		}
		_ = err
		//context.Respond(&msgs.FrameMsgRep{ErrCode: err})
	}
}

//GetID 获取uid
func (p *Player) GetID() uint64 {
	return p.UID
}

//GetLevel 获取玩家等级
func (p *Player) GetLevel() uint64 {
	return p.baseInfo.Lv
}

//GetName 获取名字
func (p *Player) GetName() string {
	return p.baseInfo.Name
}

func (p *Player) String() string {
	return fmt.Sprintf("%v:%v", p.baseInfo.Name, p.baseInfo.Uid)
}

//RegistCmd 注册game处理消息
func (p *Player) RegistCmd(opcode gameproto.C2GS_CMD, fun MessageReqFunc) {
	p.msgHandler[uint32(opcode)] = fun
}

func (p *Player) Start() msgs.GAErrorCode {
	defer util.PrintPanicStack()

	p.InitModules()
	isFirst := p.InitTable()

	p.LoadData(isFirst)

	p.StartTimer()

	log.Info("player.start...")
	for _, mod := range p.modules {
		mod.OnStart()
	}
	return msgs.OK

}

func (p *Player) isNeedCreate() bool {
	//client := db.GetGameDB()
	//var temp db.Player
	//temp.Id = p.GetID()
	//norow, _ := client.Read(&temp)
	//return norow
	return true
}

func (p *Player) RollBack(err error, client *db2.DBClient) int32 {
	if err != nil {
		client.Rollback()
		return 1
	}
	return 0
}

// 尝试InitTable第一次插入数据
func (p *Player) InitTable() bool {
	defer util.PrintPanicStack()

	if p.isNeedCreate() {
		log.Info("create new player:%v", p.GetID())
	}
	return false
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
	p.AddModule(msgs.Shop, new(PlayerShopModule))
	//p.AddModule(msgs.Chat, new(PlayerChatModule))
}

func (p *Player) AddModule(ch msgs.ChannelType, mod IPlayerModule) {
	p.modules = append(p.modules, mod)
	p.rounter[ch] = mod
}

//StartTimer 计时器(new goroutine!)
func (p *Player) StartTimer() {
	p.timer = util.StartLoopTask(time.Minute, func() {
		p.selfPID.Tell(&msgs.Tick{}) //转主线程执行
	})
}

//OnTick 定时帧
func (p *Player) OnTick() {
	p.AutoSave()
	for _, mod := range p.modules {
		mod.OnTick()
	}
}

//LoadData 从db加载数据，长时阻塞操作
func (p *Player) LoadData(isFisrt bool) {
	p.baseInfo = &msgs.UserBaseInfo{Uid: p.UID, Name: "玩家" + strconv.Itoa(int(p.UID))}
	log.Println("player loaddata:", p.baseInfo)
	for _, mod := range p.modules {
		mod.OnLoad()
	}
}

//主动离开
func (p *Player) ActiveLeave() {
	//上报
	ss := cluster.GetServicePID("session")
	msg := &msgs.UserLeave{Uid: p.GetID(), From: msgs.ST_GameServer, Reason: "gameserver acive leave"}
	ss.Tell(msg)
	//保存
	p.OnOutline(msg.Reason)
	p.selfPID.Stop() //kill!
}

func (p *Player) OnOutline(reason string) {
	log.Info("player outline,%v,=> %v", p.String(), reason)

	//更新离开时间
	now := time.Now().Unix()
	user := &db.User{Id: p.GetID(), LastLogoutTime: now}
	db.GetGameDB().Update(user, "LastLogoutTime")

	p.AutoSave()
}

func (p *Player) OnDestory() {
	p.timer.Stop()
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

//game协议发送到客户端
func (p *Player) SendGameMsg(msgId gameproto.GS2C_CMD, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Error("SendGameMsg.Marshal error:%v", err)
	}
	frame := &msgs.FrameMsg{Channel: msgs.GameServer, MsgId: uint32(msgId), RawData: data}
	p.agentPID.Tell(frame)
}

//发送消息到自己客户端
func (p *Player) SendClientMsg(c msgs.ChannelType, msgId byte, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Error("SendClientMsg.Marshal error:%v", err)
	}
	frame := &msgs.FrameMsg{Channel: c, MsgId: uint32(msgId), RawData: data}
	p.agentPID.Tell(frame)
}

//发送到其他玩家
func SendPlayerClientMsg(gatePID *actor.PID, c msgs.ChannelType, msgId byte, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Error("SendPlayerClientMsg.Marshal error:%v", err)
	}
	frame := &msgs.FrameMsg{Channel: c, MsgId: uint32(msgId), RawData: data}
	gatePID.Tell(frame)
}

//发送到其他玩家
func SendWorldMsg(c msgs.ChannelType, msgId byte, msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Error("SendPlayerClientMsg.Marshal error:%v", err)
	}
	frame := &msgs.FrameMsg{Channel: c, MsgId: uint32(msgId), RawData: data}

	result := AskCenter(&msgs.GetTypeServices{ServiceType: "gate"})
	if result != nil {
		resultIns := result.(*msgs.GetTypeServicesResult)
		if resultIns.Pids != nil {
			for _, pid := range resultIns.Pids {
				pid.Tell(&msgs.BroadcastFrameMsg{FrameMsg: frame})
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
