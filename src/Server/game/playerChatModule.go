package game

import (
	"GAServer/log"
	"GAServer/messages"

	"github.com/gogo/protobuf/proto"
)

type PlayerChatModule struct {
	PlayerModuleBase
}

//=================接口实现======================
func (m *PlayerChatModule) OnInit() {
	m.RegistCmd(uint32(messages.C2S_PrivateChat), m.PrivateChat)
	m.RegistCmd(uint32(messages.C2S_WorldChat), m.WorldChat)
}

//===============feature functions====================
func (m *PlayerChatModule) ShopBuy(data []byte) {
	//var msg messages.C2S_ShopBuyMsg
	//proto.Unmarshal(data, &msg)
	//fmt.Println("test buy something:", msg)
	//m.SendClientMsg(messages.S2C_ShopBuy, &messages.S2C_ShopBuyMsg{ItemId: msg.ItemId, Result: messages.OK})
}

func (m *PlayerChatModule) PrivateChat(data []byte) {
	var msg messages.C2S_PrivateChatMsg
	proto.Unmarshal(data, &msg)

	result := AskSession(&messages.GetSessionInfoByName{msg.TargetName})
	if result != nil {
		log.Info("AskSession PrivateChat ok:", result)
		ssInfo := result.(*messages.GetSessionInfoResult)
		if ssInfo.Result == messages.OK && ssInfo.AgentPID != nil {
			//找到玩家agent地址
			SendPlayerClientMsg(ssInfo.AgentPID,
				messages.Chat,
				byte(messages.S2C_PrivateOtherChat),
				&messages.S2C_PrivateOtherChatMsg{SendName: m.player.GetName(), Msg: msg.Msg})

			//通知自己
			m.SendClientMsg(messages.S2C_PrivateChat,
				&messages.S2C_PrivateChatMsg{TargetName: msg.TargetName, Msg: msg.Msg, Result: messages.OK})
			log.Info("send PrivateChat:", msg)
		} else {
			//没找到玩家
			m.SendClientMsg(messages.S2C_PrivateChat,
				&messages.S2C_PrivateChatMsg{Result: messages.NoFoundTarget})
			log.Info("send PrivateChat,no found:%v,%v", ssInfo.Result, msg)
		}
	}
}

func (m *PlayerChatModule) WorldChat(data []byte) {
	var msg messages.C2S_WorldChatMsg
	proto.Unmarshal(data, &msg)

	SendWorldMsg(messages.Chat, byte(messages.S2C_WorldChat),
		&messages.S2C_WorldChatMsg{SendName: m.player.GetName(), Msg: msg.Msg})
}

//发送shop消息到客户端
func (m *PlayerChatModule) SendClientMsg(msgId messages.ChatMsgType, msg proto.Message) {
	m.player.SendClientMsg(messages.Chat, byte(msgId), msg)
}
