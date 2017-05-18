package gate

import (
	"GAServer/log"

	"GAServer/cluster"
	"GAServer/messages"

	"errors"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

type AgentActor struct {
	key         string
	verified    bool
	bindAgent   *agent
	pid         *actor.PID
	parentPid   *actor.PID
	baseInfo    *messages.UserBaseInfo
	bindServers []*messages.UserBindServer
}

func NewAgentActor(ag *agent, parentPid *actor.PID) *AgentActor {
	//创建actor
	//r, err := parentPid.RequestFuture(&messages.NewChild{}, 3*time.Second).Result()
	ab := &AgentActor{verified: false, bindAgent: ag}
	pid := actor.Spawn(actor.FromInstance(ab))
	//pid, err := actor.SpawnWithParent(actor.FromInstance(ab), parentPid)
	//if err != nil {
	//	log.Println("SpawnNamed agent actor error:", err)
	//	return nil
	//}

	//pid := r.(messages.NewChildResult).Pid
	ab.pid = pid
	ab.parentPid = parentPid
	log.Println("new agent actor:", pid, "  parent:", parentPid)
	return ab
}

//收到后端消息
func (ab *AgentActor) Receive(context actor.Context) {
	//log.Println("agent.ReceviceServerMsg:", reflect.TypeOf(context.Message()))
	switch msg := context.Message().(type) {
	case *messages.FrameMsg:
		//log.Println("ReceviceServer:", msg)
		ab.SendClientRaw(msg.Channel, byte(msg.MsgId), msg.RawData)
	}
}

func (ab *AgentActor) GetChannelServer(channel messages.ChannelType) *actor.PID {
	c := messages.ChannelType(int(channel) / 100 * 100) //简单对应
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

//收到前端消息(unsafe)
func (ab *AgentActor) ReceviceClientMsg(data []byte) error {
	//log.Println("ReceviceClientMsg:", len(data))
	if len(data) < 2 {
		log.Error("AgentActor recv too short:", data)
		return errors.New("AgentActor recv too short")
	}
	//心跳包
	channel := messages.ChannelType(data[0])
	if channel == messages.Heartbeat {
		ab.SendClientRaw(channel, 0, data[2:])
		return nil
	}

	//认证
	if !ab.verified {
		return ab.CheckLogin(data)
	}

	//转发
	return ab.forward(data)
}

//验证消息
func (ab *AgentActor) CheckLogin(data []byte) error {
	channel := messages.ChannelType(data[0])
	msgid := data[1]
	if channel == messages.Login {
		msg := messages.CheckLogin{}
		proto.Unmarshal(data[2:], &msg)
		smsg := &messages.ServerCheckLogin{msg.Uid, msg.Key, ab.pid}
		//frame := &messages.FrameMsg{messages.ChannelType(channel), uint32(msgid), data[2:]}
		result, err := cluster.GetServicePID("session").Ask(smsg)
		if err == nil {
			checkResult := result.(*messages.CheckLoginResult)
			if checkResult.Result == messages.OK {
				//登录成功
				log.Println("CheckLogin success:", checkResult)
				ab.baseInfo = checkResult.BaseInfo
				ab.bindServers = checkResult.BindServers
				ab.verified = true
				ab.parentPid.Tell(&messages.AddAgentToParent{Uid: msg.Uid, Sender: ab.pid})
			} else {
				log.Println("###CheckLogin fail:", checkResult)
			}

			ab.SendClient(messages.Login, 0, checkResult)
		} else {
			log.Error("CheckLogin error :" + err.Error())

		}
	} else {
		log.Error("未验证无法发送非login消息%v,%v", channel, msgid)

	}
	return nil
}

//发送消息到客户端
func (ab *AgentActor) SendClient(c messages.ChannelType, msgId byte, msg proto.Message) {
	mdata, err := proto.Marshal(msg)
	if err != nil {
		log.Error("SendClient marshal error:%v", err)
		return
	}

	ab.SendClientRaw(c, msgId, mdata)
}
func (ab *AgentActor) SendClientRaw(c messages.ChannelType, msgId byte, mdata []byte) {
	data := []byte{byte(c), msgId}
	data = append(data, mdata...)
	ab.bindAgent.WriteMsg(data)
}

//转发
func (ab *AgentActor) forward(data []byte) error {
	channel := data[0]
	msgid := data[1]
	//test gate
	//if channel == byte(messages.Shop) {
	//	ab.SendClient(messages.Shop, byte(messages.S2C_ShopBuy), &messages.S2C_ShopBuyMsg{ItemId: 1, Result: messages.OK})
	//	return nil
	//}

	pid := ab.GetChannelServer(messages.ChannelType(channel))
	if pid == nil {
		log.Error("forward server nil:%+v,c=%v,m=%v", pid, channel, msgid)
		return nil
	}
	frame := &messages.FrameMsg{messages.ChannelType(channel), uint32(msgid), data[2:]}
	pid.Tell(frame)
	return nil
}

func (ab *AgentActor) Stop() {
	if ab.verified && ab.baseInfo != nil {
		ab.parentPid.Tell(&messages.RemoveAgentFromParent{Uid: ab.baseInfo.Uid})
	}

	ab.pid.Stop()
}
