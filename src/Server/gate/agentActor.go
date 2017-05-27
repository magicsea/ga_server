package gate

import (
	gfw "GAServer/gateframework"
	"GAServer/log"
	"Server/cluster"
	"errors"
	"gameproto"
	"gameproto/msgs"

	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

type AgentActor struct {
	key         string
	verified    bool
	bindAgent   gfw.Agent
	pid         *actor.PID
	parentPid   *actor.PID
	baseInfo    *msgs.UserBaseInfo
	bindServers []*msgs.UserBindServer
	wantDead    bool
}

func NewAgentActor(ag gfw.Agent, parentPid *actor.PID) *AgentActor {
	//创建actor
	//r, err := parentPid.RequestFuture(&msgs.NewChild{}, 3*time.Second).Result()
	ab := &AgentActor{verified: false, bindAgent: ag}
	pid := actor.Spawn(actor.FromInstance(ab))
	//pid, err := actor.SpawnWithParent(actor.FromInstance(ab), parentPid)
	//if err != nil {
	//	log.Println("SpawnNamed agent actor error:", err)
	//	return nil
	//}

	//pid := r.(msgs.NewChildResult).Pid
	ab.pid = pid
	ab.parentPid = parentPid
	log.Println("new agent actor:", pid, "  parent:", parentPid)
	return ab
}

//外部调用tell
func (ab *AgentActor) Tell(msg proto.Message) {

}

//收到后端消息
func (ab *AgentActor) Receive(context actor.Context) {
	//log.Println("agent.ReceviceServerMsg:", reflect.TypeOf(context.Message()))
	switch msg := context.Message().(type) {
	case *msgs.Kick:
		ab.OnStop()
		//todo:not safe
		ab.bindAgent.SetDead() //被动死亡，防止二次关闭
		ab.bindAgent.Close()   //关闭连接
	case *msgs.ClientDisconnect:
		//上报
		if ab.baseInfo != nil {
			ss := cluster.GetServicePID("session")
			ss.Tell(&msgs.UserLeave{Uid: ab.baseInfo.Uid, From: msgs.ST_GateServer, Reason: "client disconnect"})
		}
		ab.OnStop()
		context.Self().Stop()
	case *msgs.ReceviceClientMsg:
		//收到客户端消息
		ab.ReceviceClientMsg(msg.Rawdata)
	case *msgs.FrameMsg:
		//log.Println("ReceviceServer:", msg)
		pack := new(NetPack)
		pack.channel = msg.Channel
		pack.msgID = byte(msg.MsgId)
		pack.rawData = msg.RawData
		ab.SendClientPack(pack)
	}
}

func (ab *AgentActor) GetChannelServer(channel msgs.ChannelType) *actor.PID {
	c := msgs.ChannelType(int(channel) / 100 * 100) //简单对应
	//log.Info("GetChannelServer,%v,%v", channel, c)
	if ab.bindServers == nil {
		return nil
	}
	for _, v := range ab.bindServers {
		//log.Info("try GetChannelServer,%v", v.Channel)
		if v.Channel == c {
			return v.GetPid()
		}
	}
	return nil
}

//收到前端消息
func (ab *AgentActor) ReceviceClientMsg(data []byte) error {
	//log.Println("ReceviceClientMsg:", len(data))
	pack := new(NetPack)
	if !pack.Read(data) {
		log.Error("AgentActor recv too short:", data)
		return errors.New("AgentActor recv too short")
	}
	//心跳包
	channel := msgs.ChannelType(pack.channel)
	if channel == msgs.Heartbeat {
		ab.SendClientPack(pack)
		return nil
	}

	//认证
	if !ab.verified {
		return ab.CheckLogin(pack)
	}

	//转发
	return ab.forward(pack)
}

//验证消息
func (ab *AgentActor) CheckLogin(pack *NetPack) error {
	log.Info("checklogin....")
	msg := gameproto.PlatformUser{}
	err := proto.Unmarshal(pack.rawData, &msg)
	if err != nil {
		log.Error("CheckLogin fail:%v,msgid:%d", err, pack.msgID)
		return err
	}
	pretime := time.Now()
	smsg := &msgs.ServerCheckLogin{Uid: uint64(msg.PlatformUid), Key: msg.Key, AgentPID: ab.pid}
	//frame := &msgs.FrameMsg{msgs.ChannelType(channel), uint32(msgid), data[2:]}
	result, err := cluster.GetServicePID("session").Ask(smsg)
	if err == nil {
		checkResult := result.(*msgs.CheckLoginResult)
		if checkResult.Result == msgs.OK {
			//登录成功
			usetime := time.Now().Sub(pretime)
			log.Info("CheckLogin success:%v,time:%v", checkResult, usetime.Seconds())
			ab.baseInfo = checkResult.BaseInfo
			ab.bindServers = checkResult.BindServers
			ab.verified = true
			ab.parentPid.Tell(&msgs.AddAgentToParent{Uid: checkResult.BaseInfo.Uid, Sender: ab.pid})
		} else {
			log.Println("###CheckLogin fail:", checkResult)
		}

		ret := &gameproto.LoginReturn{ErrCode: int32(checkResult.Result), ServerTime: int32(time.Now().Unix())}
		ab.SendClient(msgs.Login, byte(gameproto.S2C_LOGIN_END), ret)

	} else {
		log.Error("CheckLogin error :" + err.Error())
	}

	return nil
}

//发送消息到客户端
func (ab *AgentActor) SendClient(c msgs.ChannelType, msgId byte, msg proto.Message) {
	pack := new(NetPack)
	pack.channel = c
	pack.msgID = msgId
	mdata, err := proto.Marshal(msg)
	if err != nil {
		log.Error("SendClient marshal error:%v", err)
		return
	}
	pack.rawData = mdata
	//log.Info("sendclient:msg%v,data:%d=>%v", pack.msgID, len(pack.rawData), pack.rawData)
	ab.SendClientPack(pack)
}

//func (ab *AgentActor) SendClientRaw(c msgs.ChannelType, msgId byte, mdata []byte) {
//	data := []byte{byte(c), msgId}
//	data = append(data, mdata...)
//	ab.bindAgent.WriteMsg(data)
//}

func (ab *AgentActor) SendClientPack(pack *NetPack) {
	data := pack.Write()
	ab.bindAgent.WriteMsg(data)
}

//转发
func (ab *AgentActor) forward(pack *NetPack) error {
	channel := pack.channel
	msgid := pack.msgID
	//test gate
	//if channel == byte(msgs.Shop) {
	//	ab.SendClient(msgs.Shop, byte(msgs.S2C_ShopBuy), &msgs.S2C_ShopBuyMsg{ItemId: 1, Result: msgs.OK})
	//	return nil
	//}

	pid := ab.GetChannelServer(channel)
	if pid == nil {
		log.Error("forward server nil:%+v,c=%v,m=%v", pid, channel, msgid)
		return nil
	}

	frame := &msgs.FrameMsg{channel, uint32(msgid), pack.rawData}
	pid.Tell(frame)
	//r, e := pid.RequestFuture(frame, time.Second*3).Result()
	//if e != nil {
	//	log.Error("forward error:id=%v, err=%v", ab.baseInfo.Uid, e)
	//}

	//rep := r.(*msgs.FrameMsgRep)
	//repMsg := &gameproto.S2C_ConfirmInfo{MsgHead: int32(msgid), Code: int32(rep.ErrCode)}
	//ab.SendClient(msgs.GameServer, byte(gameproto.S2C_CONFIRM), repMsg)
	return nil
}

func (ab *AgentActor) OnStop() {
	if ab.verified && ab.baseInfo != nil {
		ab.parentPid.Tell(&msgs.RemoveAgentFromParent{Uid: ab.baseInfo.Uid})
	}
}
