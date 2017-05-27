package game

import (
	"GAServer/log"
	"gameproto/msgs"

	"github.com/gogo/protobuf/proto"
)

type PlayerChatModule struct {
	PlayerModuleBase
}

//=================接口实现======================
func (m *PlayerChatModule) OnInit() {
	m.RegistCmd(uint32(msgs.C2S_PrivateChat), m.PrivateChat)
	m.RegistCmd(uint32(msgs.C2S_WorldChat), m.WorldChat)
}

//===============feature functions====================
func (m *PlayerChatModule) ShopBuy(data []byte) {
	//var msg msgs.C2S_ShopBuyMsg
	//proto.Unmarshal(data, &msg)
	//fmt.Println("test buy something:", msg)
	//m.SendClientMsg(msgs.S2C_ShopBuy, &msgs.S2C_ShopBuyMsg{ItemId: msg.ItemId, Result: msgs.OK})
}

func (m *PlayerChatModule) PrivateChat(data []byte) {
	var msg msgs.C2S_PrivateChatMsg
	proto.Unmarshal(data, &msg)

	result := AskSession(&msgs.GetSessionInfoByName{msg.TargetName})
	if result != nil {
		log.Info("AskSession PrivateChat ok:", result)
		ssInfo := result.(*msgs.GetSessionInfoResult)
		if ssInfo.Result == msgs.OK && ssInfo.AgentPID != nil {
			//找到玩家agent地址
			SendPlayerClientMsg(ssInfo.AgentPID,
				msgs.Chat,
				byte(msgs.S2C_PrivateOtherChat),
				&msgs.S2C_PrivateOtherChatMsg{SendName: m.player.GetName(), Msg: msg.Msg})

			//通知自己
			m.SendClientMsg(msgs.S2C_PrivateChat,
				&msgs.S2C_PrivateChatMsg{TargetName: msg.TargetName, Msg: msg.Msg, Result: msgs.OK})
			log.Info("send PrivateChat:", msg)
		} else {
			//没找到玩家
			m.SendClientMsg(msgs.S2C_PrivateChat,
				&msgs.S2C_PrivateChatMsg{Result: msgs.NoFoundTarget})
			log.Info("send PrivateChat,no found:%v,%v", ssInfo.Result, msg)

		}
	}
}

func (m *PlayerChatModule) WorldChat(data []byte) {
	var msg msgs.C2S_WorldChatMsg
	proto.Unmarshal(data, &msg)

	SendWorldMsg(msgs.Chat, byte(msgs.S2C_WorldChat),
		&msgs.S2C_WorldChatMsg{SendName: m.player.GetName(), Msg: msg.Msg})
}

//发送shop消息到客户端
func (m *PlayerChatModule) SendClientMsg(msgId msgs.ChatMsgType, msg proto.Message) {
	m.player.SendClientMsg(msgs.Chat, byte(msgId), msg)
}
