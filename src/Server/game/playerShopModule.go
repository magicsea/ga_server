package game

import (
	"GAServer/messages"
	_ "fmt"

	"github.com/gogo/protobuf/proto"
)

type PlayerShopModule struct {
	PlayerModuleBase
}

//=================接口实现======================
func (m *PlayerShopModule) OnInit() {
	m.RegistCmd(uint32(messages.C2S_ShopBuy), m.ShopBuy)
	m.RegistCmd(uint32(messages.C2S_ShopSell), m.ShopSell)

}

func (m *PlayerShopModule) OnStart() {}
func (m *PlayerShopModule) OnLoad()  {}
func (m *PlayerShopModule) OnTick()  {}
func (m *PlayerShopModule) OnSave() {
	if !m.isDataDirty {
		return
	}
	//save action here...

}
func (m *PlayerShopModule) OnDestory() {}

//===============feature functions====================
func (m *PlayerShopModule) ShopBuy(data []byte) {
	var msg messages.C2S_ShopBuyMsg
	proto.Unmarshal(data, &msg)
	//fmt.Println("test buy something:", msg)
	m.SendClientMsg(messages.S2C_ShopBuy, &messages.S2C_ShopBuyMsg{ItemId: msg.ItemId, Result: messages.OK})
}

func (m *PlayerShopModule) ShopSell(data []byte) {

}

//发送shop消息到客户端
func (m *PlayerShopModule) SendClientMsg(msgId messages.ShopMsgType, msg proto.Message) {
	m.player.SendClientMsg(messages.Shop, byte(msgId), msg)
}
